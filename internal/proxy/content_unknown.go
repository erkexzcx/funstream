package proxy

import (
	"log"
	"net/http"
)

func handleContentUnknown(w http.ResponseWriter, r *http.Request, sr *StreamRequest) {
	resp, err := getResponse(sr.Channel.ActiveLink.Link)
	if err != nil {
		log.Println("Failed to request link. Cycling and retrying...", err, sr.Channel.ActiveLink.Link)
		cycleAndRetry(w, r, sr)
		return
	}
	defer resp.Body.Close()

	sr.Channel.ActiveLink.LinkType = getLinkType(resp.Header.Get("Content-Type"))
	switch sr.Channel.ActiveLink.LinkType {
	case linkTypeMedia:
		handleEstablishedContentMedia(w, r, sr, resp)
	case linkTypeM3U8:
		// Create new M3u8 type channel
		sr.Channel.ActiveLink.M3u8Ref = &M3U8Channel{Channel: sr.Channel}
		sr.Channel.ActiveLink.M3u8Ref.link = resp.Request.URL.String()
		sr.Channel.ActiveLink.M3u8Ref.linkRoot = deleteAfterLastSlash(sr.Channel.ActiveLink.M3u8Ref.link)
		handleEstablishedContentM3U8(w, r, sr, resp, sr.Channel.ActiveLink.Link)
	default:
		http.Error(w, "invalid media type", http.StatusInternalServerError)
	}
}
