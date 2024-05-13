/*
 * 项目名称：Annotations
 * 文件名：parser.go
 * 日期：2024/05/06 16:52
 * 作者：Ben
 */

package parser

import (
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"regexp"
	"sort"

	"github.com/Annotations/model"
)

type fileEntry struct {
	key  string
	file ast.File
}

type fileEntries []fileEntry

func (list fileEntries) Len() int {
	return len(list)
}

func (list fileEntries) Less(i, j int) bool {
	return list[i].key < list[j].key
}

func (list fileEntries) Swap(i, j int) {
	list[i], list[j] = list[j], list[i]
}

func ParseSourceDir(dirName string, includeRegex string, excludeRegex string) (model.ParsedSources, error) {
	packages, err := parseDir(dirName, includeRegex, excludeRegex)
	if err != nil {
		log.Printf("error parsing dir %s: %s", dirName, err.Error())
		return model.ParsedSources{}, err
	}
	v := &astVisitor{
		Imports: map[string]string{},
	}
	for _, aPackage := range packages {
		parsePackage(aPackage, v)
	}

	embedOperationsInStructs(v)

	embedTypedefDocLinesInEnum(v)

	return model.ParsedSources{
		Structs:    v.Structs,
		Operations: v.Operations,
		Interfaces: v.Interfaces,
		Typedefs:   v.Typedefs,
		Enums:      v.Enums,
	}, nil
}

func parseDir(dirName string, includeRegex string, excludeRegex string) (map[string]*ast.Package, error) {
	var includePattern = regexp.MustCompile(includeRegex)
	var excludePattern = regexp.MustCompile(excludeRegex)

	fileSet := token.NewFileSet()
	packageMap, err := parser.ParseDir(fileSet, dirName, func(fi os.FileInfo) bool {
		if excludePattern.MatchString(fi.Name()) {
			return false
		}
		return includePattern.MatchString(fi.Name())
	}, parser.ParseComments)
	if err != nil {
		log.Printf("error parsing dir %s: %s", dirName, err.Error())
		return packageMap, err
	}

	return packageMap, nil
}

func parsePackage(aPackage *ast.Package, v *astVisitor) {

	for _, fileEntry := range sortedFileEntries(aPackage.Files) {
		v.CurrentFilename = fileEntry.key
		appEngineOnly := true
		for _, commentGroup := range fileEntry.file.Comments {
			if commentGroup != nil {
				for _, comment := range commentGroup.List {
					if comment != nil && comment.Text == "// +build !appengine" {
						appEngineOnly = false
					}
				}
			}
		}
		if appEngineOnly {
			ast.Walk(v, &fileEntry.file)
		}
	}
}

func sortedFileEntries(fileMap map[string]*ast.File) fileEntries {
	var fileEntries fileEntries = make([]fileEntry, 0, len(fileMap))
	for key, file := range fileMap {
		if file != nil {
			fileEntries = append(fileEntries, fileEntry{
				key:  key,
				file: *file,
			})
		}
	}
	sort.Sort(fileEntries)
	return fileEntries
}

func embedOperationsInStructs(visitor *astVisitor) {
	mStructMap := make(map[string]*model.Struct)
	for idx := range visitor.Structs {
		mStructMap[(&visitor.Structs[idx]).Name] = &visitor.Structs[idx]
	}
	for idx := range visitor.Operations {
		mOperation := visitor.Operations[idx]
		if mOperation.RelatedStruct != nil {
			if mStruct, ok := mStructMap[mOperation.RelatedStruct.DereferencedTypeName()]; ok {
				mStruct.Operations = append(mStruct.Operations, &mOperation)
			}
		}
	}

}

func embedTypedefDocLinesInEnum(visitor *astVisitor) {
	for idx, mEnum := range visitor.Enums {
		for _, typedef := range visitor.Typedefs {
			if typedef.Name == mEnum.Name {
				visitor.Enums[idx].DocLines = typedef.DocLines
				break
			}
		}
	}
}
