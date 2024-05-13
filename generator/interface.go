package generator

import (
	"github.com/goAnnotations/model"
)

const (
	GenfilePrefix       = "gen_"
	GenfileExcludeRegex = GenfilePrefix + ".*"
)

type Generator interface {
	Generate(inputDir string, parsedSources model.ParsedSources) error
}
