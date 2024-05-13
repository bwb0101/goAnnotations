package generator

import (
	"github.com/bwb0101/goAnnotations/model"
)

const (
	GenfilePrefix       = "gen_"
	GenfileExcludeRegex = GenfilePrefix + ".*"
)

type Generator interface {
	Generate(inputDir string, parsedSources model.ParsedSources) error
}
