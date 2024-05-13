package generator

import (
	"Annotations/model"
)

const (
	GenfilePrefix       = "gen_"
	GenfileExcludeRegex = GenfilePrefix + ".*"
)

type Generator interface {
	Generate(inputDir string, parsedSources model.ParsedSources) error
}
