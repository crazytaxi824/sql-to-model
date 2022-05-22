package db

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

func TestSubQuery(t *testing.T) {
	// 连接数据库
	pgconn := pgdriver.NewConnector(
		pgdriver.WithAddr("192.168.0.193:15432"),
		pgdriver.WithInsecure(true),
		pgdriver.WithUser("postgres"),
		pgdriver.WithPassword("123456"),
		pgdriver.WithDatabase("test"),
		pgdriver.WithTimeout(5*time.Second),
	)

	// openDB()
	sqldb := sql.OpenDB(pgconn)
	db = bun.NewDB(sqldb, pgdialect.New())

	// DEBUG: 打印sql 语句
	db.AddQueryHook(&queryHook{})

	r, err := getAllSchemaTableColumnInfo()
	if err != nil {
		t.Error(err)
		return
	}

	for i := range r {
		fmt.Printf("%+v\n", r[i])
	}
}
