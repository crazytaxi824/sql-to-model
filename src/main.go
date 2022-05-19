package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"local/src/db"
)

func main() {
	log.SetFlags(log.Lshortfile)

	// setup flags
	var dbconf db.DBConfig
	dbconf.Addr = flag.String("a", "192.168.0.193:15432", "database Addr")
	// dbconf.Addr = flag.String("a", "localhost:5432", "database Addr")
	dbconf.User = flag.String("u", "postgres", "database username")
	dbconf.Password = flag.String("p", "123456", "database password")
	// dbconf.Password = flag.String("p", "", "database password")
	dbconf.Name = flag.String("n", "test", "database name")
	// dbconf.Name = flag.String("n", "", "database name")
	dbconf.Schema = flag.String("s", "public", "database schema")
	flag.Parse()

	tables, err := db.FindsAllTable(dbconf)
	if err != nil {
		log.Println(err)
		return
	}

	// for table info to go struct format
	r := genStructContent(tables)

	// print MODEL struct
	fmt.Println(strings.Join(r, "\n"))
}

// 生成 model 结构体
func genStructContent(tables []db.Table) []string {
	var content []string
	for _, table := range tables {
		if table.Note != "" {
			content = append(content, "// "+table.Note) // table note
		}
		content = append(content, fmt.Sprintf("type %s struct {", table.Name))                                    // table name
		content = append(content, fmt.Sprintf("\tbun.BaseModel `bun:\"table:%s.%s\"`", table.Schema, table.Name)) // table name

		for _, col := range table.Columns {
			tag, typ := db.SqlTypeToGoType(col)
			if col.Note != "" {
				content = append(content, fmt.Sprintf("\t%s %s `bun:\"column:%s%s\"` // %s", structFieldName(col.Name), typ, col.Name, tag, col.Note))
			} else {
				content = append(content, fmt.Sprintf("\t%s %s `bun:\"column:%s%s\"`", structFieldName(col.Name), typ, col.Name, tag))
			}
		}

		content = append(content, "}\n") // struct end
	}

	return content
}

// snake_case to CamelCase
func structFieldName(src string) string {
	sa := strings.Split(src, "_")
	for i := range sa {
		if len(sa[i]) > 0 {
			if strings.ToUpper(sa[i]) == "ID" {
				sa[i] = "ID"
			} else {
				sa[i] = strings.ToUpper(sa[i][0:1]) + sa[i][1:]
			}
		}
	}

	return strings.Join(sa, "")
}
