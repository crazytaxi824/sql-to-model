package db

// Column 字段名和字段注释
type Column struct {
	Name     string // column_name
	DataType string // bigint, text, jsonb ...
	Dims     int    // array 的情况下 dims>0; 不是 array 的情况下 dims=0
	Note     string // column comments
	NotNull  bool   // attnotnull
}

// Table Strcut 表名和表注释
type Table struct {
	Schema  string
	Name    string // table_name
	Note    string // table comments
	Columns []Column
}

type tableObj struct {
	order  []int64
	tables map[int64]Table
}

func (t *tableObj) addTableInfo(r queryResp) {
	// table exists
	table, ok := t.tables[r.TableID]
	if ok {
		table.Columns = append(table.Columns, Column{
			Name:     r.ColumnName,
			DataType: r.ColumnType,
			Dims:     r.ColumnDims,
			Note:     r.ColumnNote,
			NotNull:  r.ColumnNotNull,
		})
		t.tables[r.TableID] = table
		return
	}

	// table not exist
	t.order = append(t.order, r.TableID)
	t.tables[r.TableID] = Table{
		Schema: r.SchemaName,
		Name:   r.TableName,
		Note:   r.TableNote,
		Columns: []Column{
			{
				Name:     r.ColumnName,
				DataType: r.ColumnType,
				Dims:     r.ColumnDims,
				Note:     r.ColumnNote,
				NotNull:  r.ColumnNotNull,
			},
		},
	}
}

func (t *tableObj) SortedOutput() []Table {
	var tables []Table
	for _, id := range t.order {
		tables = append(tables, t.tables[id])
	}

	return tables
}
