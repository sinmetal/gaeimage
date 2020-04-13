package gaeimage_test

import (
	"fmt"

	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/sinmetal/gaeimage/v2"
)

func TestImageHandlerV2_NoResize(t *testing.T) {
	// 適当なサイズで2回やってみる
	r := httptest.NewRequest("GET", "https://example.com/v2/sinmetal-image-service-dev/jun0.jpg", nil)
	w := httptest.NewRecorder()

	ImageHandler(w, r)

	resp := w.Result()

	if e, g := http.StatusOK, resp.StatusCode; e != g {
		body, _ := ioutil.ReadAll(resp.Body)
		t.Errorf("StatusCode want %v got %v. body=%v", e, g, string(body))
	}
}

func TestImageHandlerV2_Resize(t *testing.T) {
	// 適当なサイズで2回やってみる
	size := rand.Intn(600)
	for i := 0; i < 2; i++ {

		r := httptest.NewRequest("GET", fmt.Sprintf("https://example.com/v2/sinmetal-image-service-dev/jun0.jpg/=s%d", size), nil)
		w := httptest.NewRecorder()

		ImageHandler(w, r)

		resp := w.Result()

		if e, g := http.StatusOK, resp.StatusCode; e != g {
			body, _ := ioutil.ReadAll(resp.Body)
			t.Errorf("StatusCode want %v got %v. body=%v", e, g, string(body))
		}
	}
}
