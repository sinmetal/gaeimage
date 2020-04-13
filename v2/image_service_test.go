package gaeimage

import (
	"context"
	"testing"
)

func TestWithAlterBucket(t *testing.T) {
	ctx := context.Background()

	const bucket = "Hello"
	is, err := NewImageService(ctx, nil, nil, WithAlterBucket(bucket))
	if err != nil {
		t.Fatal(err)
	}

	if e, g := bucket, is.alterBucket; e != g {
		t.Errorf("want %v but got %v", e, g)
	}
}
