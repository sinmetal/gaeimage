package backend

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestImageHandlerV2(t *testing.T) {
	// 適当なサイズで2回やって、
	size := rand.Intn(600)
	for i := 0; i < 2; i++ {

		r := httptest.NewRequest("GET", fmt.Sprintf("https://example.com/v2/sinmetal/shingo_nouhau.png/=s%d", size), nil)
		w := httptest.NewRecorder()

		ImageHandlerV2(w, r)

		resp := w.Result()

		if e, g := http.StatusOK, resp.StatusCode; e != g {
			body, _ := ioutil.ReadAll(resp.Body)
			t.Errorf("StatusCode want %v got %v. body=%v", e, g, string(body))
		}
	}
}
