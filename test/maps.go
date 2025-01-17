/*
 * 项目名称：framework
 * 文件名：maps.go
 * 日期：2023/12/14 20:30
 * 作者：Ben
 */

package test

type (
	// mongodb反射需要指针进行接口转换，要保存到db的，变量要用New
	Maps[K comparable, V any] struct {
		M map[K]V // 直接使用变量，注意是否需要调用SyncOp
	}
)

// mongodb反射需要指针进行接口转换，要保存到db的，变量要用New
func NewMaps[K comparable, V any](cap int) *Maps[K, V] {
	m := Maps[K, V]{}
	return m.New(cap)
}

func (m *Maps[K, V]) SyncOp(sync bool) {
}

// mongodb反射需要指针进行接口转换，要保存到db的，变量要用New
func (*Maps[K, V]) New(cap int) (m *Maps[K, V]) {
	m = new(Maps[K, V])
	if cap == 0 {
		m.M = map[K]V{}
	} else {
		m.M = make(map[K]V, cap)
	}
	return
}
