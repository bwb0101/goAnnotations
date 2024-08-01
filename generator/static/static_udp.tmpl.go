package static

const udpHandlersTemplate = `// 由注解自动生成: 不要手动编辑
/*
API注册
@Handler(type="api", net="udp", msgId="uint16", dataPtrStruct="path|pkg.struct", validation="user")
net = "http/tcp/udp": 根据net类型分类处理
msgId = "uint16"：uint16数值
dataPtrStruct = "path|pkg.struct"：path = import的路径；pkg.struct = 反序列化时的包名结构体
validation = "user"：检测UserValue是否存在
*/

package {{.PackageName}}

import (
	{{GetImportsUdp .}}
)

func init() {
	{{GetCodesUdp .}}
}
`
