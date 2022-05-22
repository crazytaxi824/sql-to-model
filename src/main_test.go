package main

import (
	"flag"
	"fmt"
	"local/src/db"
	"log"
	"strings"
	"testing"
)

func TestStructField(t *testing.T) {
	t.Log(structFieldName("table_name"))
	t.Log(structFieldName("_table_name"))
	t.Log(structFieldName("_table_name_"))
	t.Log(structFieldName("_a_b_"))
	t.Log(structFieldName("TableName"))
	t.Log(structFieldName("table_id"))
	t.Log(structFieldName("table_Id"))
	t.Log(structFieldName("id"))
}

func TestMajor(_ *testing.T) {
	log.SetFlags(log.Lshortfile)

	// setup flags
	var dbconf db.DBConfig
	dbconf.Addr = flag.String("a", "192.168.0.193:15432", "database Addr")
	dbconf.User = flag.String("u", "postgres", "database username")
	dbconf.Password = flag.String("p", "123456", "database password")
	dbconf.Name = flag.String("n", "test", "database name")
	pkg := flag.String("k", "model", "go package name")
	flag.Parse()

	tables, err := db.FindsAllTable(dbconf)
	if err != nil {
		log.Println(err)
		return
	}

	// for table info to go struct format
	r := genStructContent(dbconf, tables, *pkg)

	// print MODEL struct
	fmt.Println(strings.Join(r, "\n"))
}
