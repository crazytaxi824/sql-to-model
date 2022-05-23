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
	var dbconf db.Config
	dbconf.Addr = flag.String("a", "172.16.238.128:15432", "database Addr")
	// dbconf.Addr = flag.String("a", "192.168.0.193:15432", "database Addr")
	dbconf.User = flag.String("u", "postgres", "database username")
	dbconf.Password = flag.String("p", "123456", "database password")
	dbconf.Name = flag.String("n", "test", "database name")
	pkg := flag.String("g", "model", "go package name")

	// query config
	var queryConf db.QueryOpts
	queryConf.Schemas = flag.String("s", "foo,view", "specify schema list. eg:'foo,bar', omitempty - all schemas")
	queryConf.Tables = flag.String("t", "", "specify table list. eg:'foo,bar', omitempty - all tables")
	queryConf.TableKind = flag.String("k", "", "specify table or view, 't','r','table'-table; 'v','view'-view; others,omitempty-tables and views")
	flag.Parse()

	tables, err := db.FindsAllTable(dbconf, queryConf)
	if err != nil {
		log.Println(err)
		return
	}

	// for table info to go struct format
	r := genStructContent(dbconf, tables, *pkg)

	// print MODEL struct
	fmt.Println(strings.Join(r, "\n"))
}
