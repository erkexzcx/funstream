package proxy

import (
	"log"
	"net/http"
)

func handleContentUnknown(w http.ResponseWriter, r *http.Request, cr *ContentRequest) {
	resp, err := response(cr.Channel.ActiveLink.Link)
	if err != nil {
		log.Println("Link request failed. Trying next one...", err, cr.Channel.ActiveLink.Link)
		cycleAndRetry(w, r, cr)
		return
	}
	defer resp.Body.Close()

	cr.Channel.ActiveLink.LinkType = getLinkType(resp.Header.Get("Content-Type"))
	switch cr.Channel.ActiveLink.LinkType {
	case linkTypeMedia:
		handleEstablishedContentMedia(w, r, cr, resp)
	case linkTypeM3U8:
		// Create new M3u8 type channel
		cr.Channel.ActiveLink.M3u8Ref = &M3U8Channel{Channel: cr.Channel}
		cr.Channel.ActiveLink.M3u8Ref.link = resp.Request.URL.String()
		cr.Channel.ActiveLink.M3u8Ref.linkRoot = deleteAfterLastSlash(cr.Channel.ActiveLink.M3u8Ref.link)
		handleEstablishedContentM3U8(w, r, cr, resp, cr.Channel.ActiveLink.Link)
	default:
		http.Error(w, "invalid media type", http.StatusInternalServerError)
	}
}
