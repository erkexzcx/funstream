package proxy

import (
	"log"
	"net/http"
	"net/url"
	"strings"
)

func channelHandler(w http.ResponseWriter, r *http.Request) {
	reqPath := strings.Replace(r.URL.RequestURI(), "/iptv/", "", 1)
	reqPathParts := strings.SplitN(reqPath, "/", 2)
	if len(reqPathParts) == 0 {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("not found"))
		return
	}

	// Unescape channel title
	var err error
	reqPathParts[0], err = url.PathUnescape(reqPathParts[0])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid request"))
		return
	}

	// Debug
	log.Println("Received", reqPathParts)

	// Find channel
	c, ok := playlist.Channels[reqPathParts[0]]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("channel not found"))
		return
	}

	// Find link reference
	c.ActiveLinkMux.RLock()
	l := c.ActiveLink
	c.ActiveLinkMux.RUnlock()

	// Find channel type
	l.Mux.RLock()
	link := l.Link
	linkType := l.LinkType
	l.Mux.RUnlock()

	// Error if channel type is unknown and request URL contains additional path
	if linkType == linkTypeUnknown && len(reqPathParts) == 2 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid request"))
		log.Println("channel type is unknown and request URL contains additional path")
		return
	}

	// Lock mutex if channel's type is unknown, so no other routine tries to identify it at the same time
	if linkType == linkTypeUnknown {
		c.ActiveLink.Mux.Lock()
		defer c.ActiveLink.Mux.Unlock()
		linkType = c.ActiveLink.LinkType // In case previous routine updated it
	}

	// Understand what do we need to do with this link
	switch linkType {
	case linkTypeUnknown:
		handleLinkUnknown(w, r, &reqPathParts[0], link, c, c.ActiveLink)
	case linkTypeMedia:
		handleStream(w, r, link, c, c.ActiveLink)
	case linkTypeStream:
		handleStream(w, r, link, c, c.ActiveLink)
	case linkTypeM3U8:

		c.ActiveLink.Mux.RLock()
		m3u8c := c.ActiveLink.M3U8C
		c.ActiveLink.Mux.RUnlock()

		var newLink string
		if len(reqPathParts) == 1 {
			newLink = m3u8c.Link()
		} else {
			newLink = m3u8c.LinkRoot() + reqPathParts[1]
		}

		if len(reqPathParts) == 1 {
			// Channel only
			handleM3U8Channel(w, r, &reqPathParts[0], newLink, m3u8c, c.ActiveLink)
		} else {
			// Channel with data (additional path)
			handleM3U8ChannelData(w, r, &reqPathParts[0], newLink, m3u8c, c.ActiveLink)
		}
	case linkTypeUnsupported:
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("unsupported channel format"))
	default:
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
	}
}
