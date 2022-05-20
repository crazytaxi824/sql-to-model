package main

import "testing"

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
