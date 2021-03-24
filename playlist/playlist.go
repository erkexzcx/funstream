package playlist

import (
	"errors"
	"io/ioutil"
	"log"
	"regexp"
	"sort"
	"strings"

	"github.com/erkexzcx/funstream/proxy"
	"gopkg.in/yaml.v2"
)

const (
	defaultLogo  = ""
	defaultGroup = "Unknown"
)

// config holds configuration from yaml file.
type config struct {
	DefaultGroup string `yaml:"default_group"`
	DefaultLogo  string `yaml:"default_logo"`
	Logos        []struct {
		Channel string `yaml:"channel"`
		Logo    string `yaml:"logo"`
	} `yaml:"logos"`
	DisplayFirst []string `yaml:"display_first"`
	Channels     []struct {
		Title string `yaml:"title"`
		URL   string `yaml:"url"`
		Logo  string `yaml:"logo"`
		Group string `yaml:"group"`
	} `yaml:"channels"`
	Playlists []struct {
		URL            string `yaml:"url"`
		RenameChannels []struct {
			From string `yaml:"from"`
			To   string `yaml:"to"`
		} `yaml:"rename_channels"`
		RenameGroups []struct {
			From string `yaml:"from"`
			To   string `yaml:"to"`
		} `yaml:"rename_groups"`
		ExcludeChannels []string `yaml:"exclude_channels"`
		ExcludeGroups   []string `yaml:"exclude_groups"`
	} `yaml:"playlists"`
}

// Playlist returns generated playlist out of configuration file.
func Playlist(flagPlaylist *string) (playlist *proxy.Playlist, err error) {
	log.Println("Reading config file...")
	c, err := readConfig(*flagPlaylist)
	if err != nil {
		return nil, err
	}

	log.Println("Ensuring default values are there...")
	c.updateDefaults()

	log.Println("Checking config for human errors...")
	if err = c.validate(); err != nil {
		return nil, err
	}

	log.Println("Normalizing cases in titles...")
	c.normalizeEverything()

	log.Println("Initiating empty internal playlist...")
	p := &proxy.Playlist{
		Channels: make(map[string]*proxy.Channel),
	}

	log.Println("Filling internal playlist...")
	c.fillChannels(p)

	// Create ordered titles list, using all available titles and titles provided in "DisplayFirst"
	log.Println("Sorting channels...")
	c.fillOrderedTitles(p)

	return p, nil
}

// Sorts channel titles and uses 'DisplayFirst' to put some channels to the top of the list
func (c *config) fillOrderedTitles(p *proxy.Playlist) {
	// DisplayFirst channels
	titles := make([]string, 0, len(c.DisplayFirst)+len(p.Channels))
	for _, title := range c.DisplayFirst {
		for chTitle := range p.Channels {
			if title == chTitle {
				titles = append(titles, title)
				break
			}
		}
	}

	// Everything else
	temporaryTitles := make([]string, 0, len(p.Channels))
	for chTitle := range p.Channels {
		temporaryTitles = append(temporaryTitles, chTitle)
	}
	sort.Strings(temporaryTitles)

	// Connect 2 slices into final one and add to playlist obj
	p.OrderedTitles = append(titles, temporaryTitles...)
}

// Downloads playlists and parses them, uses provided static channels and builds playlist.
func (c *config) fillChannels(p *proxy.Playlist) {
	addChannel := func(logo, group, title, link string) {
		// Check if such channel already exists
		channel, ok := p.Channels[title]
		if ok {
			// Channel already exists, so add link
			channel.Links = append(channel.Links, proxy.Link{Link: link})
		} else {
			// Channel does not exist, so create new with given link
			channel = &proxy.Channel{
				Links: []proxy.Link{
					{Link: link},
				},
			}
			channel.ActiveLink = &channel.Links[0]

			for _, l := range c.Logos {
				if l.Channel == title {
					channel.Logo = l.Logo
					break
				}
			}

			p.Channels[title] = channel
		}

		if channel.Logo == "" && logo != "" {
			channel.Logo = logo
		}
		if channel.Group == "" && group != "" {
			channel.Group = group
		}
	}

	log.Println("Adding defined channels...")

	for _, v := range c.Channels {
		addChannel(v.Logo, v.Group, v.Title, v.URL)
	}

	var regexChannels = regexp.MustCompile(`#EXTINF:[^\n]+,\s*[^\n]+\n+[a-z]+://[^\n]+`)

	for i, v := range c.Playlists {
		log.Println("Downloading and parsing playlist index", i)
		// Download playlist:
		contents, err := retrieveContents(v.URL)
		if err != nil {
			log.Println("Failed to download playlist: " + err.Error())
			continue
		}

		// Extract raw channels. Literally split string into substrings (raw channels)
		match := regexChannels.FindAllString(contents, -1)
		for _, rawChannel := range match {
			chLogo, chGroup, chTitle, chLink := parseRawChannel(rawChannel)

			if chTitle == "" || chLink == "" {
				log.Println("Received invalid channel from playlist!")
				continue
			}

			chGroup = normalize(chGroup)
			chTitle = normalize(chTitle)

			// First rename
			for _, renameGroup := range v.RenameGroups {
				if renameGroup.From == chGroup {
					chGroup = renameGroup.To
					break
				}
			}
			for _, renameChannel := range v.RenameChannels {
				if renameChannel.From == chTitle {
					chTitle = renameChannel.To
					break
				}
			}

			// Then exclude
			for _, excludeGroup := range v.ExcludeGroups {
				if excludeGroup == chGroup {
					goto nextChannel
				}
			}
			for _, excludeChannel := range v.ExcludeChannels {
				if excludeChannel == chTitle {
					goto nextChannel
				}
			}

			addChannel(chLogo, chGroup, chTitle, chLink)

		nextChannel:
		}

	}
}

var (
	regexChannelLogo        = regexp.MustCompile(`tvg-logo="([^"]*)"`)
	regexChannelGroup       = regexp.MustCompile(`group-title="([^"]*)"`)
	regexChannelTitleAndURL = regexp.MustCompile(`#EXTINF:[^\n]+,\s*([^\n]+)\n+([^\n]+)`)
)

// Extracts details from raw channel string
func parseRawChannel(rawChannel string) (logo, group, title, link string) {
	// Extract logo
	matchLogo := regexChannelLogo.FindAllStringSubmatch(rawChannel, -1)
	if len(matchLogo) == 1 && len(matchLogo[0]) == 2 {
		logo = matchLogo[0][1]
	}

	// Extract group
	matchGroup := regexChannelGroup.FindAllStringSubmatch(rawChannel, -1)
	if len(matchGroup) == 1 && len(matchGroup[0]) == 2 {
		group = matchGroup[0][1]
	}

	// Extract title and URL (link)
	matchTitleAndLink := regexChannelTitleAndURL.FindAllStringSubmatch(rawChannel, -1)
	if len(matchTitleAndLink) == 1 && len(matchTitleAndLink[0]) == 3 {
		title = matchTitleAndLink[0][1]
		link = matchTitleAndLink[0][2]
	}

	logo = strings.TrimSpace(logo)
	group = strings.TrimSpace(group)
	title = strings.TrimSpace(title)
	link = strings.TrimSpace(link)

	return
}

// Normalizes all titles for further comparisons.
func (c *config) normalizeEverything() {
	c.DefaultGroup = normalize(c.DefaultGroup)
	for i := 0; i < len(c.Logos); i++ {
		c.Logos[i].Channel = normalize(c.Logos[i].Channel)
	}
	for i := 0; i < len(c.DisplayFirst); i++ {
		c.DisplayFirst[i] = normalize(c.DisplayFirst[i])
	}
	for i := 0; i < len(c.Channels); i++ {
		c.Channels[i].Title = normalize(c.Channels[i].Title)
		c.Channels[i].Group = normalize(c.Channels[i].Group)
	}
	for i := 0; i < len(c.Playlists); i++ {
		for ii := 0; ii < len(c.Playlists[i].RenameChannels); ii++ {
			c.Playlists[i].RenameChannels[ii].From = normalize(c.Playlists[i].RenameChannels[ii].From)
			c.Playlists[i].RenameChannels[ii].To = normalize(c.Playlists[i].RenameChannels[ii].To)
		}
		for ii := 0; ii < len(c.Playlists[i].RenameGroups); ii++ {
			c.Playlists[i].RenameGroups[ii].From = normalize(c.Playlists[i].RenameGroups[ii].From)
			c.Playlists[i].RenameGroups[ii].To = normalize(c.Playlists[i].RenameGroups[ii].To)
		}
		for ii := 0; ii < len(c.Playlists[i].ExcludeChannels); ii++ {
			c.Playlists[i].ExcludeChannels[ii] = normalize(c.Playlists[i].ExcludeChannels[ii])
		}
		for ii := 0; ii < len(c.Playlists[i].ExcludeGroups); ii++ {
			c.Playlists[i].ExcludeGroups[ii] = normalize(c.Playlists[i].ExcludeGroups[ii])
		}
	}
}

// Inserts some default configs if user did not provide them
func (c *config) updateDefaults() {
	if c.DefaultLogo == "" {
		c.DefaultLogo = defaultLogo
	}
	if c.DefaultGroup == "" {
		c.DefaultGroup = defaultGroup
	}
}

// Validates config for human errors.
func (c *config) validate() error {
	// Check channels for missing mandatory fields. e.g. only title or only URL provided
	for _, v := range c.Channels {
		if v.Title == "" {
			return errors.New("channel's title is empty or not set")
		}
		if v.URL == "" {
			return errors.New("channel's URL is empty or not set")
		}
	}

	// Check rename_* in playlists for missing either 'to' or 'from' values
	for _, v := range c.Playlists {
		for _, vv := range v.RenameChannels {
			if vv.From == "" {
				return errors.New("playlist channels cannot be renamed because 'from' is empty or not set")
			}
			if vv.To == "" {
				return errors.New("playlist channels cannot be renamed because 'to' is empty or not set")
			}
		}
		for _, vv := range v.RenameGroups {
			if vv.From == "" {
				return errors.New("playlist groups cannot be renamed because 'from' is empty or not set")
			}
			if vv.To == "" {
				return errors.New("playlist groups cannot be renamed because 'to' is empty or not set")
			}
		}
	}

	return nil
}

func readConfig(path string) (*config, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var c config

	err = yaml.Unmarshal(content, &c)
	if err != nil {
		return nil, err
	}

	return &c, nil
}
