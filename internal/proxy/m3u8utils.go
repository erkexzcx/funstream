package proxy

import (
	"bufio"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"
)

const m3U8Timeout = 3 * time.Second

var m3u8channels map[string]*M3U8Channel

// M3U8Channel stores information about m3u8 channel
type M3U8Channel struct {
	Channel *Channel

	linkOneAtOnceMux sync.Mutex // To ensure that only one routine at once is accessing & updating link

	link    string
	linkMux sync.RWMutex

	linkRoot    string
	linkRootMux sync.RWMutex
}

// Link ...
func (c *M3U8Channel) Link() string {
	c.linkMux.RLock()
	defer c.linkMux.RUnlock()
	return c.link
}

// SetLink ...
func (c *M3U8Channel) SetLink(s string) {
	c.linkMux.Lock()
	defer c.linkMux.Unlock()
	c.link = s
}

// LinkRoot ...
func (c *M3U8Channel) LinkRoot() string {
	c.linkRootMux.RLock()
	defer c.linkRootMux.RUnlock()
	return c.linkRoot
}

// SetLinkRoot ...
func (c *M3U8Channel) SetLinkRoot(s string) {
	c.linkRootMux.Lock()
	defer c.linkRootMux.Unlock()
	c.linkRoot = s
}

// ----------

// Link ...
func (c *M3U8Channel) newRedirectedLink(s string) {
	c.SetLink(s)
	c.SetLinkRoot(deleteAfterLastSlash(s))
}

func deleteAfterLastSlash(str string) string {
	return str[0 : strings.LastIndex(str, "/")+1]
}

var reURILinkExtract = regexp.MustCompile(`URI="([^"]*)"`)

func rewriteLinks(scanner *bufio.Scanner, prefix, linkRoot string) string {
	var sb strings.Builder

	linkRootURL, _ := url.Parse(linkRoot) // It will act as a base URL for full URLs

	modifyLink := func(link string) string {
		var l string

		switch {
		case strings.HasPrefix(link, "//"):
			tmpURL, _ := url.Parse(link)
			tmp2URL, _ := url.Parse(tmpURL.RequestURI())
			link = (linkRootURL.ResolveReference(tmp2URL)).String()
			l = strings.ReplaceAll(link, linkRoot, "")
		case strings.HasPrefix(link, "/"):
			tmp2URL, _ := url.Parse(link)
			link = (linkRootURL.ResolveReference(tmp2URL)).String()
			l = strings.ReplaceAll(link, linkRoot, "")
		default:
			l = link
		}

		return prefix + l
	}

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "#") {
			line = modifyLink(line)
		} else if strings.Contains(line, "URI=\"") && !strings.Contains(line, "URI=\"\"") {
			link := reURILinkExtract.FindStringSubmatch(line)[1]
			line = reURILinkExtract.ReplaceAllString(line, `URI="`+modifyLink(link)+`"`)
		}
		sb.WriteString(line)
		sb.WriteByte('\n')
	}

	return sb.String()
}
