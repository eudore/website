package tool

import (
	"database/sql"
	"fmt"
)

func ScanRows(stmt *sql.Stmt, args ...interface{}) ([]map[string]interface{}, error) {
	rows, err := stmt.Query(args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	length := len(columns)
	value := make([]interface{}, length)
	columnPointers := make([]interface{}, length)
	for i := 0; i < length; i++ {
		columnPointers[i] = &value[i]
	}

	var datas []map[string]interface{}
	for rows.Next() {
		rows.Scan(columnPointers...)
		data := make(map[string]interface{}, length)
		for i := 0; i < length; i++ {
			data[columns[i]] = value[i]
		}
		datas = append(datas, data)
	}
	return datas, rows.Err()
}

var sqls = map[string]string{}

func SaveSql(key, val string) {
	sqls[key] = val
}

func SaveSqls(data map[string]string) {
	for key, val := range data {
		sqls[key] = val
	}
}

func InitStmt(db *sql.DB, data map[**sql.Stmt]string) error {
	for key, val := range data {
		if v2, ok := sqls[val]; ok {
			val = v2
		}
		stmt, err := db.Prepare(val)
		if err != nil {
			return fmt.Errorf("init stmt sql '%s',error: %v", val, err)
		}
		*key = stmt
	}
	return nil
}
