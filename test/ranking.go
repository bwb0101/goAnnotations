/*
 * 项目名称：goAnnotations
 * 文件名：ranking.go
 * 日期：2024/12/18 17:05
 * 作者：Ben
 */

package test

type (
	Ranking struct {
		Id      int32 `pk:""`
		Ranking *Maps[int32, int32]
	}
)
