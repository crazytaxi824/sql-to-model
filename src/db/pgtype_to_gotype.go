package db

import (
	"strings"
)

type GoType struct {
	Tag, Type string
	Manual    bool
}

// map pgsql data type to golang data type
func SqlTypeToGoType(col Column) GoType {
	var gt GoType

	// 判断是否是 array 类型, 获取 array 维度.
	// However, the current implementation ignores any supplied array size limits, i.e. 'text[3][2]' 等于 'text[][]'
	if col.Dims > 0 {
		col.DataType = strings.ReplaceAll(col.DataType, "[]", "")
		gt.Tag = ",type:" + col.DataType + ",array"
		gt.Type = strings.Repeat("[]", col.Dims)
	} else {
		gt.Tag = ",type:" + col.DataType
	}

	switch col.DataType {
	case "bigint": // int8
		if col.NotNull || col.Dims > 0 {
			gt.Type += "int64"
		} else {
			gt.Type = "*int64"
		}
	case "integer": // int/int4
		if col.NotNull || col.Dims > 0 {
			gt.Type += "int"
		} else {
			gt.Type = "*int"
		}
	case "smallint": // int2
		if col.NotNull || col.Dims > 0 {
			gt.Type += "int16"
		} else {
			gt.Type = "*int16"
		}
	case "decimal", "numeric", "double precision":
		if col.NotNull || col.Dims > 0 {
			gt.Type += "float64"
		} else {
			gt.Type = "*float64"
		}
	case "real":
		if col.NotNull || col.Dims > 0 {
			gt.Type += "float32"
		} else {
			gt.Type = "*float32"
		}
	case "text":
		if col.NotNull || col.Dims > 0 {
			gt.Type += "string"
		} else {
			gt.Type = "*string"
		}
	case "boolean":
		if col.NotNull || col.Dims > 0 {
			gt.Type += "bool"
		} else {
			gt.Type = "*bool"
		}
	case "bytea":
		gt.Type += "[]byte"
	case "inet":
		gt.Type += "net.IP" //  也是 []byte 类型
	case "json", "jsonb":
		// json array 和 native array 的储存方式是不同的.
		// json array 储存的就是 json 格式 [1,2,3], 可以通过 pgsql 的 json 操作符操作单个元素.
		// native array 储存的是 {x,x,x} 格式, 如果要修改单个元素需要全部读出来, 然后修改后再写入.
		gt.Type = "-- map[string]interface{}|[]slice|json.RawMessage --"
		gt.Manual = true
	default:
		gt.Type = "-- 请手动绑定数据类型 --"
		gt.Manual = true
	}

	return gt
}
