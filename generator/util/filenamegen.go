package util

import (
	"path"

	"github.com/Annotations/generator"
)

func Prefixed(filenamePath string) string {
	dir, filename := path.Split(filenamePath)
	return dir + generator.GenfilePrefix + filename
}
