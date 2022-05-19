package db

import (
	"context"
)

// Column 字段名和字段注释
type Column struct {
	Name     string `bun:"column:name"`    // table_column_name
	DataType string `bun:"column:type"`    // bigint, text, jsonb ...
	Dims     int    `bun:"column:dims"`    // array 的情况下 dims>0; 不是 array 的情况下 dims=0
	Note     string `bun:"column:note"`    // comments
	NotNull  bool   `bun:"column:notnull"` // attnotnull
}

// Table Strcut 表名和表注释
type Table struct {
	Schema  string   `bun:"column:schema"`
	Name    string   `bun:"column:name"`
	Note    string   `bun:"column:note"`
	Columns []Column `bun:"-"`
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
