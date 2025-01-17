/*
 * 项目名称：Annotations
 * 文件名：main.go
 * 日期：2024/05/06 16:23
 * 作者：Ben
 */

package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/bwb0101/goAnnotations/generator"
	"github.com/bwb0101/goAnnotations/generator/api"
	codeModel "github.com/bwb0101/goAnnotations/generator/model"
	"github.com/bwb0101/goAnnotations/model"
	"github.com/bwb0101/goAnnotations/parser"
)

const (
	excludeMatchPattern = "^" + generator.GenfilePrefix + ".*.go$"
)

var (
	dir         *string
	mode        *string
	pkgName     *string
	static_func *bool
)

func main() {
	processArgs()
	// s := "D:\\Works\\github\\goAnnotations\\test"
	// dir = &s
	pkgs, _ := parser.ParseSourceDir(*dir, "^.*.go$", excludeMatchPattern)
	// b, _ := json.MarshalIndent(pkgs, "", "\t")
	// fmt.Println(string(b))
	runAllGenerators(*dir, pkgs)
}

func runAllGenerators(inputDir string, parsedSources model.ParsedSources) {
	parsedSources.PkgName = *pkgName
	for name, g := range map[string]generator.Generator{
		"api":   api.NewGeneratorApi(),
		"model": codeModel.NewGeneratorModel(),
	} {
		err := g.Generate(inputDir, parsedSources)
		if err != nil {
			log.Printf("Error generating module %s: %s", name, err)
			os.Exit(-1)
		}
	}
}

func processArgs() {
	dir = flag.String("dir", "", "要检查的目录")
	mode = flag.String("model", "", "检查模式")
	pkgName = flag.String("pkg", "", "包名")
	static_func = flag.Bool("static_func", false, "检查非struct的方法")

	flag.Parse()

	if dir == nil || *dir == "" {
		printUsage()
	}
}

func printUsage() {
	_, _ = fmt.Fprintf(os.Stderr, "\n用法:\n")
	_, _ = fmt.Fprintf(os.Stderr, " %s [flags]\n", os.Args[0])
	flag.PrintDefaults()
	_, _ = fmt.Fprintf(os.Stderr, "\n")
	os.Exit(1)
}
