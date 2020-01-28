package proxy

import (
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/valyala/fasthttp"
)

func handleM3U8Channel(ctx *fasthttp.RequestCtx, escapedTitle, unescapedTitle *string, link string, c *M3U8Channel, l *Link) {
	retry := func(newLink string) {
		handleM3U8Channel(ctx, escapedTitle, unescapedTitle, newLink, c, l)
	}

	cycleAndRetry := func() {
		// Link is not working - try to switch to the next one and reload
		res := c.Channel.cycleLink()
		if !res {
			ctx.Error("no working channels", http.StatusInternalServerError)
			return
		}
		l = c.Channel.ActiveLink
		newLink := l.Link
		handleM3U8Channel(ctx, escapedTitle, unescapedTitle, newLink, c, l)
	}

	log.Println("Channel type is M3U8 (channel only) and working on it!")

	req, resp, err := getRequest(link, -1)
	defer fasthttp.ReleaseResponse(resp)
	defer fasthttp.ReleaseRequest(req)
	if err != nil {
		log.Println("Failed to request link. Cycling and retrying...")
		cycleAndRetry()
		return
	}

	statusCode := resp.StatusCode()

	// If redirect
	if statusCode >= 300 && statusCode < 400 {
		log.Println("Got redirection response...")
		myURL, err := url.Parse(link)
		if err != nil {
			log.Println("Failed to build redirect url. Cycling and retrying...")
			cycleAndRetry()
			return
		}
		nextURL, err := url.Parse(string(resp.Header.Peek("Location")))
		if err != nil {
			log.Println("Failed to build redirect url. Cycling and retrying...")
			cycleAndRetry()
			return
		}
		newLink := myURL.ResolveReference(nextURL).String()
		retry(newLink)
		return
	}

	// If not OK
	if statusCode < 200 || statusCode >= 300 {
		log.Println("Got not OK response. Cycling and retrying...")
		cycleAndRetry()
		return
	}

	// Update URL in case we got redirect:
	c.newRedirectedLink(link)

	// Set timestamp of last session update:
	c.SetSessionUpdatedNow()

	linkRoot := c.LinkRoot()
	prefix := "http://" + string(ctx.Host()) + "/iptv/" + *escapedTitle + "/"
	origContent := string(resp.Body())
	content := []byte(rewriteLinks(origContent, prefix, linkRoot))

	// Cache mux is already locked
	c.SetLinkCache(content)
	c.SetLinkCacheCreatedNow()

	ctx.SetContentTypeBytes(resp.Header.ContentType())
	ctx.SetStatusCode(http.StatusOK)
	ctx.SetBody(content)
}

func handleM3U8ChannelData(ctx *fasthttp.RequestCtx, escapedTitle, unescapedTitle *string, link string, c *M3U8Channel, l *Link) {
	retry := func(newLink string) {
		handleM3U8ChannelData(ctx, escapedTitle, unescapedTitle, newLink, c, l)
	}

	cycleAndRetry := func() {
		// Link is not working - try to switch to the next one and reload
		res := c.Channel.cycleLink()
		if !res {
			ctx.Error("no working channels", http.StatusInternalServerError)
			return
		}
		l = c.Channel.ActiveLink
		newLink := l.Link
		handleM3U8ChannelData(ctx, escapedTitle, unescapedTitle, newLink, c, l)
	}

	// Try to load from cache first
	if contents, ok := m3u8TSCache.Get(link); ok {
		log.Println("Serving media cache...")
		ctx.Success("application/vnd.apple.mpegurl", contents.([]byte))
		return
	}

	req, resp, err := getRequest(link, -1)
	defer fasthttp.ReleaseResponse(resp)
	defer fasthttp.ReleaseRequest(req)
	if err != nil {
		log.Println("Failed to request link. Cycling and retrying...")
		cycleAndRetry()
		return
	}

	statusCode := resp.StatusCode()

	// If redirect
	if statusCode >= 300 && statusCode < 400 {
		log.Println("Got redirection response...")
		myURL, err := url.Parse(link)
		if err != nil {
			log.Println("Failed to build redirect url. Cycling and retrying...")
			cycleAndRetry()
			return
		}
		nextURL, err := url.Parse(string(resp.Header.Peek("Location")))
		if err != nil {
			log.Println("Failed to build redirect url. Cycling and retrying...")
			cycleAndRetry()
			return
		}
		newLink := myURL.ResolveReference(nextURL).String()
		retry(newLink)
		return
	}

	// If not OK
	if statusCode < 200 || statusCode >= 300 {
		log.Println("Got not OK response. Cycling and retrying...")
		cycleAndRetry()
		return
	}

	// Set session update timestamp
	c.SetSessionUpdatedNow()

	// Find content type
	contentType := string(resp.Header.ContentType())

	if (contentType == "application/vnd.apple.mpegurl" || contentType == "application/x-mpegurl") && strings.Contains(strings.ToLower(link), ".m3u8") {
		// If we reach this code block - it means we got redirect without HTTP code 3**
		c.newRedirectedLink(link)

		linkRoot := c.LinkRoot()
		prefix := "http://" + string(ctx.Host()) + "/iptv/" + *escapedTitle + "/"
		origContent := string(resp.Body())
		content := []byte(rewriteLinks(origContent, prefix, linkRoot))

		ctx.SetContentType(contentType)
		ctx.SetStatusCode(statusCode)
		ctx.SetBody(content)
	} else if strings.HasPrefix(contentType, "video/") || strings.HasPrefix(contentType, "audio/") {
		// TS files
		content := resp.Body()
		ctx.SetContentType(contentType)
		ctx.SetStatusCode(200)
		ctx.SetBody(content)

		go m3u8TSCache.SetDefault(link, content) // Save to cache
	}

}
