package playlist

import (
	"testing"
)

var testDataNormalize [][2]string = [][2]string{
	{"Trt World HD", "TRT World HD"},
	{"LNK HD", "LNK HD"},
	{"TV3 HD TEST", "TV3 HD Test"},
	{"tv3", "TV3"},
	{"РоссияTV", "РоссияTV"},
	{"Россия 1 HD", "Россия 1 HD"},
}

func TestNormalize(t *testing.T) {
	for _, pair := range testDataNormalize {
		res := normalize(pair[0])
		if res != pair[1] {
			t.Errorf("Incorrect normalization! Excepted %s, got %s", pair[1], res)
		}
	}
}
