package db

import (
	"strings"
)

type GoType struct {
	Tag, Type string
	Manual    bool
}

// map pgsql data type to golang data type
// TODO
// case "uuid":
//   1. install extentsion `sudo apt install postgresql-contrib-14`
//   2. then install the extension in each database you are going to use it, `CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`
//   gt.Type += "uuid.UUID" // import "github.com/google/uuid"
// case "character(n)":
//   需要获取 atttypmod(attribute Type modifier) 信息.
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

	case "timestamptz", "timestamp":
		if col.NotNull || col.Dims > 0 {
			gt.Type += "time.Time"
		} else {
			gt.Type = "*time.Time"
		}

	case "bytea":
		gt.Type += "[]byte"

	case "inet":
		gt.Type += "net.IP" //  也是 []byte 类型

	case "json", "jsonb":
		// json array 和 native array 的储存方式是不同的.
		// json array 储存的就是 json 格式 [1,2,3], 可以通过 pgsql 的 json 操作符操作单个元素.
		// native array 储存的是 {x,x,x} 格式, 如果要修改单个元素需要全部读出来, 然后修改后再写入.
		gt.Type = "<map[string]interface{}|[]slice|json.RawMessage>"

		// 需要手动确定 Data Type
		gt.Manual = true

	default:
		gt.Type = "<please declare Type manually>"
		gt.Manual = true // 需要手动确定 Data Type
	}

	return gt
}
