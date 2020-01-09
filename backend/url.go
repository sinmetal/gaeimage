package backend

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
)

var ErrNotFound = errors.New("not found")

type ImageOption struct {
	Bucket string
	Object string
	Size   int
}

func BuildImageOption(path string) (*ImageOption, error) {
	var ret ImageOption
	var found bool

	blocks := strings.Split(path, "/")
	if len(blocks) < 4 {
		return nil, ErrNotFound
	}
	ret.Bucket = blocks[2]
	ret.Object = blocks[3]

	r := regexp.MustCompile(`=s[\d]+`)

	l := r.FindAllStringSubmatch(path, -1)
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
