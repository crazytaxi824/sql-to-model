package main

import (
	"context"
	"flag"
	"io/ioutil"
	"log"
	"strings"

	"github.com/go-pg/pg/v9"
)

var (
	db          *pg.DB
	fileContent string
)

// Model 字段名和字段注释
type Model struct {
	ColumnName string `pg:"name"`
	DataType   string `pg:"type"`
	Note       string `pg:"note"`
}

// Table Strcut 表名和表注释
type TableStrcut struct {
	TabName string `pg:"tabname"`
	Note    string `pg:"note"`
}

var convert *bool
var tagJSON *bool

func main() {
	log.SetFlags(log.Lshortfile)

	outputFilePath := flag.String("o", "./Desktop/db_model.go", "gen model file from database")
	dbAddr := flag.String("a", "127.0.0.1", "database Addr")
	dbPort := flag.String("p", "5432", "database Addr port")
	dbUser := flag.String("u", "postgres", "database username")
	dbPwd := flag.String("pwd", "", "database password, default - empty string")
	dbDB := flag.String("db", "", "database name, default - empty string")
	convert = flag.Bool("c", false, "convert ID int64 type to string —— bool DEFAULT false")
	tagJSON = flag.Bool("j", false, "true, no omitempty —— bool DEFAULT false")
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
		getTableModel(table)
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
func getTableModel(table TableStrcut) {
	// 查询所有表的所有结构
	var models []Model

	_, err := db.Query(&models, `SELECT col_description(a.attrelid,a.attnum) as note,format_type(a.atttypid,a.atttypmod) as type,a.attname as name FROM pg_class as c,pg_attribute as a where c.relname = '`+table.TabName+`' and a.attrelid = c.oid and a.attnum>0 and format_type(a.atttypid,a.atttypmod) <> '-'`)
	if err != nil {
		log.Println(err.Error())
		return
	}

	// 添加到文件内容中
	genFileContent(models, table)
}

func genFileContent(models []Model, table TableStrcut) {
	fileContent = fileContent + "// " + underLineToCamel(table.TabName) + " " + table.Note + "\r\n"
	fileContent = fileContent + "type " + underLineToCamel(table.TabName) + " struct{\r\n"
	fileContent = fileContent + "tableName struct{} `pg:\"" + table.TabName + "\"` \r\n"
	for _, model := range models {
		l := len(model.ColumnName)
		if *convert && l > 1 {
			if model.ColumnName[l-2:l] == "id" && model.DataType == "bigint" {
				fileContent = fileContent + underLineToCamel(model.ColumnName) + " string `pg:\"" + model.ColumnName + "\" json:\"" + underLineToJSONCamel(model.ColumnName)
				if *tagJSON {
					fileContent = fileContent + "\"` " + "// " + model.Note + "\r\n"
				} else {
					fileContent = fileContent + ",omitempty\"` " + "// " + model.Note + "\r\n"
				}
			} else {
				suffix, dataType := sqlTypeToGoType(model.DataType)
				fileContent = fileContent + underLineToCamel(model.ColumnName) + " " + dataType + " `pg:\"" + model.ColumnName + suffix + "\" json:\"" + underLineToJSONCamel(model.ColumnName)
				if *tagJSON {
					fileContent = fileContent + "\"` " + "// " + model.Note + "\r\n"
				} else {
					fileContent = fileContent + ",omitempty\"` " + "// " + model.Note + "\r\n"
				}
			}
		} else {
			suffix, dataType := sqlTypeToGoType(model.DataType)
			fileContent = fileContent + underLineToCamel(model.ColumnName) + " " + dataType + " `pg:\"" + model.ColumnName + suffix + "\" json:\"" + underLineToJSONCamel(model.ColumnName)
			if *tagJSON {
				fileContent = fileContent + "\"` " + "// " + model.Note + "\r\n"
			} else {
				fileContent = fileContent + ",omitempty\"` " + "// " + model.Note + "\r\n"
			}
		}
	}
	fileContent += "}\r\n\r\n"
}

func sqlTypeToGoType(dataType string) (string, string) {
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
		// return ",json", finalType
	case "json":
		finalType = "map[string]interface{}"
		// return ",json", finalType
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
			prefix += "[]"
		}
		finalType = prefix + finalType
		return ",array", finalType
	}
	return "", finalType
}

func underLineToCamel(underLineStr string) string {
	defer func() {
		if r := recover(); r != nil {
			log.Println(r)
		}
	}()
	var CamelName string

	ulStr := strings.TrimSpace(underLineStr)
	// 判断id后面是否有值
	length := len(ulStr)
	if length >= 2 {
		if ulStr[length-2:length] == "id" {
			ulStr = ulStr[:length-2] + "ID"
		}
	}

	if length >= 3 {
		if ulStr[length-3:length] == "url" {
			ulStr = ulStr[:length-3] + "URL"
		}
	}

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
		CamelName = ulSlice[0]
		for i := 1; i < length; i++ {
			CamelName = CamelName + strings.ToUpper(string(ulSlice[i][0])) + ulSlice[i][1:]
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

type hook struct{}

func (hook) BeforeQuery(ctx context.Context, qe *pg.QueryEvent) (context.Context, error) {
	return ctx, nil
}

func (hook) AfterQuery(ctx context.Context, qe *pg.QueryEvent) error {
	log.Println(qe.FormattedQuery())
	return nil
}
