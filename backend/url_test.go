package backend

import (
	"testing"

	"github.com/google/go-cmp/cmp"
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
		want error
	}{
		{"not found", "/", ErrNotFound},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			_, err := BuildImageOption(tt.url)
			if err == nil {
				t.Errorf("not error")
			}
			if err != tt.want {
				t.Errorf("want %+v but got %+v", tt.want, err)
			}
		})
	}
}
