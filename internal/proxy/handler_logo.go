package proxy

import (
	"net/http"
	"net/url"
	"strings"
)

func logoHandler(w http.ResponseWriter, r *http.Request) {
	title := strings.Replace(r.URL.Path, "/logo/", "", 1)
	unescapedTitle, err := url.PathUnescape(title)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid request"))
		return
	}

	// Find channel reference
	channel, ok := playlist.Channels[unescapedTitle]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("channel not found"))
		return
	}

	// Find real URL of logo
	channel.LogoCacheMux.Lock()
	defer channel.LogoCacheMux.Unlock()
	if len(channel.LogoCache) == 0 {
		img, contentType, err := download(channel.Logo)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("internal server error"))
			return
		}
		channel.LogoCache = img
		channel.LogoCacheContentType = contentType
	}

	w.Header().Set("Content-Type", channel.LogoCacheContentType)
	w.Write(channel.LogoCache)
}
