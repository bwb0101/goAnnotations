package util

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/goAnnotations/model"
)

func GetPackageNameForStructs(structs []model.Struct) (string, error) {
	if len(structs) == 0 {
		return "", nil // Need at least one struct to determine package-name
	}
	packageName := structs[0].PackageName
	for _, s := range structs {
		if s.PackageName != packageName {
			return "", fmt.Errorf("list of structs has multiple package-names")
		}
	}
	return packageName, nil
}

func getPackageNameForEnums(enums []model.Enum) (string, error) {
	if len(enums) == 0 {
		return "", nil // Need at least one enum to determine package-name
	}
	packageName := enums[0].PackageName
	for _, s := range enums {
		if s.PackageName != packageName {
			return "", fmt.Errorf("list of enums has multiple package-names")
		}
	}
	return packageName, nil
}

func GetPackageNameForEnumsOrStructs(enums []model.Enum, structs []model.Struct) (string, error) {
	if len(enums) == 0 && len(structs) == 0 {
		return "", fmt.Errorf("need at least one enum or struct to determine package-name")
	}
	var packageNameEnums, packageNameStructs string
	var err error
	if len(enums) > 0 {
		packageNameEnums, err = getPackageNameForEnums(enums)
		if err != nil {
			return "", err
		}
	}
	if len(structs) > 0 {
		packageNameStructs, err = GetPackageNameForStructs(structs)
		if err != nil {
			return "", err
		}
	}
	if packageNameEnums == packageNameStructs || packageNameStructs == "" {
		return packageNameEnums, nil
	}
	if packageNameEnums == "" {
		return packageNameStructs, nil
	}
	return "", fmt.Errorf("list of enums and structs has multiple package-names")
}

func DetermineTargetPath(inputDir string, packageName string) (string, error) {
	if inputDir == "" || packageName == "" {
		return "", fmt.Errorf("input params not set")
	}

	goPath := os.Getenv("GOPATH")
	if goPath != "" {
		// Perform some additional check when still using GOPATH
		workDir, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("error getting working dir:%s", err)
		}

		if !strings.Contains(workDir, goPath) {
			return "", fmt.Errorf("code %s lives outside GOPATH:%s", workDir, goPath)
		}
	}

	baseDir := path.Base(inputDir)
	if baseDir == "." || baseDir == packageName {
		return inputDir, nil
	}
	return fmt.Sprintf("%s/%s", inputDir, packageName), nil
}

type Info struct {
	Src            string
	TargetFilename string
	TemplateName   string
	TemplateString string
	FuncMap        template.FuncMap
	Data           interface{}
}

func Generate(twd Info) error {
	// _, _ = fmt.Fprintf(os.Stderr, "%s: Generated go file '%s' based on source '%s'\n", "Annotations", twd.TargetFilename, twd.Src)

	w, err := createFile(twd.TargetFilename)
	if err != nil {
		return err
	}
	defer w.Close()

	t := template.New(twd.TemplateName).Funcs(twd.FuncMap)
	t, err = t.Parse(twd.TemplateString)
	if err != nil {
		return err
	}
	t.DefinedTemplates()

	var b bytes.Buffer
	if err = t.Execute(&b, twd.Data); err != nil {
		return err
	}
	if bs, err := format.Source(b.Bytes()); err != nil {
		return err
	} else if _, err = w.Write(bs); err != nil {
		return err
	}

	return nil
}

func createFile(filename string) (*os.File, error) {
	err := os.MkdirAll(filepath.Dir(filename), 0777)
	if err != nil {
		return nil, err
	}
	w, err := os.Create(filename)
	if err != nil {
		return nil, err
	}
	return w, nil
}
