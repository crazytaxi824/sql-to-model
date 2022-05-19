package util

import "testing"

func TestStructField(t *testing.T) {
	t.Log(StructFieldName("table_name"))
	t.Log(StructFieldName("_table_name"))
	t.Log(StructFieldName("_table_name_"))
	t.Log(StructFieldName("_a_b_"))
	t.Log(StructFieldName("TableName"))
	t.Log(StructFieldName("table_id"))
	t.Log(StructFieldName("table_Id"))
	t.Log(StructFieldName("id"))
}
