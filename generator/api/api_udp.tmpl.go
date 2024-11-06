package api

const udpHandlersTemplate = `// 由注解自动生成: 不要手动编辑
/*
API注册
@Handler(type="api", net="udp", msgId="uint16", dataPtrStruct="path|pkg.struct", validation="user", bodyType="0/1")
net = "http/tcp/udp": 根据net类型分类处理
msgId = "uint16"：uint16数值
dataPtrStruct = "path|pkg.struct"：path = import的路径；pkg.struct = 反序列化时的包名结构体
validation = "user"：检测KcpConnect->UserValue是否为nil
bodyType = "0/1"：默认0，1 framebody类型
*/

package {{.PackageName}}

import (
	{{GetImportsUdp .}}
)

func init() {
	{{GetCodesUdp .}}
}
`
