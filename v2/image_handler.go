package v2

import (
	"fmt"
	"net/http"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/sinmetal/gaeimage"
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

	o, err := gaeimage.BuildImageOption(strings.Join(l[1:], "/"))
	if err == gaeimage.ErrInvalidArgument {
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte("invalid argument"))
		if err != nil {
			fmt.Printf("failed write to response. err%+v", err)
		}
		return
	} else if err == gaeimage.ErrResizeArgument {
		w.WriteHeader(http.StatusBadRequest)
		if err != nil {
			fmt.Printf("failed write to response. err%+v", err)
		}
		return
	} else if err != nil {
		fmt.Printf("failed %+v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	o.CacheControlMaxAge = 3600

	gcs, err := storage.NewClient(ctx)
	if err != nil {
		fmt.Printf("failed storage.NewClient %+v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	gomas := goma.NewStorageService(ctx, gcs)
	is, err := NewImageService(ctx, gcs, gomas)
	if err != nil {
		fmt.Printf("failed NewImageService %+v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = is.ReadAndWrite(ctx, w, o)
	if err == gaeimage.ErrNotFound {
		fmt.Printf("404: bucket=%v,object=%v,err=%+v\n", o.Bucket, o.Object, err)
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Printf("failed ReadAndWrite bucket=%v,object=%v err=%+v\n", o.Bucket, o.Object, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
