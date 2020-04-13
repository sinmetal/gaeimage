package gaeimage

import (
	"context"
	"fmt"
	"image"
	"io"
	"net/http"
	"strconv"

	"cloud.google.com/go/storage"
	"github.com/morikuni/failure"
	"github.com/sinmetal/gaeimage"
	"github.com/sinmetal/goma"
	"github.com/vvakame/sdlog/aelog"
)

type serviceOptions struct {
	alterBucket string
}

type ServiceOption func(*serviceOptions)

// WithAlterBucket is 変換後の画像を保存するBucketを任意の固定したBucketに設定する時に利用する
// 未指定の場合の default の命名規則は `alter-{original-bucket}`
func WithAlterBucket(alterBucket string) ServiceOption {
	return func(ops *serviceOptions) {
		ops.alterBucket = alterBucket
	}
}

type ImageService struct {
	alterBucket string
	gcs         *storage.Client
	goma        *goma.StorageService
}

func NewImageService(ctx context.Context, gcs *storage.Client, goma *goma.StorageService, options ...ServiceOption) (*ImageService, error) {
	s := &ImageService{
		gcs:  gcs,
		goma: goma,
	}

	opt := serviceOptions{}
	for _, o := range options {
		o(&opt)
	}

	s.alterBucket = opt.alterBucket

	return s, nil
}

// ReadAndWrite is Cloud Storage から読み込んだImageをhttp.ResponseWriterに書き込む
// gaeimage.ImageOptionにより画像の変換が求められている場合、変換後Object保存用Bucketを参照し、すでにあればそれを書き込む
// 変換後Object保存用Bucketに変換されたObjectがない場合、変換したImageを作成し、変換後Object保存用Bucketに保存して、それを書き込む
func (s *ImageService) ReadAndWrite(ctx context.Context, w http.ResponseWriter, o *ImageOption) error {
	var bucket = o.Bucket
	var object = o.Object
	var resize bool
	if o.Size > 0 {
		resize = true
		bucket = s.BucketOfAlteredObject(o.Bucket)
		object = s.ObjectOfAltered(o.Object, o.Size)
	}

	orgAttrs, err := s.gcs.Bucket(o.Bucket).Object(o.Object).Attrs(ctx)
	if err == storage.ErrObjectNotExist {
		return failure.New(gaeimage.NotFound) // オリジナル画像がない場合はNotFoundを返す
	} else if err != nil {
		return failure.Wrap(err, failure.WithCode(gaeimage.InternalError),
			failure.Messagef("failed storage.object.attrs"),
			failure.Context{
				"bucket": o.Bucket,
				"object": o.Object,
			})
	}

	if !resize {
		or, err := s.gcs.Bucket(bucket).Object(object).NewReader(ctx)
		if err != nil {
			return failure.Wrap(err, failure.WithCode(gaeimage.InternalError),
				failure.Messagef("failed storage.object.NewReader"),
				failure.Context{
					"bucket": o.Bucket,
					"object": o.Object,
				})
		}
		if o.CacheControlMaxAge > 0 {
			w.Header().Set("cache-control", fmt.Sprintf("public, max-age=%d", o.CacheControlMaxAge))
		}
		w.Header().Set("last-modified", orgAttrs.Created.Format(http.TimeFormat))
		w.Header().Set("content-length", fmt.Sprintf("%d", orgAttrs.Size))
		w.Header().Set("content-type", orgAttrs.ContentType)
		w.WriteHeader(http.StatusOK)
		_, err = io.Copy(w, or)
		if err != nil {
			return failure.Wrap(err, failure.WithCode(gaeimage.InternalError),
				failure.Messagef("failed write to response"),
				failure.Context{
					"bucket": o.Bucket,
					"object": o.Object,
				})
		}
		return nil
	}

	var img image.Image
	var gt *goma.GomaType
	_, err = s.gcs.Bucket(bucket).Object(object).Attrs(ctx)
	if err == storage.ErrObjectNotExist {
		if !resize {
			return failure.New(gaeimage.NotFound, failure.Context{
				"bucket": o.Bucket,
				"object": o.Object,
			})
		}

		img, gt, err = s.ResizeToGCS(ctx, o)
		if err != nil {
			return failure.Wrap(err, failure.WithCode(gaeimage.InternalError),
				failure.Messagef("failed ResizeToGCS"),
				failure.Context{
					"bucket": o.Bucket,
					"object": o.Object,
					"size":   strconv.Itoa(o.Size),
				})
		}

		if o.CacheControlMaxAge > 0 {
			w.Header().Set("cache-control", fmt.Sprintf("public, max-age=%d", o.CacheControlMaxAge))
		}

		// file sizeが分からなかったので、content-length付けてないが、Google Frontendが付けてくれる
		w.Header().Set("last-modified", orgAttrs.Created.Format(http.TimeFormat)) // last-modifiedはオリジナルの画像のものを返す
		w.Header().Set("content-type", gt.ContentType)
		w.WriteHeader(http.StatusOK)
		if err := goma.Write(w, img, gt.FormatType); err != nil {
			aelog.Errorf(ctx, "failed goma.Write to response. err=%+v\n", err)
		}
		return nil
	} else if err != nil {
		return failure.Wrap(err, failure.WithCode(gaeimage.InternalError),
			failure.Messagef("failed storage.object.attrs"),
			failure.Context{
				"bucket": o.Bucket,
				"object": o.Object,
				"size":   strconv.Itoa(o.Size),
			})
	}
	return nil
}

// ResizeToGCS is 画像をリサイズしてCloud Storageに保存する
func (s *ImageService) ResizeToGCS(ctx context.Context, o *ImageOption) (image.Image, *goma.GomaType, error) {
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
	if s.alterBucket != "" {
		return s.alterBucket
	}
	return fmt.Sprintf("alter-%s", bucket)
}

// ObjectOfAltered is 変換後Object Name
func (s *ImageService) ObjectOfAltered(object string, size int) string {
	return fmt.Sprintf("%s_s%d", object, size)
}
