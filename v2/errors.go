package gaeimage

import (
	"github.com/morikuni/failure"
	"github.com/sinmetal/gaeimage"
)

func IsErrInvalidArgument(err error) bool {
	return failure.Is(err, gaeimage.InvalidArgument)
}

func IsErrNotFound(err error) bool {
	return failure.Is(err, gaeimage.NotFound)
}

func IsErrInternalError(err error) bool {
	return failure.Is(err, gaeimage.InternalError)
}
