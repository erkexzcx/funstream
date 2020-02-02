package proxy

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
)

var userAgent string

// Start starts web server and servers playlist
func (p *Playlist) Start(port, userAgentString string) {
	playlist = p
	userAgent = userAgentString

	// Some global vars
	m3u8channels = make(map[string]*M3U8Channel, len(p.Channels))

	http.HandleFunc("/iptv", playlistHandler)
	http.HandleFunc("/iptv/", channelHandler)
	http.HandleFunc("/logo/", logoHandler)

	log.Println("Web server should be started by now!")

	printAvailableAddresses(port)

	panic(http.ListenAndServe(":"+port, nil))
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
