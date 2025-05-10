/*
 * 项目名称：goAnnotations
 * 文件名：ranking.go
 * 日期：2024/12/18 17:05
 * 作者：Ben
 */

package test

type (
	Ranking struct {
		T  string
		Id int32 `pk:""`
		_h int32
	}
	Ranking2 struct {
		T string
		N string
	}
)
