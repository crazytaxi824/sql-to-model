package db

import (
	"context"

	"github.com/uptrace/bun"
)

// Column 字段名和字段注释
type Column struct {
	TableName string `bun:"column:table_name"` // table_name, 用于绑定到 Table 结构体
	Name      string `bun:"column:name"`       // table_column_name
	DataType  string `bun:"column:type"`       // bigint, text, jsonb ...
	Dims      int    `bun:"column:dims"`       // array 的情况下 dims>0; 不是 array 的情况下 dims=0
	Note      string `bun:"column:note"`       // comments
	NotNull   bool   `bun:"column:notnull"`    // attnotnull
}

// Table Strcut 表名和表注释
type Table struct {
	Schema  string   `bun:"column:schema"`
	Name    string   `bun:"column:name"`
	Note    string   `bun:"column:note"`
	Columns []Column `bun:"-"`
}

// 查询数据库内的所有表
func getAllTable() ([]Table, error) {
	tables, err := getAllTableInfo()
	if err != nil {
		return nil, err
	}

	err = getTableColumnsInfo(tables)
	if err != nil {
		return nil, err
	}

	return tables, nil
}

func getAllTableInfo() ([]Table, error) {
	// `select * from pg_namespace;` all schema
	rows, err := db.Query(`
		SELECT
			obj_description(a.oid) as note,
			a.relname as name,
			b.nspname as schema
		FROM pg_class a
		JOIN pg_namespace b
		ON b.oid = a.relnamespace
		WHERE a.relkind IN ('r', 'v') AND b.nspname !~ '^pg_' AND nspname <> 'information_schema';`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	var tables []Table
	err = db.ScanRows(context.Background(), rows, &tables)
	if err != nil {
		return nil, err
	}

	return tables, nil
}

// NOTE: 这里是在传入的 []Table 上添加 column 信息, 所以 []Table 本身不能进行 append 和 slice 等操作.
func getTableColumnsInfo(tables []Table) error {
	var tableNames []string
	for _, table := range tables {
		tableNames = append(tableNames, table.Name)
	}

	// 这里使用 IN 查询, 避免 N+1 问题.
	rows, err := db.Query(`
		SELECT
			c.relname as table_name,
			col_description(a.attrelid,a.attnum) as note,
			format_type(a.atttypid,a.atttypmod) as type,
			a.attname as name,
			a.attnotnull as notnull,
			a.attndims as dims
		FROM pg_class as c
		JOIN pg_attribute as a
		ON a.attrelid = c.oid
		WHERE c.relname IN (?) and a.attnum>0 AND format_type(a.atttypid,a.atttypmod) <> '-';`, bun.In(tableNames))
	if err != nil {
		return err
	}
	defer rows.Close()

	err = rows.Err()
	if err != nil {
		return err
	}

	var cols []Column
	err = db.ScanRows(context.Background(), rows, &cols)
	if err != nil {
		return err
	}

	// 处理 Table 结构体和 cols 的关系
	for i := range tables {
		for j := range cols {
			if tables[i].Name == cols[j].TableName {
				tables[i].Columns = append(tables[i].Columns, cols[j])
			}
		}
	}

	return nil
}
