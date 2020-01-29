package proxy

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/valyala/fasthttp"
)

// Mutex is locked when working in this function!
func handleLinkUnknown(ctx *fasthttp.RequestCtx, escapedTitle, unescapedTitle *string, link string, c *Channel, l *Link) {
	retry := func(newLink string) {
		handleLinkUnknown(ctx, escapedTitle, unescapedTitle, newLink, c, l)
	}

	cycleAndRetry := func() {
		// Link is not working - try to switch to the next one and reload
		res := c.cycleLinkNoMux()
		if !res {
			ctx.Error("no working channels", http.StatusInternalServerError)
			return
		}
		l = c.ActiveLink
		newLink := l.Link
		handleLinkUnknown(ctx, escapedTitle, unescapedTitle, newLink, c, l)
	}

	log.Println("Channel type is unknown and working on it!")

	// We don't know what to expect, so just load URL and check content type of response
	resp, err := getRequestStandard(link, -1)
	if err != nil {
		log.Println("Failed to request link. Cycling and retrying...")
		cycleAndRetry()
		return
	}

	statusCode := resp.StatusCode

	if statusCode < 200 || statusCode >= 300 {
		defer resp.Body.Close()
	}

	// If redirect
	if statusCode >= 300 && statusCode < 400 {
		log.Println("Got redirection response...")
		myURL, err := url.Parse(link)
		if err != nil {
			log.Println("Failed to build redirect url. Cycling and retrying...")
			cycleAndRetry()
			return
		}
		nextURL, err := url.Parse(resp.Header.Get("Location"))
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

	contentType := resp.Header.Get("Content-Type")
	l.LinkType = getLinkType(contentType)

	switch l.LinkType {
	case linkTypeUnsupported:
		ctx.Error("Unsupported channel format", http.StatusServiceUnavailable)
	case linkTypeM3U8:
		log.Println("Processing type: M3U8")
		defer resp.Body.Close()
		// Create new M3u8 type channel
		m3u8c := &M3U8Channel{Channel: c}
		m3u8channels[*unescapedTitle] = m3u8c

		m3u8c.link = resp.Request.URL.String()
		m3u8c.linkRoot = deleteAfterLastSlash(m3u8c.link)

		prefix := "http://" + string(ctx.Host()) + "/iptv/" + *escapedTitle + "/"
		origContent, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			ctx.Error("internal server error", http.StatusInternalServerError)
			return
		}
		content := []byte(rewriteLinks(string(origContent), prefix, m3u8c.linkRoot))

		ctx.SetContentType(contentType)
		ctx.SetStatusCode(200)
		ctx.SetBody(content)

		m3u8c.linkCache = content
		m3u8c.linkCacheCreated = time.Now()
	case linkTypeMedia:
		log.Println("Processing type: Media")
		handleEstablishedStream(ctx, resp)
	case linkTypeStream:
		log.Println("Processing type: Stream")
		handleEstablishedStream(ctx, resp)
	}
}
