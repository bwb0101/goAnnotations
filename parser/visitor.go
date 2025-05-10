/*
 * 项目名称：Annotations
 * 文件名：visitor.go
 * 日期：2024/05/06 16:50
 * 作者：Ben
 */

package parser

import (
	"go/ast"
	"path/filepath"
	"strings"

	"github.com/bwb0101/goAnnotations/model"
)

type astVisitor struct {
	CurrentFilename string
	PackageName     string
	Filename        string
	Imports         map[string]string
	Structs         []model.Struct
	Operations      []model.Operation // 非struct的方法注解
	Interfaces      []model.Interface
	Typedefs        []model.Typedef
	Enums           []model.Enum
}

func (v *astVisitor) Visit(node ast.Node) ast.Visitor {
	if node != nil {

		// package-name is in isolated node
		if packageName, ok := extractPackageName(node); ok {
			v.PackageName = packageName
		}

		// extract all imports into a map
		v.extractGenDeclImports(node)

		v.parseAsStruct(node)
		v.parseAsTypedef(node)
		v.parseAsEnum(node)
		v.parseAsInterFace(node)
		v.parseAsOperation(node)

	}
	return v
}

func (v *astVisitor) extractGenDeclImports(node ast.Node) {
	if genDecl, ok := node.(*ast.GenDecl); ok {
		for _, spec := range genDecl.Specs {
			if importSpec, ok := spec.(*ast.ImportSpec); ok {
				quotedImport := importSpec.Path.Value
				unquotedImport := strings.Trim(quotedImport, "\"")
				init, last := filepath.Split(unquotedImport)
				if init == "" {
					last = init
				}
				v.Imports[last] = unquotedImport
			}
		}
	}
}

func (v *astVisitor) parseAsStruct(node ast.Node) {
	if mStructs := extractGenDeclForStruct(node, v.Imports); mStructs != nil {
		for _, mStruct := range mStructs {
			mStruct.PackageName = v.PackageName
			mStruct.Filename = v.CurrentFilename
			v.Structs = append(v.Structs, *mStruct)
		}
	}
}

func (v *astVisitor) parseAsTypedef(node ast.Node) {
	if mTypedef := extractGenDeclForTypedef(node); mTypedef != nil {
		mTypedef.PackageName = v.PackageName
		mTypedef.Filename = v.CurrentFilename
		v.Typedefs = append(v.Typedefs, *mTypedef)
	}
}

func (v *astVisitor) parseAsEnum(node ast.Node) {
	if mEnum := extractGenDeclForEnum(node); mEnum != nil {
		mEnum.PackageName = v.PackageName
		mEnum.Filename = v.CurrentFilename
		v.Enums = append(v.Enums, *mEnum)
	}
}

func (v *astVisitor) parseAsInterFace(node ast.Node) {
	// if interfaces, get its methods
	if mInterface := extractInterface(node, v.Imports); mInterface != nil {
		mInterface.PackageName = v.PackageName
		mInterface.Filename = v.CurrentFilename
		v.Interfaces = append(v.Interfaces, *mInterface)
	}
}

func (v *astVisitor) parseAsOperation(node ast.Node) {
	// if mOperation, get its signature
	if mOperation := extractOperation(node, v.Imports); mOperation != nil {
		mOperation.PackageName = v.PackageName
		mOperation.Filename = v.CurrentFilename
		v.Operations = append(v.Operations, *mOperation)
	}
}

func extractPackageName(node ast.Node) (string, bool) {
	if file, ok := node.(*ast.File); ok {
		if file.Name != nil {
			return file.Name.Name, true
		}
		return "", true
	}
	return "", false
}

// ------------------------------------------------------ STRUCT -------------------------------------------------------

func extractGenDeclForStruct(node ast.Node, imports map[string]string) []*model.Struct {
	if genDecl, ok := node.(*ast.GenDecl); ok {
		// Continue parsing to see if it is a struct
		if mStructs := extractSpecsForStruct(genDecl.Specs, imports); mStructs != nil {
			// Docline of struct (that could contain annotations) appear far before the details of the struct
			for _, mStruct := range mStructs {
				mStruct.DocLines = extractComments(genDecl.Doc)
			}
			return mStructs
		}
	}
	return nil
}

func extractSpecsForStruct(specs []ast.Spec, imports map[string]string) (mStructs []*model.Struct) {
	for _, spec := range specs {
		if typeSpec, ok := spec.(*ast.TypeSpec); ok {
			if structType, ok := typeSpec.Type.(*ast.StructType); ok {
				mStructs = append(mStructs, &model.Struct{
					Name:   typeSpec.Name.Name,
					Fields: extractFieldList(structType.Fields, imports),
				})
			}
		}
	}
	return
}

// ------------------------------------------------------ TYPEDEF ------------------------------------------------------

func extractGenDeclForTypedef(node ast.Node) *model.Typedef {
	if genDecl, ok := node.(*ast.GenDecl); ok {
		// Continue parsing to see if it a struct
		if mTypedef := extractSpecsForTypedef(genDecl.Specs); mTypedef != nil {
			mTypedef.DocLines = extractComments(genDecl.Doc)
			return mTypedef
		}
	}
	return nil
}

func extractSpecsForTypedef(specs []ast.Spec) *model.Typedef {
	if len(specs) >= 1 {
		if typeSpec, ok := specs[0].(*ast.TypeSpec); ok {
			mTypedef := model.Typedef{
				Name: typeSpec.Name.Name,
			}
			if ident, ok := typeSpec.Type.(*ast.Ident); ok {
				mTypedef.Type = ident.Name
			}
			return &mTypedef
		}
	}
	return nil
}

// ------------------------------------------------------- ENUM --------------------------------------------------------

func extractGenDeclForEnum(node ast.Node) *model.Enum {
	if genDecl, ok := node.(*ast.GenDecl); ok {
		// Continue parsing to see if it is an enum
		// Docs live in the related typedef
		return extractSpecsForEnum(genDecl.Specs)
	}
	return nil
}

func extractSpecsForEnum(specs []ast.Spec) *model.Enum {
	if typeName, ok := extractEnumTypeName(specs); ok {
		mEnum := model.Enum{
			Name:         typeName,
			EnumLiterals: []model.EnumLiteral{},
		}
		for _, spec := range specs {
			if valueSpec, ok := spec.(*ast.ValueSpec); ok {
				enumLiteral := model.EnumLiteral{
					Name: valueSpec.Names[0].Name,
				}
				for _, value := range valueSpec.Values {
					if basicLit, ok := value.(*ast.BasicLit); ok {
						enumLiteral.Value = strings.Trim(basicLit.Value, "\"")
						break
					}
				}
				mEnum.EnumLiterals = append(mEnum.EnumLiterals, enumLiteral)
			}
		}
		return &mEnum
	}
	return nil
}

func extractEnumTypeName(specs []ast.Spec) (string, bool) {
	for _, spec := range specs {
		if valueSpec, ok := spec.(*ast.ValueSpec); ok {
			if valueSpec.Type != nil {
				for _, name := range valueSpec.Names {
					if ident, ok := valueSpec.Type.(*ast.Ident); ok {
						if name.Obj.Kind == ast.Con {
							return ident.Name, true
						}
					}
				}
			}
		}
	}
	return "", false
}

// ----------------------------------------------------- INTERFACE -----------------------------------------------------

func extractInterface(node ast.Node, imports map[string]string) *model.Interface {
	if genDecl, ok := node.(*ast.GenDecl); ok {
		// Continue parsing to see if it an interface
		if mInterface := extractSpecsForInterface(genDecl.Specs, imports); mInterface != nil {
			// Docline of interface (that could contain annotations) appear far before the details of the struct
			mInterface.DocLines = extractComments(genDecl.Doc)
			return mInterface
		}
	}
	return nil
}

func extractSpecsForInterface(specs []ast.Spec, imports map[string]string) *model.Interface {
	if len(specs) >= 1 {
		if typeSpec, ok := specs[0].(*ast.TypeSpec); ok {
			if interfaceType, ok := typeSpec.Type.(*ast.InterfaceType); ok {
				return &model.Interface{
					Name:    typeSpec.Name.Name,
					Methods: extractInterfaceMethods(interfaceType.Methods, imports),
				}
			}
		}
	}
	return nil
}

func extractInterfaceMethods(fieldList *ast.FieldList, imports map[string]string) []model.Operation {
	methods := make([]model.Operation, 0)
	for _, field := range fieldList.List {
		if len(field.Names) > 0 {
			if funcType, ok := field.Type.(*ast.FuncType); ok {
				methods = append(methods, model.Operation{
					DocLines:   extractComments(field.Doc),
					Name:       field.Names[0].Name,
					InputArgs:  extractFieldList(funcType.Params, imports),
					OutputArgs: extractFieldList(funcType.Results, imports),
				})
			}
		}
	}
	return methods
}

// ----------------------------------------------------- OPERATION -----------------------------------------------------

func extractOperation(node ast.Node, imports map[string]string) *model.Operation {
	if funcDecl, ok := node.(*ast.FuncDecl); ok {
		mOperation := model.Operation{
			DocLines: extractComments(funcDecl.Doc),
		}

		if funcDecl.Recv != nil {
			fields := extractFieldList(funcDecl.Recv, imports)
			if len(fields) >= 1 {
				mOperation.RelatedStruct = &(fields[0])
			}
		}

		if funcDecl.Name != nil {
			mOperation.Name = funcDecl.Name.Name
		}

		if funcDecl.Type.Params != nil {
			mOperation.InputArgs = extractFieldList(funcDecl.Type.Params, imports)
		}

		if funcDecl.Type.Results != nil {
			mOperation.OutputArgs = extractFieldList(funcDecl.Type.Results, imports)
		}
		return &mOperation
	}
	return nil
}
