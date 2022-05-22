## sql-to-model

#### 扫描postgresql数据库，创建go模型文件。

查询方式:

```sql
-- 查询所有 table and view
SELECT
	obj_description(a.oid) as note,  -- table comments
	a.relname as name,               -- table/view name
	b.nspname as schema,             -- schema
	a.relkind as kind                -- r=table | v=view
FROM pg_class a
JOIN pg_namespace b
ON b.oid = a.relnamespace
WHERE a.relkind IN ('r', 'v') AND b.nspname !~ '^pg_' AND nspname <> 'information_schema';

-- 查询所有 table/view 的字段属性 attributes
SELECT
	c.relname as table_name,
	col_description(a.attrelid,a.attnum) as note,  -- column comments
	format_type(a.atttypid,a.atttypmod) as type,   -- data type
	a.attname as name,
	a.attnotnull as notnull,
	a.attndims as dims   -- array 维度. 如果是 array dims > 0; 如果不是 array, dims = 0
FROM pg_class as c       -- pg_class  table/view 属性
JOIN pg_attribute as a   -- pg_attribute column 属性
ON a.attrelid = c.oid
WHERE c.relname IN (?) and a.attnum>0 AND format_type(a.atttypid,a.atttypmod) <> '-';
```

其他查询方式: information_schema, 这是系统自动生成的视图. 该方式没有 id 显示, 没办法显示 table comments. 不推荐.

```sql
SELECT *
FROM information_schema.views  -- view only
WHERE table_schema NOT IN ('information_schema', 'pg_catalog');

SELECT *
FROM information_schema.schemata
WHERE schema_name NOT IN ('information_schema', 'pg_catalog', 'pg_toast');  -- schema

SELECT *
FROM information_schema."tables"
WHERE table_schema NOT IN ('information_schema', 'pg_catalog');  -- table and view
```

#### 使用方式：

```bash
$ stm -a 127.0.0.1 -p 5432 -pwd xxx -db postgres -o ~/Desktop/go_model.go
```


#### 参数设置：
```bash
  -a string
    	database Addr (default "127.0.0.1")
      
  -c	convert ID int64 type to string —— bool DEFAULT false
  
  -db string
    	database name, default - empty string
      
  -j	true, no omitempty —— bool DEFAULT false
  
  -o string
    	gen model file from database (default "./Desktop/db_model.go")
      
  -p string
    	database Addr port (default "5432")
      
  -pwd string
    	database password, default - empty string
      
  -u string
    	database username (default "postgres")
```

