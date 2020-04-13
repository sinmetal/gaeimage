package gaeimage

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/morikuni/failure"
	"github.com/sinmetal/gaeimage"
)

func TestBuildImageOption(t *testing.T) {
	cases := []struct {
		name string
		url  string
		want *ImageOption
	}{
		{"s32", "/hoge/fuga/=s32", &ImageOption{Bucket: "hoge", Object: "fuga", Size: 32}},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := BuildImageOption(tt.url)
			if err != nil {
				t.Fatal(err)
			}
			if !cmp.Equal(got, tt.want) {
				t.Errorf("want %+v but got %+v", tt.want, got)
			}
		})
	}
}

func TestBuildImageOptionError(t *testing.T) {
	cases := []struct {
		name string
		url  string
		want failure.StringCode
	}{
		{"invalid argument", "/", gaeimage.InvalidArgument},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			_, err := BuildImageOption(tt.url)
			if err == nil {
				t.Errorf("not error")
			}
			code, ok := failure.CodeOf(err)
			if !ok {
				t.Errorf("want %+v but got nothing code. err=%+v", tt.want, err)
			}
			if e, g := tt.want, code; e != g {
				t.Errorf("want %+v but got %+v", e, g)
			}
		})
	}
}
