package gaeimage

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestImageHandler(t *testing.T) {
	r := httptest.NewRequest("GET", "https://example.com/v1/sinmetal/shingo_nouhau.png/=s32", nil)
	w := httptest.NewRecorder()

	ImageHandler(w, r)

	resp := w.Result()

	if e, g := http.StatusOK, resp.StatusCode; e != g {
		body, _ := ioutil.ReadAll(resp.Body)
		t.Errorf("StatusCode want %v got %v. body=%v", e, g, string(body))
	}
}
