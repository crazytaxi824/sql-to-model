package main

import (
	"flag"
	"io/ioutil"
	"log"
	"strings"

	"github.com/go-ini/ini"
	"github.com/go-pg/pg"
)

var (
	db          *pg.DB
	cfg         *ini.File
	fileContent string
)

// Model 字段名和字段注释
type Model struct {
	ColumnName string `sql:"name"`
	DataType   string `sql:"type"`
	Note       string `sql:"note"`
}

// TableStrcut 表名和表注释
type TableStrcut struct {
	TabName string `sql:"tabname"`
	Note    string `sql:"note"`
}

func main() {
	log.SetFlags(log.Lshortfile)

	outputFilePath := flag.String("o", "./Desktop/raymodel.go", "gen model file from database")
	dbAddr := flag.String("a", "127.0.0.1", "database Addr")
	dbPort := flag.String("p", "5432", "database Addr port")
	dbUser := flag.String("u", "postgres", "database username")
	dbPwd := flag.String("pwd", "", "database password, default - empty string")
	dbDB := flag.String("db", "", "database name, default - empty string")
	convert := flag.Bool("c", false, "convert ID int64 type to string —— bool DEFAULT false")
	flag.Parse()

	// 连接数据库
	// openDB()
	db = pg.Connect(&pg.Options{
		Addr:     *dbAddr + ":" + *dbPort,
		User:     *dbUser,
		Password: *dbPwd,
		Database: *dbDB,
	})
	// 打印sql 语句
	// db.AddQueryHook(hook{})

	// 写package
	fileContent = "package \r\n\r\n"

	tables, err := getAllTableNames()
	if err != nil {
		log.Println(err.Error())
		return
	}

	for _, table := range tables {
		getTableModel(table, *convert)
	}

	// 写文件
	writeFile(*outputFilePath)
}

// 查询数据库内的所有表
func getAllTableNames() ([]TableStrcut, error) {
	var table []TableStrcut
	_, err := db.Query(&table, "SELECT obj_description(oid) as note, relname as tabname FROM pg_class WHERE relkind = 'r' AND relname NOT LIKE 'pg_%' AND relname NOT LIKE 'sql_%' ORDER BY relname;")
	if err != nil {
		return nil, err
	}

	return table, nil
}

// 查询表中的所有字段和类型
func getTableModel(table TableStrcut, convert bool) {
	// 查询所有表的所有结构
	var models []Model

	_, err := db.Query(&models, `SELECT col_description(a.attrelid,a.attnum) as note,format_type(a.atttypid,a.atttypmod) as type,a.attname as name FROM pg_class as c,pg_attribute as a where c.relname = '`+table.TabName+`' and a.attrelid = c.oid and a.attnum>0 and format_type(a.atttypid,a.atttypmod) <> '-'`)
	if err != nil {
		log.Println(err.Error())
		return
	}

	// 添加到文件内容中
	genFileContent(models, table, convert)
}

func genFileContent(models []Model, table TableStrcut, convert bool) {
	fileContent = fileContent + "// " + underLineToCamel(table.TabName) + " " + table.Note + "\r\n"
	fileContent = fileContent + "type " + underLineToCamel(table.TabName) + " struct{\r\n"
	fileContent = fileContent + "tableName struct{} `sql:\"" + table.TabName + "\"` \r\n"
	for _, model := range models {
		if model.DataType != "jsonb" {
			l := len(model.ColumnName)
			if convert && l > 1 {
				if model.ColumnName[l-2:l] == "id" {
					fileContent = fileContent + underLineToCamel(model.ColumnName) + " string `sql:\"" + model.ColumnName + "\" json:\"" + underLineToJSONCamel(model.ColumnName) + ",omitempty\"` " + "//" + model.Note + "\r\n"
				} else {
					fileContent = fileContent + underLineToCamel(model.ColumnName) + " " + sqlTypeToGoType(model.DataType) + " `sql:\"" + model.ColumnName + "\" json:\"" + underLineToJSONCamel(model.ColumnName) + ",omitempty\"` " + "//" + model.Note + "\r\n"
				}
			} else {
				fileContent = fileContent + underLineToCamel(model.ColumnName) + " " + sqlTypeToGoType(model.DataType) + " `sql:\"" + model.ColumnName + "\" json:\"" + underLineToJSONCamel(model.ColumnName) + ",omitempty\"` " + "//" + model.Note + "\r\n"
			}
		} else {
			fileContent = fileContent + underLineToCamel(model.ColumnName) + " " + sqlTypeToGoType(model.DataType) + " `pg:\"" + model.ColumnName + ",json\" json:\"" + underLineToJSONCamel(model.ColumnName) + ",omitempty\"` " + "//" + model.Note + "\r\n"
		}

	}
	fileContent = fileContent + "}\r\n\r\n"
}

func sqlTypeToGoType(dataType string) string {
	var finalType string
	n := strings.Count(dataType, "[]")
	if n > 0 {
		dataType = strings.Replace(dataType, "[]", "", -1)
	}
	switch dataType {
	case "bigint":
		finalType = "int64"
	case "integer":
		finalType = "int"
	case "smallint":
		finalType = "int"
	case "decimal":
		finalType = "float64"
	case "numeric":
		finalType = "float64"
	case "double precision":
		finalType = "float64"
	case "real":
		finalType = "float32"
	case "text":
		finalType = "string"
	case "jsonb":
		finalType = "map[string]interface{}"
	case "json":
		finalType = "map[string]interface{}"
	case "boolean":
		finalType = "bool"
	case "timestamptz":
		finalType = "time.Time"
	case "bytea":
		finalType = "[]byte"
	case "inet":
		finalType = "net.IP"
	case "cidr":
		finalType = "net.IPNet"
	default:
		finalType = "-- 请手动绑定数据类型 --"
	}
	if n > 0 {
		var prefix string
		for i := 0; i < n; i++ {
			prefix = prefix + "[]"
		}
		finalType = prefix + finalType
	}
	return finalType
}

func underLineToCamel(underLineStr string) string {
	defer func() {
		if r := recover(); r != nil {
			log.Println(r)
		}
	}()
	var CamelName string

	ulStr := strings.TrimSpace(underLineStr)
	ulStr = strings.Replace(ulStr, "id", "ID", -1)
	ulSlice := strings.Split(ulStr, "_")
	for _, v := range ulSlice {
		if len(v) > 0 {
			CamelName = CamelName + strings.ToUpper(string(v[0])) + v[1:]
		}
	}

	return CamelName
}

func underLineToJSONCamel(underLineStr string) string {
	defer func() {
		if r := recover(); r != nil {
			log.Println(r)
		}
	}()
	var CamelName string

	ulStr := strings.TrimSpace(underLineStr)
	ulSlice := strings.Split(ulStr, "_")
	length := len(ulSlice)

	if length > 1 {
		for i := 1; i < length; i++ {
			CamelName = ulSlice[0] + strings.ToUpper(string(ulSlice[i][0])) + ulSlice[i][1:]
		}
	} else {
		CamelName = ulStr
	}

	return CamelName
}

// 写文件
func writeFile(outputFilePath string) {
	// goFile := fileContent
	err := ioutil.WriteFile(outputFilePath, []byte(fileContent), 0644)
	if err != nil {
		log.Println(err.Error())
		return
	}
	log.Println("写入完成")
}

func openDB() {
	sect := cfg.Section("database")
	db = pg.Connect(&pg.Options{
		Addr:     sect.Key("addr").String(),
		User:     sect.Key("user").MustString("app"),
		Password: sect.Key("password").String(),
		Database: sect.Key("database").MustString("game"),
	})

	db.AddQueryHook(hook{})
}

type hook struct{}

func (hook) BeforeQuery(qe *pg.QueryEvent) {}

func (hook) AfterQuery(qe *pg.QueryEvent) {
	log.Println(qe.FormattedQuery())
}

func loadConfig(filePath string) error {
	var err error
	cfg, err = ini.Load(filePath)
	if err != nil {
		return err
	}
	return nil
}
