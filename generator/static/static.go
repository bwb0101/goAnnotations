/*
 * 项目名称：Annotations
 * 文件名：static.go
 * 日期：2024/05/06 17:34
 * 作者：Ben
 */

package static

import (
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/goAnnotations/generator"
	"github.com/goAnnotations/generator/util"
	"github.com/goAnnotations/model"
)

type templateData struct {
	PackageName string
	TargetDir   string
	// Services    []model.Operation
	Imports   map[string]string
	Codes     map[string]map[string]string
	CodesList []string
}

type Generator struct {
	targetFilename string
}

func (eg *Generator) Generate(inputDir string, parsedSources model.ParsedSources) error {
	pkgName := ""
	var datas = map[string]*templateData{}
	var data *templateData
	//
	for _, operation := range parsedSources.Operations {
		if operation.PackageName != pkgName { // 同package合成一个文件
			data = &templateData{Imports: map[string]string{`"framework/common/net_fw"`: `"framework/common/net_fw"`}, Codes: make(map[string]map[string]string)}
			pkgName = operation.PackageName
			// if targetDir, err := util.DetermineTargetPath(inputDir, pkgName); err != nil {
			// 	return err
			// } else {
			datas[pkgName] = data
			data.PackageName = pkgName
			// data.TargetDir = targetDir
			// }
		}
		parseAnnotation(operation, data)
	}
	return generate(datas, inputDir)
}

func NewGenerator() generator.Generator {
	return &Generator{}
}

func generate(datas map[string]*templateData, dir string) error {
	for _, data := range datas {
		if err := util.Generate(util.Info{
			Data:           *data,
			Src:            data.PackageName,
			TargetFilename: filepath.Join(dir, generator.GenfilePrefix+"api_handler.go"),
			TemplateName:   "static",
			TemplateString: handlersTemplate,
			FuncMap:        customTemplateFuncs,
		}); err != nil {
			return err
		}
	}
	return nil
}

func parseAnnotation(op model.Operation, data *templateData) {
	for _, line := range op.DocLines {
		line = strings.TrimSpace(line[2:])
		if strings.HasPrefix(line, "@Handler") { // @Handler(type="...")
			line = strings.TrimSpace(line[len("@Handler"):])
			line = line[1 : len(line)-1] // 去掉括号
			lines := strings.Split(line, ",")
			if strings.Contains(lines[0], "api") { // type="api"
				parseHandlerApi(lines[1:], data, op.Filename+op.Name, op.Name)
			} else if strings.Contains(lines[0], "valid.limit") {
				parseHandlerValid_limit(lines[1:], data, op.Filename+op.Name)
			} else if strings.Contains(lines[0], "valid.file") {
				parseHandlerValid_file(lines[1:], data, op.Filename+op.Name)
			}
		}
	}
}

// @Handler(type="api", net = "http/tcp", path = "/reg", bodyLimit = n, resp = "object", validation = "token")
func parseHandlerApi(words []string, data *templateData, key, apiName string) {
	if data.Codes[key] == nil {
		data.Codes[key] = map[string]string{}
		data.CodesList = append(data.CodesList, key)
	}
	data.Codes[key] = map[string]string{"api": apiName}
	for _, ll := range words {
		kv := strings.Split(ll, "=")
		switch k := strings.TrimSpace(kv[0]); k {
		case "net":
			if v := strings.TrimSpace(kv[1]); v == `"http"` {
				data.Codes[key][k] = "net_fw.HTNET_type_http"
			}
		case "path":
			data.Codes[key][k] = strings.TrimSpace(kv[1])
		case "bodyLimit":
			v := strings.TrimSpace(kv[1])
			if v == "" {
				v = "0"
			}
			data.Codes[key][k] = v
		case "bodyLimitPkg":
			if v := strings.TrimSpace(kv[1]); len(v) > 2 {
				data.Imports[v] = strings.TrimSpace(kv[1])
			}
		case "resp":
			if v := strings.TrimSpace(kv[1]); v == `"object"` {
				data.Codes[key][k] = "true"
			}
		case "validation":
			if v := strings.TrimSpace(kv[1]); v == `"token"` {
				data.Codes[key][k] = "net_fw.Validation_type_token"
			}
		}
	}
}

// @Handler(type="valid.limit", pkg="", func="")
func parseHandlerValid_limit(words []string, data *templateData, key string) {
	if data.Codes[key] == nil {
		data.Codes[key] = make(map[string]string)
		data.CodesList = append(data.CodesList, key)
	}
	for _, ll := range words {
		kv := strings.Split(ll, "=")
		switch k := strings.TrimSpace(kv[0]); k {
		case "pkg":
			if v := strings.TrimSpace(kv[1]); len(v) > 2 {
				data.Imports[v] = strings.TrimSpace(kv[1])
			}
		case "func":
			if f := strings.ReplaceAll(strings.TrimSpace(kv[1]), "\"", ""); f != "" {
				data.Codes[key]["valid.limit"] = strings.ReplaceAll(strings.TrimSpace(kv[1]), "\"", "")
			}
		}
	}
}

// @Handler(type="valid.file", pkg="", func="", headsize=n)
func parseHandlerValid_file(words []string, data *templateData, key string) {
	if data.Codes[key] == nil {
		data.Codes[key] = make(map[string]string)
		data.CodesList = append(data.CodesList, key)
	}
	funcstr, headsize := "", ""
	for _, ll := range words {
		kv := strings.Split(ll, "=")
		switch k := strings.TrimSpace(kv[0]); k {
		case "pkg":
			if v := strings.TrimSpace(kv[1]); len(v) > 2 {
				data.Imports[v] = strings.TrimSpace(kv[1])
			}
		case "func":
			funcstr = strings.ReplaceAll(strings.TrimSpace(kv[1]), "\"", "")
		case "headsize":
			headsize = strings.TrimSpace(kv[1])
		}
	}
	if funcstr != "" {
		data.Imports[`"github.com/valyala/fasthttp/zzz/mime/multipart"`] = `"github.com/valyala/fasthttp/zzz/mime/multipart"`
		data.Codes[key]["valid.file"] = fmt.Sprintf("&multipart.MyValidHeader{ValidFormFileFormat: %s, ValidHeadSize: %s}", funcstr, headsize)
	}
}

func GetImports(o templateData) string {
	var str []string
	for _, imp := range o.Imports {
		str = append(str, imp)
	}
	return strings.Join(str, "\n")
}

func GetCodes(o templateData) string {
	var strs []string
	for _, cl := range o.CodesList {
		mm := o.Codes[cl]
		str := "net_fw.NewHtNetHandler("
		for _, order := range apiOrders {
			c := mm[order[0]]
			if c == "" {
				c = order[1]
			}
			str += c + ","
		}
		strs = append(strs, str[:len(str)-1]+")") // 去掉最后一个逗号
	}
	return strings.Join(strs, "\n")
}

var customTemplateFuncs = template.FuncMap{
	"GetImports": GetImports,
	"GetCodes":   GetCodes,
}

var apiOrders = [][]string{
	{"net", ""},
	{"path", ""},
	{"resp", "true"},
	{"validation", "net_fw.Validation_type_none"},
	{"bodyLimit", "0"},
	{"api", ""},
	{"valid.limit", "nil"},
	{"valid.file", "nil"},
}
