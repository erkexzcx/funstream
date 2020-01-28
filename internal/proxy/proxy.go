package proxy

import (
	"log"
	"strings"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/valyala/fasthttp"
)

var userAgent string

// Start starts web server and servers playlist
func (p *Playlist) Start(port, userAgentString string) {
	playlist = p
	userAgent = userAgentString

	// Some global vars
	m3u8channels = make(map[string]*M3U8Channel, len(p.Channels))
	m3u8TSCache = cache.New(time.Minute, 10*time.Second) // Store cache for 1 minute and clear every 10 seconds

	m := func(ctx *fasthttp.RequestCtx) {
		path := string(ctx.Path())
		switch {
		case path == "/iptv":
			playlistHandler(ctx)
		case strings.HasPrefix(path, "/iptv/"):
			channelHandler(ctx)
		case strings.HasPrefix(path, "/logo/"):
			logoHandler(ctx)
		default:
			ctx.Error("not found", fasthttp.StatusNotFound)
		}
	}

	log.Println("Web server should be started by now!")
	panic(fasthttp.ListenAndServe(":"+port, m))
}
