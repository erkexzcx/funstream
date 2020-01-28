package proxy

import (
	"log"
	"net/http"
	"net/url"

	"github.com/valyala/fasthttp"
)

// As by now - I have no idea how to create multiple streams out of single stream, so we just copy/paste stream bits and that's it
// In this file both media and octet streams

func handleStream(ctx *fasthttp.RequestCtx, escapedTitle, unescapedTitle *string, link string, c *Channel, l *Link) {
	retry := func(newLink string) {
		handleStream(ctx, escapedTitle, unescapedTitle, newLink, c, l)
	}

	cycleAndRetry := func() {
		// Link is not working - try to switch to the next one and reload
		res := c.cycleLink()
		if !res {
			ctx.Error("no working channels", http.StatusInternalServerError)
			return
		}
		l = c.ActiveLink
		newLink := l.Link
		handleStream(ctx, escapedTitle, unescapedTitle, newLink, c, l)
	}

	log.Println("Channel type is Media/Stream and working on it!")

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

	handleEstablishedStream(ctx, resp)
}

func handleEstablishedStream(ctx *fasthttp.RequestCtx, resp *http.Response) {
	// TODO - resource leak here - resp.Body is not closed.
	// How TF should I do it??? Any other solution just prevents VLC player
	// from playing content if I close... :/

	ctx.SetStatusCode(http.StatusOK)
	ctx.SetContentType(resp.Header.Get("Content-Type"))
	ctx.SetBodyStream(resp.Body, -1)
}
