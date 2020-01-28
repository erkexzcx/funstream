package proxy

import (
	"fmt"
	"net/url"

	"github.com/valyala/fasthttp"
)

func playlistHandler(ctx *fasthttp.RequestCtx) {
	// Write HTTP headers
	ctx.SetStatusCode(fasthttp.StatusOK)

	// Write HTTP body
	fmt.Fprintln(ctx, "#EXTM3U")
	for _, title := range playlist.OrderedTitles {
		link := "http://" + string(ctx.Host()) + "/iptv/" + url.QueryEscape(title)
		logo := "http://" + string(ctx.Host()) + "/logo/" + url.QueryEscape(title)
		fmt.Fprintf(ctx, "#EXTINF:-1 tvg-logo=\"%s\" group-title=\"%s\", %s\n%s\n", logo, (playlist.Channels[title]).Group, title, link)
	}
}
