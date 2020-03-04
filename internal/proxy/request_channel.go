package proxy

import "net/http"

func channelHandler(w http.ResponseWriter, r *http.Request) {
	sr, err := getStreamRequest(w, r, "/iptv/")
	if err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	// Lock this channel, so no other routine can request upstream servers
	sr.Channel.LinksMux.Lock()
	defer sr.Channel.LinksMux.Unlock()

	streamRequestHandler(w, r, sr)
}

func streamRequestHandler(w http.ResponseWriter, r *http.Request, sr *StreamRequest) {
	switch sr.Channel.ActiveLink.LinkType {
	case linkTypeUnknown:
		handleContentUnknown(w, r, sr)
	case linkTypeM3U8:
		handleContentM3U8(w, r, sr)
	case linkTypeMedia:
		handleContentMedia(w, r, sr)
	default:
		http.Error(w, "invalid media type", http.StatusInternalServerError)
	}
}
