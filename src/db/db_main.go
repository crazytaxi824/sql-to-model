package db

import (
	"database/sql"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

var db *bun.DB

type Config struct {
	Addr     *string
	User     *string
	Password *string
	Name     *string
}

func openDB(conf Config) {
	// 连接数据库
	pgconn := pgdriver.NewConnector(
		pgdriver.WithAddr(*conf.Addr),
		pgdriver.WithInsecure(true),
		pgdriver.WithUser(*conf.User),
		pgdriver.WithPassword(*conf.Password),
		pgdriver.WithDatabase(*conf.Name),
		pgdriver.WithTimeout(5*time.Second),
	)

	// openDB()
	sqldb := sql.OpenDB(pgconn)
	db = bun.NewDB(sqldb, pgdialect.New())

	// DEBUG: 打印sql 语句
	// db.AddQueryHook(&queryHook{})
}

func FindsAllTable(conf Config, query QueryOpts) ([]Table, error) {
	openDB(conf)

	// 获取所有 table
	tables, err := getAllTable(query)
	if err != nil {
		return nil, err
	}

	return tables, nil
}

// 查询数据库内的所有表
func getAllTable(query QueryOpts) ([]Table, error) {
	resp, err := getAllSchemaTableColumnInfo(query)
	if err != nil {
		return nil, err
	}

	// 初始化 TableInfo
	var ti tableObj
	ti.tables = make(map[int64]Table)

	// 添加数据
	for i := range resp {
		ti.addTableInfo(resp[i])
	}

	return ti.SortedOutput(), nil
}
