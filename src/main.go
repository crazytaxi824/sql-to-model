package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"local/src/util"
	"log"
	"strings"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

var db *bun.DB

// Column 字段名和字段注释
type Column struct {
	Name     string `bun:"column:name"`
	DataType string `bun:"column:type"`
	Dims     int    `bun:"column:dims"` // array 的情况下 dims>0; 不是 array 的情况下 dims=0
	Note     string `bun:"column:note"`
	NotNull  bool   `bun:"column:notnull"`
}

// Table Strcut 表名和表注释
type Table struct {
	Schema  string   `bun:"column:schema"`
	Name    string   `bun:"column:name"`
	Note    string   `bun:"column:note"`
	Columns []Column `bun:"-"`
}

func main() {
	log.SetFlags(log.Lshortfile)

	dbAddr := flag.String("a", "localhost:5432", "database Addr")
	dbUser := flag.String("u", "postgres", "database username")
	dbPwd := flag.String("p", "", "database password")
	dbName := flag.String("n", "test", "database name")
	dbSchema := flag.String("s", "public", "database schema")
	flag.Parse()

	// 连接数据库
	pgconn := pgdriver.NewConnector(
		pgdriver.WithAddr(*dbAddr),
		pgdriver.WithInsecure(true),
		pgdriver.WithUser(*dbUser),
		pgdriver.WithPassword(*dbPwd),
		pgdriver.WithDatabase(*dbName),
		pgdriver.WithTimeout(5*time.Second),
	)

	// openDB()
	sqldb := sql.OpenDB(pgconn)
	db = bun.NewDB(sqldb, pgdialect.New())

	// DEBUG: 打印sql 语句
	// db.AddQueryHook(&util.QueryHook{})

	// 获取所有 table
	tables, err := getAllTableNames(*dbSchema)
	if err != nil {
		log.Println(err.Error())
		return
	}

	// 获取每一个 table 的所有 column
	for k := range tables {
		tables[k].Columns, err = getTableModel(tables[k].Name)
		if err != nil {
			log.Println(err)
			return
		}
	}

	// 生成内容
	r := genStructContent(tables)
	fmt.Println(strings.Join(r, "\n"))
}

// 生成 model 结构体
func genStructContent(tables []Table) []string {
	var content []string
	for _, table := range tables {
		if table.Note != "" {
			content = append(content, "// "+table.Note) // table note
		}
		content = append(content, fmt.Sprintf("type %s struct {", table.Name))                                    // table name
		content = append(content, fmt.Sprintf("\tbun.BaseModel `bun:\"table:%s.%s\"`", table.Schema, table.Name)) // table name

		for _, col := range table.Columns {
			tag, typ := sqlTypeToGoType(col)
			if col.Note != "" {
				content = append(content, fmt.Sprintf("\t%s %s `bun:\"column:%s%s\"` // %s", util.StructFieldName(col.Name), typ, col.Name, tag, col.Note))
			} else {
				content = append(content, fmt.Sprintf("\t%s %s `bun:\"column:%s%s\"`", util.StructFieldName(col.Name), typ, col.Name, tag))
			}
		}

		content = append(content, "}\n") // struct end
	}

	return content
}

// 查询数据库内的所有表
func getAllTableNames(schema string) ([]Table, error) {
	rows, err := db.Query(`
		SELECT
			obj_description(a.oid) as note,
			a.relname as name,
			b.nspname as schema
		FROM pg_class a
		JOIN pg_namespace b
		ON b.oid = a.relnamespace
		WHERE a.relkind = 'r' AND b.nspname = ?;`, schema)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	var table []Table
	err = db.ScanRows(context.Background(), rows, &table)
	if err != nil {
		return nil, err
	}

	return table, nil
}

// 查询表中的所有字段和类型
func getTableModel(tableName string) ([]Column, error) {
	rows, err := db.Query(`
		SELECT
			col_description(a.attrelid,a.attnum) as note,
			format_type(a.atttypid,a.atttypmod) as type,
			a.attname as name,
			a.attnotnull as notnull,
			a.attndims as dims
		FROM pg_class as c
		JOIN pg_attribute as a
		ON a.attrelid = c.oid
		WHERE c.relname = ? and a.attnum>0 and format_type(a.atttypid,a.atttypmod) <> '-';`, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	var models []Column
	err = db.ScanRows(context.Background(), rows, &models)
	if err != nil {
		return nil, err
	}

	return models, nil
}

// map pgsql data type to golang data type
func sqlTypeToGoType(col Column) (string, string) {
	var finalType string
	var finalTag string

	// 判断是否是 array 类型, 获取 array 维度.
	// However, the current implementation ignores any supplied array size limits, i.e. 'text[3][2]' 等于 'text[][]'
	if col.Dims > 0 {
		col.DataType = strings.ReplaceAll(col.DataType, "[]", "")
		finalTag += ",type:" + col.DataType + ",array"
	} else {
		finalTag += ",type:" + col.DataType
	}

	switch col.DataType {
	case "bigint": // int8
		if col.NotNull || col.Dims > 0 {
			finalType = "int64"
		} else {
			finalType = "*int64"
		}
	case "integer": // int/int4
		if col.NotNull || col.Dims > 0 {
			finalType = "int"
		} else {
			finalType = "*int"
		}
	case "smallint": // int2
		if col.NotNull || col.Dims > 0 {
			finalType = "int16"
		} else {
			finalType = "*int16"
		}
	case "decimal", "numeric", "double precision":
		if col.NotNull || col.Dims > 0 {
			finalType = "float64"
		} else {
			finalType = "*float64"
		}
	case "real":
		if col.NotNull || col.Dims > 0 {
			finalType = "float32"
		} else {
			finalType = "*float32"
		}
	case "text":
		if col.NotNull || col.Dims > 0 {
			finalType = "string"
		} else {
			finalType = "*string"
		}
	case "json", "jsonb":
		// json array 和 native array 的储存方式是不同的.
		// json array 储存的就是 json 格式 [1,2,3], 可以通过 pgsql 的 json 操作符操作单个元素.
		// native array 储存的是 {x,x,x} 格式, 如果要修改单个元素需要全部读出来, 然后修改后再写入.
		finalType = "-- map[string]interface{}|[]slice|json.RawMessage --"
	case "boolean":
		if col.NotNull || col.Dims > 0 {
			finalType = "bool"
		} else {
			finalType = "*bool"
		}
	case "bytea":
		finalType = "[]byte"
	case "inet":
		finalType = "net.IP" // []byte
	default:
		finalType = "-- 请手动绑定数据类型 --"
	}

	// 如果是 array, 则需要使用 []type 类型, 同时在 tag 中添加 ",array"
	if col.Dims > 0 {
		finalType = strings.Repeat("[]", col.Dims) + finalType
		return finalTag, finalType
	}

	return finalTag, finalType
}
