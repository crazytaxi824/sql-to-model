## sql-to-model

#### 扫描postgresql数据库，创建go模型文件。

查询方式:

```sql
-- 查询所有 user schema
SELECT * FROM pg_namespace WHERE nspname !~ '^pg_' AND nspname <> 'information_schema';

-- 根据 schema oid 查询所有 table and view.
-- IN (2200, 16410, 16418) 都是上面 sql 查出来的 schema oid.
-- relkind = 'r' 表示是 table; 'v' 表示是 view.
SELECT * FROM pg_class WHERE relnamespace IN (2200, 16410, 16418) AND relkind IN ('r', 'v');

-- 根据 table/view oid 查询所有 columns attributes.
-- IN (16385, 16435, 16393...) 都是上面 sql 查询出来的 table oid.
SELECT *
FROM pg_attribute 
WHERE attrelid IN (16385, 16435, 16393...) AND attnum>0 AND format_type(atttypid, atttypmod) <> '-';
```

合并所有查询: 使用 subquery

```sql
-- 这里是查询所有 column attributes 的 query.
SELECT
	b.*,
	a.attname as column_name,
	col_description(a.attrelid,a.attnum) as column_note,
	format_type(a.atttypid,a.atttypmod) as column_type,
	a.attnum as column_num,       -- column 在 table 中的顺序, 主要用于排序.
	a.attnotnull as not_null,     -- column 是否可以为 null
	a.attndims as dims            -- array 维度. column type 不是 array 时, dims = 0; 是 array 时 dims > 0.
FROM pg_attribute as a
JOIN (
	-- 这里是查询所有 table and view 的 subquery.
	SELECT 
		s.oid as schema_id,
		s.nspname as schema_name,
		c.oid as table_id,
		c.relname as table_name,
		obj_description(c.oid) as table_note,
		c.relkind as table_or_view    -- 'r' = table; 'v' = view
	FROM pg_class as c
	JOIN (
			-- 这里是查询所有 schema 的 subquery.
			SELECT * 
			FROM pg_namespace 
			WHERE nspname !~ '^pg_' AND nspname <> 'information_schema'
		) as s
	ON c.relnamespace = s.oid
	WHERE c.relnamespace IN (s.oid) AND c.relkind IN ('r', 'v')
) as b
ON a.attrelid = b.table_id
WHERE attrelid IN (b.table_id) AND attnum>0 AND format_type(atttypid, atttypmod) <> '-'
ORDER BY (b.table_id, a.attnum) ASC;
```

<br />

其他查询方式: information_schema, 这是系统自动生成的视图集合(schema).

information_schema 内有很多自动生成的视图(view), 包括 tables, views, shcemata, columns, attributes... 等 view.

该方式没有 id 显示, 没办法显示 table comments. 不推荐.

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

<br />

## TODO

- array_agg()

- jsonb_agg()
