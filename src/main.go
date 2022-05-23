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
	dbconf.Addr = flag.String("a", "localhost:5432", "database Addr\n")
	dbconf.User = flag.String("u", "postgres", "database username\n")
	dbconf.Password = flag.String("p", "", "database password\n")
	dbconf.Name = flag.String("n", "test", "database name\n")
	pkg := flag.String("g", "model", "go package name\n")

	// query config
	var queryConf db.QueryConf
	queryConf.Schemas = flag.String("s", "", "specify schema list. eg:'foo,bar', omitempty - all schemas\n")
	queryConf.Tables = flag.String("t", "", "specify table list. eg:'foo,bar', omitempty - all tables\n")
	queryConf.TableKind = flag.String("k", "", "specify table or view\n't', 'r', 'table' - table only;\n'v', 'view' - view only;\nomitempty, others - tables and views\n")
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

// 生成 model 结构体
func genStructContent(conf db.DBConfig, tables []db.Table, pkg string) []string {
	var content = []string{
		fmt.Sprintf("// all tables from database: \"%s\"", *conf.Name),
		"package " + pkg,
		"",
		"import (", "\t\"github.com/uptrace/bun\"", ")\n",
	}

	for _, table := range tables {
		if table.Note != "" {
			content = append(content, "// "+table.Note) // table comments
		}

		content = append(content,
			fmt.Sprintf("type %s struct {", structFieldName(table.Name))) // table name
		content = append(content,
			fmt.Sprintf("\tbun.BaseModel `bun:\"table:%s.%s\"`", table.Schema, table.Name)) // table name tag

		for _, col := range table.Columns {
			gt := db.SqlTypeToGoType(col)

			var structField string

			// 是否需要手动确定 struct field Data Type
			if gt.Manual {
				structField = fmt.Sprintf("\t// %s %s `bun:\"column:%s%s\"`", structFieldName(col.Name), gt.Type, col.Name, gt.Tag)
			} else {
				structField = fmt.Sprintf("\t%s %s `bun:\"column:%s%s\"`", structFieldName(col.Name), gt.Type, col.Name, gt.Tag)
			}

			// 如果 sql column 有 comments
			if col.Note != "" {
				structField += fmt.Sprintf(" // %s", col.Note)
			}

			content = append(content, structField)
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
