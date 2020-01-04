package backend

import (
	"errors"
	"regexp"
	"strconv"
)

var ErrNotFound = errors.New("not found")

type ImageOption struct {
	Size int
}

func BuildImageOption(url string) (*ImageOption, error) {
	var ret ImageOption
	var found bool

	r := regexp.MustCompile(`=s[\d]+`)

	l := r.FindAllStringSubmatch(url, -1)
	if len(l) > 0 {
		v := l[len(l)-1]
		vv := v[0]
		size, err := strconv.Atoi(vv[2:])
		if err != nil {
			return nil, err
		}
		ret.Size = size
		found = true
	}

	if found {
		return &ret, nil
	}
	return nil, ErrNotFound
}
