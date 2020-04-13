package gaeimage

import (
	"net/http"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/sinmetal/goma"
	"github.com/vvakame/sdlog/aelog"
)

func ImageHandler(w http.ResponseWriter, r *http.Request) {
	ctx := aelog.WithHTTPRequest(r.Context(), r)

	l := strings.Split(r.URL.Path, "/")
	if len(l) < 4 {
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte("invalid argument"))
		if err != nil {
			aelog.Errorf(ctx, "failed write to response. err%+v", err)
		}
		return
	}

	o, err := BuildImageOption(strings.Join(l[1:], "/"))
	if IsErrInvalidArgument(err) {
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte("invalid argument"))
		if err != nil {
			aelog.Errorf(ctx, "failed write to response. err%+v", err)
		}
		return
	} else if err != nil {
		aelog.Errorf(ctx, "failed %+v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	o.CacheControlMaxAge = 3600

	gcs, err := storage.NewClient(ctx)
	if err != nil {
		aelog.Errorf(ctx, "failed storage.NewClient %+v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	gomas := goma.NewStorageService(ctx, gcs)
	is, err := NewImageService(ctx, gcs, gomas)
	if err != nil {
		aelog.Errorf(ctx, "failed NewImageService %+v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = is.ReadAndWrite(ctx, w, o)
	if IsErrNotFound(err) {
		aelog.Infof(ctx, "404: bucket=%v,object=%v,err=%+v", o.Bucket, o.Object, err)
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		aelog.Errorf(ctx, "failed ReadAndWrite bucket=%v,object=%v err=%+v\n", o.Bucket, o.Object, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
