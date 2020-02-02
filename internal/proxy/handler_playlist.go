package proxy

import (
	"fmt"
	"net/http"
	"net/url"
)

func playlistHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "#EXTM3U")
	for _, title := range playlist.OrderedTitles {
		link := "http://" + r.Host + "/iptv/" + url.QueryEscape(title)
		logo := "http://" + r.Host + "/logo/" + url.QueryEscape(title)
		fmt.Fprintf(w, "#EXTINF:-1 tvg-logo=\"%s\" group-title=\"%s\", %s\n%s\n", logo, playlist.Channels[title].Group, title, link)
	}
}
