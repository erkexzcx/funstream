package config

import (
	"testing"
)

func TestParseRawChannel(t *testing.T) {
	rawChannel := `#EXTINF:0 tvg-id="" tvg-logo="https://tvtau.net/12345/NOVIJMIR.png" group-title="Ru" timeshift="0" catchup="append" catchup-source="&utc=${start}&lutc=${timestamp}", Novij Mir
http://tvtau.net/channel/109/hls/bn1lc?type=m3u8`
	// t.Errorf("Duplicating test data in rows %d and %d: '%s'.", k, kk, v.Provided)
	logo, group, title, link := parseRawChannel(rawChannel)
	if logo != "https://tvtau.net/12345/NOVIJMIR.png" {
		t.Errorf("Invalid logo. Excepted %s, got %s", "https://tvtau.net/12345/NOVIJMIR.png", logo)
	}
	if group != "Ru" {
		t.Errorf("Invalid group. Excepted %s, got %s", "Ru", group)
	}
	if title != "Novij Mir" {
		t.Errorf("Invalid title. Excepted %s, got %s", "Novij Mir", title)
	}
	if link != "http://tvtau.net/channel/109/hls/bn1lc?type=m3u8" {
		t.Errorf("Invalid link. Excepted %s, got %s", "http://tvtau.net/channel/109/hls/bn1lc?type=m3u8", link)
	}
}
