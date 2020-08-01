package proxy

import (
	"strings"
	"sync"
)

const (
	linkTypeUnknown = 0
	linkTypeM3U8    = 1
	linkTypeMedia   = 2
)

// Playlist stores list of *Channel as well as some other settings, such as on-top-of-the-list channels
type Playlist struct {
	OrderedTitles []string
	Channels      map[string]*Channel
}

// Channel stores TV channel details.
type Channel struct {
	LinksMux        sync.Mutex
	Links           []Link
	ActiveLink      *Link
	ActiveLinkIndex int // Store index number of ActiveLink in Links list.
	CycleCount      int // Store count so we can track how many times we cycled through Links list.

	LogoCacheMux         sync.Mutex
	LogoCache            []byte
	LogoCacheContentType string

	// Synchronization is not required for below fields
	Logo  string
	Group string
}

// Link stores actual link to channel + reference to M3U8 channel if it's M3U8 type.
type Link struct {
	Link     string       // Actual link
	LinkType int          // Default is 0 (unknown)
	M3u8Ref  *M3U8Channel // Reference. For non M3U8 channels it will be empty
}

var playlist *Playlist

// Cycles links
func (c *Channel) cycleLink() bool {
	if c.ActiveLinkIndex == len(c.Links)-1 {
		c.ActiveLinkIndex = 0
	} else {
		c.ActiveLinkIndex++
	}
	c.ActiveLink = &c.Links[c.ActiveLinkIndex]

	c.CycleCount++
	if c.CycleCount == len(c.Links) {
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
	case strings.HasPrefix(contentType, "video/") || strings.HasPrefix(contentType, "audio/") || contentType == "application/octet-stream":
		return linkTypeMedia
	default:
		return linkTypeMedia
	}
}
