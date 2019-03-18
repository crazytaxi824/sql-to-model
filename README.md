## sql-to-model

#### 扫描postgresql数据库，创建go模型文件。



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
