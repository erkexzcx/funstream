package proxy

import (
	"strings"
	"sync"
)

// LinkTypeUnknown is for channel's URLs that are not yet identified
// LinkTypeMedia is for channel's that return content-type of video/audio
// LinkTypeM3U8 is for channel's that are M3U8
// LinkTypeM3U8 is for channel's that return content-type of application/octet-stream
const (
	linkTypeUnknown     = 0
	linkTypeMedia       = 1
	linkTypeM3U8        = 2
	linkTypeStream      = 3
	linkTypeUnsupported = 4
)

// Playlist stores list of *Channel as well as some other settings, such as on-top-of-the-list channels
type Playlist struct {
	OrderedTitles []string
	Channels      map[string]*Channel
}

// Link stores stream URL and mutex.
type Link struct {
	Link     string
	LinkType int
	M3U8C    *M3U8Channel
	Mux      sync.RWMutex
}

// Channel stores TV channel details.
type Channel struct {
	Links           []Link
	ActiveLink      *Link
	ActiveLinkIndex int
	ActiveLinkMux   sync.RWMutex

	CycleCount    int
	CycleCountMux sync.Mutex

	Logo                 string
	LogoCache            []byte
	LogoCacheContentType string
	LogoCacheMux         sync.Mutex
	Group                string
}

var playlist *Playlist

// Same as 'cycleLinkNoMux()', but additionally locks ActiveLinkMux
func (c *Channel) cycleLink() bool {
	c.ActiveLinkMux.Lock()
	defer c.ActiveLinkMux.Unlock()
	return c.cycleLinkNoMux()
}

// Cycles links
func (c *Channel) cycleLinkNoMux() bool {
	c.CycleCountMux.Lock()
	defer c.CycleCountMux.Unlock()

	if c.ActiveLinkIndex == len(c.Links)-1 {
		c.ActiveLinkIndex = 0
	} else {
		c.ActiveLinkIndex++
	}
	c.ActiveLink = &c.Links[c.ActiveLinkIndex]

	c.CycleCount++
	if c.CycleCount >= len(c.Links)*2 {
		c.CycleCount = 0
		return false
	}
	return true
}

func getLinkType(contentType string) int {
	contentType = strings.ToLower(contentType)
	switch {
	case contentType == "application/vnd.apple.mpegurl" || contentType == "application/x-mpegurl":
		return linkTypeM3U8
	case strings.HasPrefix(contentType, "video/") || strings.HasPrefix(contentType, "audio/"):
		return linkTypeMedia
	case contentType == "application/octet-stream":
		return linkTypeStream
	default:
		return linkTypeUnsupported
	}
}
