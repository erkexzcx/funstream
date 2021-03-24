package main

import (
	"flag"
	"log"

	"github.com/erkexzcx/funstream/playlist"
)

var flagBind = flag.String("bind", "0.0.0.0:8888", "bind IP and port")
var flagPlaylist = flag.String("playlist", "funstream.yml", "path to the playlist")

//var flagUserAgent = flag.String("useragent", "VLC/3.0.2.LibVLC/3.0.2", "custom user agent for web requests")

func main() {
	flag.Parse()

	// Get playlist
	p, err := playlist.Playlist(flagPlaylist)
	if err != nil {
		log.Fatalln(err)
	}

	// Start proxy server
	p.Start(flagBind)
}
