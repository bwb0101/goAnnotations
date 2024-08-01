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

	"github.com/bwb0101/goAnnotations/generator"
	"github.com/bwb0101/goAnnotations/generator/util"
	"github.com/bwb0101/goAnnotations/model"
)

type templateData struct {
	PackageName string
	TargetDir   string
	// Services    []model.Operation
	httpImports   map[string]string
	httpCodes     map[string]map[string]string
	httpCodesList []string
	//
	tcpImports   map[string]string
	tcpCodes     map[string]map[string]string
	tcpCodesList []string
	//
	udpImports   map[string]string
	udpCodes     map[string]map[string]string
	udpCodesList []string
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
			data = &templateData{
				httpImports: map[string]string{`"framework/common/net_fw"`: `"framework/common/net_fw"`}, httpCodes: make(map[string]map[string]string),
				tcpImports: map[string]string{`"framework/common/net_fw"`: `"framework/common/net_fw"`}, tcpCodes: make(map[string]map[string]string),
				udpImports: map[string]string{`"framework/common/net_fw"`: `"framework/common/net_fw"`}, udpCodes: make(map[string]map[string]string),
			}
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
	if err := generate_http(datas, inputDir); err == nil {
		if err := generate_tcp(datas, inputDir); err == nil {
			return generate_udp(datas, inputDir)
		}
	}
	return nil
}

func NewGenerator() generator.Generator {
	return &Generator{}
}

func generate_http(datas map[string]*templateData, dir string) error {
	for _, data := range datas {
		if len(data.httpCodes) > 0 {
			if err := util.Generate(util.Info{
				Data:           *data,
				Src:            data.PackageName,
				TargetFilename: filepath.Join(dir, generator.GenfilePrefix+"http_api_handler.go"),
				TemplateName:   "static_http",
				TemplateString: httpHandlersTemplate,
				FuncMap:        customHttpTemplateFuncs,
			}); err != nil {
				return err
			}
		}
	}
	return nil
}

func generate_tcp(datas map[string]*templateData, dir string) error {
	for _, data := range datas {
		if len(data.tcpCodes) > 0 {
			if err := util.Generate(util.Info{
				Data:           *data,
				Src:            data.PackageName,
				TargetFilename: filepath.Join(dir, generator.GenfilePrefix+"tcp_api_handler.go"),
				TemplateName:   "static_tcp",
				TemplateString: tcpHandlersTemplate,
				FuncMap:        customTcpTemplateFuncs,
			}); err != nil {
				return err
			}
		}
	}
	return nil
}

func generate_udp(datas map[string]*templateData, dir string) error {
	for _, data := range datas {
		if len(data.udpCodes) > 0 {
			if err := util.Generate(util.Info{
				Data:           *data,
				Src:            data.PackageName,
				TargetFilename: filepath.Join(dir, generator.GenfilePrefix+"udp_api_handler.go"),
				TemplateName:   "static_udp",
				TemplateString: udpHandlersTemplate,
				FuncMap:        customUdpTemplateFuncs,
			}); err != nil {
				return err
			}
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
	net := ""
	for _, ll := range words {
		kv := strings.Split(ll, "=")
		switch k := strings.TrimSpace(kv[0]); k {
		case "net":
			if net = strings.TrimSpace(kv[1]); net == `"http"` {
				if data.httpCodes[key] == nil {
					data.httpCodes[key] = map[string]string{"api": apiName, "api_method": apiName}
					data.httpCodesList = append(data.httpCodesList, key)
				}
				//
				data.httpCodes[key][k] = "net_fw.HTNET_type_http"
			} else if net == `"tcp"` {
				if data.tcpCodes[key] == nil {
					data.tcpCodes[key] = map[string]string{"api": apiName, "api_method": apiName}
					data.tcpCodesList = append(data.tcpCodesList, key)
				}
			} else if net == `"udp"` {
				if data.udpCodes[key] == nil {
					data.udpCodes[key] = map[string]string{"api": apiName, "api_method": apiName}
					data.udpCodesList = append(data.udpCodesList, key)
				}
			}
		case "path":
			data.httpCodes[key][k] = strings.TrimSpace(kv[1])
		case "msgId":
			if v := strings.TrimSpace(kv[1]); v != "" {
				v = v[1 : len(v)-1] // 去掉两边引号
				if net == `"tcp"` {
					data.tcpCodes[key][k] = v
				} else if net == `"udp"` {
					data.udpCodes[key][k] = v
				}
			}
		case "bodyLimit":
			// v := strings.TrimSpace(kv[1])
			// if v == "" {
			// 	v = "0"
			// }
			data.httpCodes[key][k] = strings.TrimSpace(kv[1])
		case "resp":
			if v := strings.TrimSpace(kv[1]); v == `"object"` {
				data.httpCodes[key][k] = "true"
			}
		case "validation":
			if v := strings.TrimSpace(kv[1]); v == `"token"` {
				data.httpCodes[key][k] = "net_fw.Validation_type_token"
			} else if v == `"user"` {
				data.httpCodes[key][k] = "true"
			}
		case "dataPtrStruct":
			if v := strings.TrimSpace(kv[1]); v != "" {
				v = v[1 : len(v)-1] // 去掉两边引号
				vv := strings.Split(v, "|")
				if net == `"http"` {
					data.httpImports[fmt.Sprintf(`"%s"`, vv[0])] = fmt.Sprintf(`"%s"`, vv[0])
					data.httpCodes[key][k] = fmt.Sprintf("func() any{ return &%s{}}", vv[1])
				} else if net == `"tcp"` {
					data.tcpImports[fmt.Sprintf(`"%s"`, vv[0])] = fmt.Sprintf(`"%s"`, vv[0])
					data.tcpCodes[key][k] = fmt.Sprintf("func() any{ return &%s{}}", vv[1])
				} else if net == `"udp"` {
					data.udpImports[fmt.Sprintf(`"%s"`, vv[0])] = fmt.Sprintf(`"%s"`, vv[0])
					data.udpCodes[key][k] = fmt.Sprintf("func() any{ return &%s{}}", vv[1])
				}
			}
		case "bodyType":
			data.httpCodes[key][k] = strings.TrimSpace(kv[1])
		}
	}
}

// @Handler(type="valid.limit", pkg="", func="")
func parseHandlerValid_limit(words []string, data *templateData, key string) {
	if data.httpCodes[key] == nil {
		data.httpCodes[key] = make(map[string]string)
		data.httpCodesList = append(data.httpCodesList, key)
	}
	for _, ll := range words {
		kv := strings.Split(ll, "=")
		switch k := strings.TrimSpace(kv[0]); k {
		case "pkg":
			if v := strings.TrimSpace(kv[1]); len(v) > 2 {
				data.httpImports[v] = strings.TrimSpace(kv[1])
			}
		case "func":
			if f := strings.ReplaceAll(strings.TrimSpace(kv[1]), "\"", ""); f != "" {
				data.httpCodes[key]["valid.limit"] = strings.ReplaceAll(strings.TrimSpace(kv[1]), "\"", "")
			}
		}
	}
}

// @Handler(type="valid.file", pkg="", func="", headsize=n)
func parseHandlerValid_file(words []string, data *templateData, key string) {
	if data.httpCodes[key] == nil {
		data.httpCodes[key] = make(map[string]string)
		data.httpCodesList = append(data.httpCodesList, key)
	}
	funcstr, headsize := "", ""
	for _, ll := range words {
		kv := strings.Split(ll, "=")
		switch k := strings.TrimSpace(kv[0]); k {
		case "pkg":
			if v := strings.TrimSpace(kv[1]); len(v) > 2 {
				data.httpImports[v] = strings.TrimSpace(kv[1])
			}
		case "func":
			funcstr = strings.ReplaceAll(strings.TrimSpace(kv[1]), "\"", "")
		case "headsize":
			headsize = strings.TrimSpace(kv[1])
		}
	}
	if funcstr != "" {
		data.httpImports[`"github.com/valyala/fasthttp/zzz/mime/multipart"`] = `"github.com/valyala/fasthttp/zzz/mime/multipart"`
		data.httpCodes[key]["valid.file"] = fmt.Sprintf("&multipart.MyValidHeader{ValidFormFileFormat: %s, ValidHeadSize: %s}", funcstr, headsize)
	}
}

func GetImportsHttp(o templateData) string {
	var str []string
	for _, imp := range o.httpImports {
		str = append(str, imp)
	}
	return strings.Join(str, "\n")
}

func GetImportsTcp(o templateData) string {
	var str []string
	for _, imp := range o.tcpImports {
		str = append(str, imp)
	}
	return strings.Join(str, "\n")
}

func GetImportsUdp(o templateData) string {
	var str []string
	for _, imp := range o.udpImports {
		str = append(str, imp)
	}
	return strings.Join(str, "\n")
}

func GetCodesHttp(o templateData) string {
	var strs []string
	for _, cl := range o.httpCodesList {
		mm := o.httpCodes[cl]
		str := "net_fw.NewHtNetHandler("
		for _, order := range httpApiOrders {
			c := mm[order[0]]
			if c == "" {
				c = order[1]
			} else {
				if len(order) > 2 {
					if order[2] == "str" {
						c = "\"" + c + "\""
					}
				}
			}
			str += c + ","
		}
		strs = append(strs, str[:len(str)-1]+")") // 去掉最后一个逗号
	}
	return strings.Join(strs, "\n")
}

func GetCodesTcp(o templateData) string {
	var strs []string
	for _, cl := range o.tcpCodesList {
		mm := o.tcpCodes[cl]
		str := "net_fw.NewTcpNetHandler("
		for _, order := range httpApiOrders {
			c := mm[order[0]]
			if c == "" {
				c = order[1]
			} else {
				if len(order) > 2 {
					if order[2] == "str" {
						c = "\"" + c + "\""
					}
				}
			}
			str += c + ","
		}
		strs = append(strs, str[:len(str)-1]+")") // 去掉最后一个逗号
	}
	return strings.Join(strs, "\n")
}

func GetCodesUdp(o templateData) string {
	var strs []string
	for _, cl := range o.udpCodesList {
		mm := o.udpCodes[cl]
		str := "net_fw.NewKcpNetHandler("
		for _, order := range httpApiOrders {
			c := mm[order[0]]
			if c == "" {
				c = order[1]
			} else {
				if len(order) > 2 {
					if order[2] == "str" {
						c = "\"" + c + "\""
					}
				}
			}
			str += c + ","
		}
		strs = append(strs, str[:len(str)-1]+")") // 去掉最后一个逗号
	}
	return strings.Join(strs, "\n")
}

var customHttpTemplateFuncs = template.FuncMap{
	"GetImportsHttp": GetImportsHttp,
	"GetCodesHttp":   GetCodesHttp,
}

var customTcpTemplateFuncs = template.FuncMap{
	"GetImportsTcp": GetImportsTcp,
	"GetCodesTcp":   GetCodesTcp,
}

var customUdpTemplateFuncs = template.FuncMap{
	"GetImportsUdp": GetImportsUdp,
	"GetCodesUdp":   GetCodesUdp,
}

var httpApiOrders = [][]string{
	{"net", ""},
	{"path", ""},
	{"resp", "false"},
	{"validation", "net_fw.Validation_type_none"},
	{"bodyLimit", "0"},
	{"api", "nil"},
	{"api_method", "", "str"},
	{"valid.limit", "nil"},
	{"valid.file", "nil"},
	{"dataPtrStruct", "nil"},
	{"bodyType", "0"},
}

// msgId uint16, call func(*KcpRequestInfo) []fw_udp.Frame, callMethod string, takePtrStruct func() proto.Message
var tcpUdpApiOrders = [][]string{
	{"msgId", "-1"},
	{"api", "nil"},
	{"api_method", "", "str"},
	{"dataPtrStruct", "nil"},
	{"validation", "false"},
}
