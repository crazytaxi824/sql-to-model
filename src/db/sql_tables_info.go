package db

import (
	"context"
	"strings"

	"github.com/uptrace/bun"
)

type queryResp struct {
	SchemaID      int64  `bun:"column:schema_id"`   // schema ID
	SchemaName    string `bun:"column:schema_name"` // schema name
	TableID       int64  `bun:"column:table_id"`    // table ID
	TableName     string `bun:"column:table_name"`  // table name
	TableNote     string `bun:"column:table_note"`  // table comments
	TableKind     string `bun:"column:table_kind"`  // r-table; v-view
	ColumnName    string `bun:"column:column_name"` // column name
	ColumnNote    string `bun:"column:column_note"` // column comments
	ColumnType    string `bun:"column:column_type"` // column data type
	ColumnNotNull bool   `bun:"column:not_null"`    // 是否允许 null
	ColumnDims    int    `bun:"column:dims"`        // array 类型维度
}

type QueryConf struct {
	// schema name list, seperate with ',' e.g.: "foo,bar",
	// omitempty - all schemas
	Schemas *string

	// table name list, seperate with ',' e.g.: "foo,bar",
	// omitempty - all tables
	// if 'Tables' is specified, then 'TableKind' will be ignored.
	Tables *string

	// 'r','t','table'-table only;
	// 'v','view'-view only,
	// omitempty,other - tables and views,
	TableKind *string
}

func getAllSchemaTableColumnInfo(queryConf QueryConf) ([]queryResp, error) {
	schemas := getSchemas(*queryConf.Schemas)
	tables := getTables(schemas, *queryConf.Tables, *queryConf.TableKind)
	return getColumnsInfo(tables)
}

// schema subQuery
func getSchemas(schemaName string) *bun.SelectQuery {
	schemas := db.NewSelect().Table("pg_namespace").
		Column("oid", "nspname")

	if schemaName == "" {
		// list all schema in database
		schemas = schemas.Where("nspname !~ ?", "^pg_").
			Where("nspname <> ?", "information_schema")
	} else {
		// specify schema name
		sl := strings.Split(schemaName, ",")
		schemas = schemas.Where("nspname IN (?)", bun.In(sl))
	}

	return schemas
}

// table subQuery
func getTables(schemas *bun.SelectQuery, tableName, tableKind string) *bun.SelectQuery {
	tables := db.NewSelect().TableExpr("? AS c", bun.Ident("pg_class")).
		ColumnExpr("? AS schema_id", bun.Ident("s.oid")).
		ColumnExpr("? AS schema_name", bun.Ident("s.nspname")).
		ColumnExpr("? AS table_id", bun.Ident("c.oid")).
		ColumnExpr("? AS table_name", bun.Ident("c.relname")).
		ColumnExpr("? as table_kind", bun.Ident("c.relkind")).
		ColumnExpr("? AS table_note", bun.Safe("obj_description(c.oid)")).
		Join("JOIN (?) AS s", schemas).
		JoinOn("c.relnamespace = s.oid").
		Where("c.relnamespace IN (s.oid)") // table in schema list

	if tableName != "" {
		// specify table name
		ts := strings.Split(tableName, ",")
		tables = tables.Where("c.relname IN (?)", bun.In(ts)).
			Where("c.relkind IN ('r', 'v')")

	} else {
		switch tableKind {
		case "r", "t", "table":
			tables = tables.Where("c.relkind = 'r'")

		case "v", "view":
			tables = tables.Where("c.relkind = 'v'")

		default:
			tables = tables.Where("c.relkind IN ('r', 'v')")
		}
	}

	return tables
}

func getColumnsInfo(tables *bun.SelectQuery) ([]queryResp, error) {
	// 根据 table/view oid 查询所有 columns attributes
	// SELECT * FROM pg_attribute WHERE attrelid IN (table_ids...) AND attnum>0 AND format_type(atttypid, atttypmod) <> '-';
	var resp []queryResp
	err := db.NewSelect().TableExpr("? as a", bun.Ident("pg_attribute")).
		Column("b.*").
		ColumnExpr("? AS column_name", bun.Ident("a.attname")).
		ColumnExpr("? AS not_null", bun.Ident("a.attnotnull")).
		ColumnExpr("? AS dims", bun.Ident("a.attndims")).
		ColumnExpr("? AS column_note", bun.Safe("col_description(a.attrelid,a.attnum)")).
		ColumnExpr("? AS column_type", bun.Safe("format_type(a.atttypid,a.atttypmod)")).
		Join("JOIN (?) AS b", tables).
		JoinOn("a.attrelid = b.table_id").
		Where("a.attrelid IN (b.table_id) AND a.attnum>0 AND format_type(a.atttypid, a.atttypmod) <> '-'").
		Order("b.table_kind ASC").
		Order("b.schema_id ASC").
		Order("b.table_id ASC").
		Order("a.attnum ASC").
		Scan(context.Background(), &resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
