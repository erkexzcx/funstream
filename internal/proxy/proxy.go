package proxy

import (
	"fmt"
	"log"
	"net"
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
	m3u8TSCache = cache.New(30*time.Second, 10*time.Second) // Store cache for 1 minute and clear every 10 seconds

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

	printAvailableAddresses(port)

	panic(fasthttp.ListenAndServe(":"+port, m))
}

func printAvailableAddresses(port string) {
	fmt.Println()
	fmt.Println("Available URLs:")
	addresses, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println("\tUnable to retrieve IP addresses...")
		return
	}
	for _, v := range addresses {
		address := v.String()
		if strings.Contains(address, "::") {
			continue
		}
		fmt.Println("\thttp://" + strings.SplitN(address, "/", 2)[0] + ":" + port + "/iptv")
	}
	fmt.Println()
}
