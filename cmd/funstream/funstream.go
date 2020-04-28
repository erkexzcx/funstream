package main

import (
	"flag"
	"log"

	"github.com/erkexzcx/funstream/internal/playlist"
)

var flagBind = flag.String("bind", "0.0.0.0:8989", "bind IP and port")
var flagUserAgent = flag.String("useragent", "VLC/3.0.2.LibVLC/3.0.2", "custom user agent for web requests")
var flagPlaylist = flag.String("playlist", "funstream_playlist.yaml", "path to the playlist")

func main() {
	flag.Parse()

	// Get playlist
	p, err := playlist.Playlist(flagPlaylist, flagUserAgent)
	if err != nil {
		log.Fatalln(err)
	}

	// Start proxy server
	p.Start(flagBind, flagUserAgent)
}
