package v2

import (
	"context"
	"fmt"
	"image"
	"io"
	"net/http"

	"cloud.google.com/go/storage"
	"github.com/pkg/errors"
	"github.com/sinmetal/gaeimage"
	"github.com/sinmetal/goma"
	"github.com/vvakame/sdlog/aelog"
)

type ImageService struct {
	gcs  *storage.Client
	goma *goma.StorageService
}

func NewImageService(ctx context.Context, gcs *storage.Client, goma *goma.StorageService) (*ImageService, error) {
	return &ImageService{
		gcs:  gcs,
		goma: goma,
	}, nil
}

// ReadAndWrite is Cloud Storage から読み込んだImageをhttp.ResponseWriterに書き込む
// gaeimage.ImageOptionにより画像の変換が求められている場合、変換後Object保存用Bucketを参照し、すでにあればそれを書き込む
// 変換後Object保存用Bucketに変換されたObjectがない場合、変換したImageを作成し、変換後Object保存用Bucketに保存して、それを書き込む
func (s *ImageService) ReadAndWrite(ctx context.Context, w http.ResponseWriter, o *gaeimage.ImageOption) error {
	var bucket = o.Bucket
	var object = o.Object
	var resize bool
	if o.Size > 0 {
		resize = true
		bucket = s.BucketOfAlteredObject(o.Bucket)
		object = s.ObjectOfAltered(o.Object, o.Size)
	}

	var img image.Image
	var gt *goma.GomaType
	attrs, err := s.gcs.Bucket(bucket).Object(object).Attrs(ctx)
	if err == storage.ErrObjectNotExist {
		if !resize {
			return gaeimage.ErrNotFound
		}

		img, gt, err = s.ResizeToGCS(ctx, o)
		if err != nil {
			return errors.Wrap(gaeimage.ErrCreateCacheImage, fmt.Sprintf("failed ResizeToGCS() option=%+v, err=%+v", o, err))
		}

		if o.CacheControlMaxAge > 0 {
			w.Header().Set("cache-control", fmt.Sprintf("public, max-age=%d", o.CacheControlMaxAge))
		}

		// file sizeが分からなかったので、content-length付けてないが、Google Frontendが付けてくれる
		w.Header().Set("last-modified", attrs.Created.Format(http.TimeFormat))
		w.Header().Set("content-type", gt.ContentType)
		w.WriteHeader(http.StatusOK)
		if err := goma.Write(w, img, gt.FormatType); err != nil {
			aelog.Errorf(ctx, "failed goma.Write to response. err=%+v\n", err)
		}
		return nil
	} else if err != nil {
		return errors.Wrap(gaeimage.ErrInternalError, fmt.Sprintf("failed storage.object.attrs() option=%+v, err=%+v", o, err))
	}

	or, err := s.gcs.Bucket(bucket).Object(object).NewReader(ctx)
	if err != nil {
		return errors.Wrap(gaeimage.ErrInternalError, fmt.Sprintf("failed storage.object.NewReader() option=%+v, err=%+v", o, err))
	}
	if o.CacheControlMaxAge > 0 {
		w.Header().Set("cache-control", fmt.Sprintf("public, max-age=%d", o.CacheControlMaxAge))
	}
	w.Header().Set("last-modified", attrs.Created.Format(http.TimeFormat))
	w.Header().Set("content-length", fmt.Sprintf("%d", attrs.Size))
	w.Header().Set("content-type", attrs.ContentType)
	w.WriteHeader(http.StatusOK)
	_, err = io.Copy(w, or)
	if err != nil {
		return errors.Wrap(gaeimage.ErrInternalError, fmt.Sprintf("failed gcs.object copy to response. err=%+v\n", err))
	}
	return nil
}

// ResizeToGCS is 画像をリサイズしてCloud Storageに保存する
func (s *ImageService) ResizeToGCS(ctx context.Context, o *gaeimage.ImageOption) (image.Image, *goma.GomaType, error) {
	img, gt, err := s.goma.Read(ctx, o.Bucket, o.Object)
	if err != nil {
		return nil, nil, err
	}
	img = goma.ResizeToFitLongSide(img, o.Size)
	if err := s.goma.Write(ctx, img, gt.FormatType, s.BucketOfAlteredObject(o.Bucket), s.ObjectOfAltered(o.Object, o.Size)); err != nil {
		return nil, nil, err
	}
	return img, gt, nil
}

// BucketOfAlteredObject is 変換後Objectを保存するBucket
func (s *ImageService) BucketOfAlteredObject(bucket string) string {
	return fmt.Sprintf("alter-%s", bucket)
}

// ObjectOfAltered is 変換後Object Name
func (s *ImageService) ObjectOfAltered(object string, size int) string {
	return fmt.Sprintf("%s_s%d", object, size)
}
