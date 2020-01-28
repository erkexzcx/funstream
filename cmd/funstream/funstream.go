package main

import (
	"flag"
	"log"

	"github.com/erkexzcx/funstream/internal/config"
)

var port = flag.String("port", "8989", "custom port number for web server")
var userAgent = flag.String("useragent", "VLC/3.0.2.LibVLC/3.0.2", "custom user agent for web requests")
var playlistPath = flag.String("playlist", "funstream_playlist.yaml", "path to the playlist")

func main() {
	flag.Parse()

	// Get playlist
	p, err := config.Playlist(*playlistPath, *userAgent)
	if err != nil {
		log.Fatalln(err)
	}

	// Start proxy server
	p.Start(*port, *userAgent)
}
