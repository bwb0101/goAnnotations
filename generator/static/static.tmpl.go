package static

const handlersTemplate = `// 由注解自动生成: 不要手动编辑
/*
API注册
@Handler(type="api", net="http/tcp", path="/reg", bodyLimit=n, bodyLimitPkg="", resp="object", validation="token")
net = "http/tcp": 根据net类型分类处理
path = "/xxx": 请求路径
bodyLimit = n：当前请求体限制(k) 0 = 默认服务器配置；
bodyLimitPkg = "xxx/xxx": 包路径(限制用的是变量)
resp = "object"：返回的对象需要进行序列化；
validation = "token" / ""：不为空需要验证(目前只支持token)；空或不写：忽略验证

访问限制
@Handler(type="valid.limit", pkg="", func="")
pkg = 包名: xxx/xxx
func = 方法名: xxx.xxx

上传文件时验证文件头是否合法
@Handler(type="valid.file", pkg="", func="", headsize=n)
pkg = 包名: xxx/xxx
func = 方法名: xxx.xxx
headsize = 验证文件头大小: body[:headsize]
*/

package {{.PackageName}}

import (
	{{GetImports .}}
)

func init() {
	{{GetCodes .}}
}
`
