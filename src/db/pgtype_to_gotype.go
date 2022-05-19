package db

import (
	"strings"
)

// map pgsql data type to golang data type
func SqlTypeToGoType(col Column) (string, string) {
	var finalType string
	var finalTag string

	// 判断是否是 array 类型, 获取 array 维度.
	// However, the current implementation ignores any supplied array size limits, i.e. 'text[3][2]' 等于 'text[][]'
	if col.Dims > 0 {
		col.DataType = strings.ReplaceAll(col.DataType, "[]", "")
		finalTag = ",type:" + col.DataType + ",array"
		finalType = strings.Repeat("[]", col.Dims)
	} else {
		finalTag = ",type:" + col.DataType
	}

	switch col.DataType {
	case "bigint": // int8
		if col.NotNull || col.Dims > 0 {
			finalType += "int64"
		} else {
			finalType = "*int64"
		}
	case "integer": // int/int4
		if col.NotNull || col.Dims > 0 {
			finalType += "int"
		} else {
			finalType = "*int"
		}
	case "smallint": // int2
		if col.NotNull || col.Dims > 0 {
			finalType += "int16"
		} else {
			finalType = "*int16"
		}
	case "decimal", "numeric", "double precision":
		if col.NotNull || col.Dims > 0 {
			finalType += "float64"
		} else {
			finalType = "*float64"
		}
	case "real":
		if col.NotNull || col.Dims > 0 {
			finalType += "float32"
		} else {
			finalType = "*float32"
		}
	case "text":
		if col.NotNull || col.Dims > 0 {
			finalType += "string"
		} else {
			finalType = "*string"
		}
	case "boolean":
		if col.NotNull || col.Dims > 0 {
			finalType += "bool"
		} else {
			finalType = "*bool"
		}
	case "bytea":
		finalType += "[]byte"
	case "inet":
		finalType += "net.IP" //  也是 []byte 类型
	case "json", "jsonb":
		// json array 和 native array 的储存方式是不同的.
		// json array 储存的就是 json 格式 [1,2,3], 可以通过 pgsql 的 json 操作符操作单个元素.
		// native array 储存的是 {x,x,x} 格式, 如果要修改单个元素需要全部读出来, 然后修改后再写入.
		finalType = "-- map[string]interface{}|[]slice|json.RawMessage --"
	default:
		finalType = "-- 请手动绑定数据类型 --"
	}

	return finalTag, finalType
}
