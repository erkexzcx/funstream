package proxy

import (
	"log"
	"net/http"
)

var userAgent string

// Start starts web server and servers playlist
func (p *Playlist) Start(flagBind, flagUserAgent *string) {
	playlist = p
	userAgent = *flagUserAgent

	// Some global vars
	m3u8channels = make(map[string]*M3U8Channel, len(p.Channels))

	http.HandleFunc("/iptv", playlistHandler)
	http.HandleFunc("/iptv/", channelHandler)
	http.HandleFunc("/logo/", logoHandler)

	log.Println("Web server should be started!")

	panic(http.ListenAndServe(*flagBind, nil))
}
