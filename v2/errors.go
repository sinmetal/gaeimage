package gaeimage

import (
	"github.com/morikuni/failure"
	"github.com/sinmetal/gaeimage"
)

func IsErrInvalidArgument(err error) bool {
	return isError(err, gaeimage.InvalidArgument)
}

func IsErrNotFound(err error) bool {
	return isError(err, gaeimage.NotFound)
}

func IsErrInternalError(err error) bool {
	return isError(err, gaeimage.InternalError)
}

func isError(err error, code failure.StringCode) bool {
	v, ok := failure.CodeOf(err)
	if !ok {
		return false
	}
	return v == code
}
