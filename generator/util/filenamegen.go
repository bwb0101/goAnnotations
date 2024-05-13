package util

import (
	"path"

	"github.com/bwb0101/goAnnotations/generator"
)

func Prefixed(filenamePath string) string {
	dir, filename := path.Split(filenamePath)
	return dir + generator.GenfilePrefix + filename
}
