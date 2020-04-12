package gaeimage

import "github.com/pkg/errors"

var ErrInvalidArgument = errors.New("invalid argument")
var ErrResizeArgument = errors.New("invalid resize argument")
var ErrCreateCacheImage = errors.New("failed create cache image")
var ErrNotFound = errors.New("not found")
var ErrInternalError = errors.New("internal error")
