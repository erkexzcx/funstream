package proxy

import (
	"log"
	"net/http"
)

// Start starts web server and servers playlist
func (p *Playlist) Start(flagBind *string) {
	playlist = p

	http.HandleFunc("/iptv", playlistHandler)
	http.HandleFunc("/iptv/", channelHandler)
	http.HandleFunc("/logo/", logoHandler)

	log.Println("Web server should be started!")

	panic(http.ListenAndServe(*flagBind, nil))
}
