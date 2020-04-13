package gaeimage

import (
	"github.com/sinmetal/gaeimage"
)

type ImageOption struct {
	Bucket             string
	Object             string
	Size               int
	CacheControlMaxAge int // Pathから指定はできず、App側で指定する
}

// BuildImageOption is Request URLからImageOptionを生成する
// App Engine Image Serviceと同じ雰囲気のURLを利用する時に使う
//
// 期待する形式
// `/{bucket}/{object}`
// `/{bucket}/{object}/=sXXX`
func BuildImageOption(path string) (*ImageOption, error) {
	o, err := gaeimage.BuildImageOption(path)
	if err != nil {
		return nil, err
	}
	return &ImageOption{
		Bucket:             o.Bucket,
		Object:             o.Object,
		Size:               o.Size,
		CacheControlMaxAge: o.CacheControlMaxAge,
	}, nil
}
