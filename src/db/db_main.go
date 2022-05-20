package db

import (
	"database/sql"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

var db *bun.DB

type DBConfig struct {
	Addr     *string
	User     *string
	Password *string
	Name     *string
}

func openDB(conf DBConfig) {
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

func FindsAllTable(conf DBConfig) ([]Table, error) {
	openDB(conf)

	// 获取所有 table
	tables, err := getAllTable()
	if err != nil {
		return nil, err
	}

	return tables, nil
}
