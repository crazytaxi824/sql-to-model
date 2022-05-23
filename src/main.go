package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"local/src/db"
)

type flagConfig struct {
	db    db.Config
	query db.QueryOpts

	// other flags
	gopkg *string
}

func setFlags() flagConfig {
	var fc flagConfig

	// database config flags
	fc.db.Addr = flag.String("addr", "localhost:5432", "database Addr")
	fc.db.User = flag.String("user", "postgres", "database username")
	fc.db.Password = flag.String("password", "", "database password")
	fc.db.Name = flag.String("database", "test", "database name")

	// query flags
	fc.query.Schemas = flag.String("schema", "", "specify schema list. eg:'foo,bar'\nomitempty - all schemas")
	fc.query.Tables = flag.String("table", "", "specify table list. eg:'foo,bar'\nomitempty - all tables")
	fc.query.TableKind = flag.String("kind", "", "specify table or view.\n't', 'r', 'table' - table only;\n'v', 'view' - view only;\nomitempty, others - tables and views")

	// other flags
	fc.gopkg = flag.String("gopkg", "model", "go package name")

	// alias
	addr := flag.Lookup("addr")
	flag.Var(addr.Value, "a", fmt.Sprintf("alias to '-%s'", addr.Name))

	user := flag.Lookup("user")
	flag.Var(user.Value, "u", fmt.Sprintf("alias to '-%s'", user.Name))

	password := flag.Lookup("password")
	flag.Var(password.Value, "p", fmt.Sprintf("alias to '-%s'", password.Name))

	database := flag.Lookup("database")
	flag.Var(database.Value, "db", fmt.Sprintf("alias to '-%s'", database.Name))

	schema := flag.Lookup("schema")
	flag.Var(schema.Value, "s", fmt.Sprintf("alias to '-%s'", schema.Name))

	table := flag.Lookup("table")
	flag.Var(table.Value, "t", fmt.Sprintf("alias to '-%s'", table.Name))

	kind := flag.Lookup("kind")
	flag.Var(kind.Value, "k", fmt.Sprintf("alias to '-%s'", kind.Name))

	gopkg := flag.Lookup("gopkg")
	flag.Var(gopkg.Value, "g", fmt.Sprintf("alias to '-%s'", gopkg.Name))

	flag.Parse()

	return fc
}

func main() {
	log.SetFlags(log.Lshortfile)

	fc := setFlags()

	tables, err := db.FindsAllTable(fc.db, fc.query)
	if err != nil {
		log.Println(err)
		return
	}

	// for table info to go struct format
	r := genStructContent(fc.db, tables, *fc.gopkg)

	// print MODEL struct
	fmt.Println(strings.Join(r, "\n"))
}

// 生成 model 结构体
func genStructContent(conf db.Config, tables []db.Table, pkg string) []string {
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
