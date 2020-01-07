package backend

import (
	"context"
	"fmt"
	"image"
	"io"
	"net/http"

	"cloud.google.com/go/storage"
	"github.com/sinmetal/goma"
)

func ImageHandlerV2(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	o, err := BuildImageOption(r.URL.Path)
	if err == ErrNotFound {
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

	var bucket = o.Bucket
	var object = o.Object
	var resize bool
	if o.Size > 0 {
		resize = true
		bucket = resizeBucket(o.Bucket)
		object = resizeObject(o.Object, o.Size)
	}

	var img image.Image
	var gt *goma.GomaType
	attrs, err := gcs.Bucket(bucket).Object(object).Attrs(ctx)
	if err != nil {
		if resize && err == storage.ErrObjectNotExist {
			img, gt, err = resizeToGCS(ctx, s, o)
			if err != nil {
				fmt.Printf("failed resizeToGCS bucket=%v,object=%v err=%+v\n", o.Bucket, o.Object, err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			w.Header().Set("cache-control", "public, max-age=3600")
			w.Header().Set("content-type", gt.ContentType)
			w.WriteHeader(http.StatusOK)
			if err := goma.Write(w, img, gt.FormatType); err != nil {
				fmt.Printf("failed goma.Write to response. err=%+v\n", err)
			}
			return
		} else if err == storage.ErrObjectNotExist {
			fmt.Printf("404: bucket=%v,object=%v\n", bucket, object)
			w.WriteHeader(http.StatusNotFound)
			return
		} else {
			fmt.Printf("failed gcs.Attrs bucket=%v,object=%v err=%+v\n", o.Bucket, o.Object, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	or, err := gcs.Bucket(bucket).Object(object).NewReader(ctx)
	if err != nil {
		fmt.Printf("failed gcs.NewReader bucket=%v,object=%v err=%+v\n", o.Bucket, o.Object, err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Header().Set("cache-control", "public, max-age=3600")
	w.Header().Set("content-type", attrs.ContentType)
	w.WriteHeader(http.StatusOK)
	_, err = io.Copy(w, or)
	if err != nil {
		fmt.Printf("failed gcs.Read to response. err=%+v\n", err)
	}
}

func resizeToGCS(ctx context.Context, s *goma.StorageService, o *ImageOption) (image.Image, *goma.GomaType, error) {
	img, gt, err := s.Read(ctx, o.Bucket, o.Object)
	if err != nil {
		return nil, nil, err
	}
	img = goma.ResizeToFitLongSide(img, o.Size)
	if err := s.Write(ctx, img, gt.FormatType, resizeBucket(o.Bucket), resizeObject(o.Object, o.Size)); err != nil {
		return nil, nil, err
	}
	return img, gt, nil
}

func resizeBucket(bucket string) string {
	return fmt.Sprintf("resize-%s", bucket)

}

func resizeObject(object string, size int) string {
	return fmt.Sprintf("%s_s%d", object, size)
}
