package gaeimage

import (
	"fmt"
	"net/http"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/morikuni/failure"
	"github.com/sinmetal/goma"
)

func ImageHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	l := strings.Split(r.URL.Path, "/")
	if len(l) < 4 {
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte("invalid argument"))
		if err != nil {
			fmt.Printf("failed write to response. err%+v", err)
		}
	}

	o, err := BuildImageOption(strings.Join(l[1:], "/"))
	if failure.Is(err, InvalidArgument) {
		fmt.Printf("404: %+v\n", err)
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Printf("failed %+v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	gcs, err := storage.NewClient(ctx)
	if err != nil {
		fmt.Printf("failed storage.NewClient %+v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	s := goma.NewStorageService(ctx, gcs)

	img, gt, err := s.Read(ctx, o.Bucket, o.Object)
	if err != nil {
		fmt.Printf("failed goma.Read bucket=%v,object=%v err=%+v\n", o.Bucket, o.Object, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	const min = 0
	const max = 2560
	if o.Size < min || max <= o.Size {
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte(fmt.Sprintf("Size ranges from %d to %d", min, max)))
		if err != nil {
			fmt.Printf("failed write to response. err=%+v\n", err)
		}
		return
	}

	if o.Size > 0 {
		img = goma.ResizeToFitLongSide(img, o.Size)
	}

	w.Header().Set("cache-control", "public, max-age=3600")
	w.Header().Set("content-type", gt.ContentType)
	w.WriteHeader(http.StatusOK)
	if err := goma.Write(w, img, gt.FormatType); err != nil {
		fmt.Printf("failed goma.Write to response. err=%+v\n", err)
	}
}
