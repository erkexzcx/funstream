package proxy

import (
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/valyala/fasthttp"
)

func quickWrite(ctx *fasthttp.RequestCtx, content []byte, contentType string, httpStatus int) {
	ctx.SetContentType(contentType)
	ctx.SetStatusCode(httpStatus)
	ctx.SetBody(content)
}

func channelHandler(ctx *fasthttp.RequestCtx) {
	reqPath := strings.Replace(string(ctx.RequestURI()), "/iptv/", "", 1)
	reqPathParts := strings.SplitN(reqPath, "/", 2)
	if len(reqPathParts) == 0 {
		ctx.Error("not found", fasthttp.StatusNotFound)
		return
	}

	// Debug
	log.Println("Received", reqPathParts)

	// Unescape title
	unescapedTitle, err := url.QueryUnescape(reqPathParts[0])
	if err != nil {
		ctx.Error("invalid request", http.StatusBadRequest)
		return
	}

	// Find channel
	c, ok := playlist.Channels[unescapedTitle]
	if !ok {
		ctx.Error("channel not found", http.StatusNotFound)
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
		ctx.Error("invalid request", http.StatusBadRequest)
		return
	}

	// Lock mutex if channel's type is unknown, so no other routine tries to identify it at the same time
	c.ActiveLink.Mux.Lock()
	if linkType == linkTypeUnknown {
		handleLinkUnknown(ctx, &reqPathParts[0], &unescapedTitle, link, c, c.ActiveLink)
		c.ActiveLink.Mux.Unlock()
		return
	}
	c.ActiveLink.Mux.Unlock()

	// Understand what do we need to do with this link
	switch linkType {
	case linkTypeMedia:
		handleStream(ctx, &reqPathParts[0], &unescapedTitle, link, c, c.ActiveLink)
	case linkTypeStream:
		handleStream(ctx, &reqPathParts[0], &unescapedTitle, link, c, c.ActiveLink)
	case linkTypeM3U8:

		m3u8c := m3u8channels[unescapedTitle]

		var newLink string
		if len(reqPathParts) == 1 {
			newLink = m3u8c.Link()
		} else {
			newLink = m3u8c.LinkRoot() + reqPathParts[1]
		}

		if len(reqPathParts) == 1 {
			// Channel only
			handleM3U8Channel(ctx, &reqPathParts[0], &unescapedTitle, newLink, m3u8c, c.ActiveLink)
		} else {
			// Channel with data (additional path)
			handleM3U8ChannelData(ctx, &reqPathParts[0], &unescapedTitle, newLink, m3u8c, c.ActiveLink)
		}
	case linkTypeUnsupported:
		ctx.Error("unsupported channel format", http.StatusServiceUnavailable)
	default:
		ctx.Error("internal server error", http.StatusInternalServerError)
	}
}
