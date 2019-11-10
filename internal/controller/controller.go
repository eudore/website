package controller

import (
	"context"
	"database/sql"
	"strings"
	// "encoding/json"
	"fmt"
	"github.com/eudore/eudore"
)

type (
	ControllerWebsite struct {
		eudore.ContextData
		*sql.DB
	}
)

func NewControllerWejass(db *sql.DB) *ControllerWebsite {
	return &ControllerWebsite{
		DB: db,
	}
}

// Init 实现控制器初始方法。
func (ctl *ControllerWebsite) Init(ctx eudore.Context) error {
	ctl.Context = ctx
	return nil
}

// Release 实现控制器释放方法。
func (ctl *ControllerWebsite) Release() error {
	return nil
}

// Inject 方法实现控制器注入到路由器的方法,调用ControllerBaseInject方法注入。
func (ctl *ControllerWebsite) Inject(controller eudore.Controller, router eudore.RouterMethod) error {
	return eudore.ControllerBaseInject(controller, router.SetParam("prefix", router.GetParam("route")))
}

// ControllerRoute 方法返回默认路由信息。
func (ctl *ControllerWebsite) ControllerRoute() map[string]string {
	return nil
}

// GetRouteParam 方法添加路由参数信息。
func (ctl *ControllerWebsite) GetRouteParam(pkg, name, method string) string {
	pos := strings.LastIndexByte(pkg, '/') + 1
	if pos != 0 {
		pkg = pkg[pos:]
	}
	if strings.HasSuffix(name, "Controller") {
		name = name[:len(name)-len("Controller")]
	}
	if pkg == "task" {
		return ""
	}
	return fmt.Sprintf("action=%s:%s:%s", pkg, name, method)
}

// QueryCount
func (ctl *ControllerWebsite) QueryCount(query string, args ...interface{}) int {
	return ctl.QueryCountContext(ctl.Context.Context(), query, args...)
}

// QueryCountContext
func (ctl *ControllerWebsite) QueryCountContext(ctx context.Context, query string, args ...interface{}) int {
	var count int
	err := ctl.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		ctl.Error(err)
		return 0
	}
	return count
}

// QueryPages 方法对查询结构分页，使用url参数size和page作为分页参数，对查询sql后附加sql: " limit %d OFFSET %d"。
func (ctl *ControllerWebsite) QueryPages(query string, args ...interface{}) ([]map[string]interface{}, error) {
	var size = ctl.GetQueryDefaultInt("size", 20)
	var page = ctl.GetQueryInt("page") * size
	pagesql := fmt.Sprintf(" limit %d OFFSET %d", size, page)
	return ctl.QueryRows(query+pagesql, args...)
}

// QueryContextPages 方法对查询结构分页，使用url参数size和page作为分页参数，对查询sql后附加sql: " limit %d OFFSET %d"。
func (ctl *ControllerWebsite) QueryPagesContext(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
	var size = ctl.GetQueryDefaultInt("size", 20)
	var page = ctl.GetQueryInt("page") * size
	pagesql := fmt.Sprintf(" limit %d OFFSET %d", size, page)
	return ctl.QueryRowsContext(ctx, query+pagesql, args...)
}

// Exec executes a query without returning any rows.
// The args are for any placeholder parameters in the query.
func (ctl *ControllerWebsite) Exec(query string, args ...interface{}) (sql.Result, error) {
	return ctl.ExecContext(ctl.Context.Context(), query, args...)
}

func (ctl *ControllerWebsite) QueryRows(query string, args ...interface{}) ([]map[string]interface{}, error) {
	return ctl.QueryRowsContext(ctl.Context.Context(), query, args...)
}

func (ctl *ControllerWebsite) QueryRowsContext(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
	rows, err := ctl.Query(query, args...)
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

func (ctl *ControllerWebsite) QueryJSON(query string, args ...interface{}) (map[string]interface{}, error) {
	return ctl.QueryJSONContext(ctl.Context.Context(), query, args...)
}

func (ctl *ControllerWebsite) QueryJSONContext(ctx context.Context, query string, args ...interface{}) (map[string]interface{}, error) {
	rows, err := ctl.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return nil, sql.ErrNoRows
	}

	length := len(columns)
	value := make([]interface{}, length)
	columnPointers := make([]interface{}, length)
	for i := 0; i < length; i++ {
		columnPointers[i] = &value[i]
	}

	err = rows.Scan(columnPointers...)
	if err != nil {
		return nil, err
	}

	data := make(map[string]interface{}, length)
	for i := 0; i < length; i++ {
		data[columns[i]] = value[i]
	}
	return data, rows.Close()
}

func (ctl *ControllerWebsite) QueryBind(data interface{}, query string, args ...interface{}) error {
	return ctl.QueryBindContext(ctl.Context.Context(), data, query, args...)
}
func (ctl *ControllerWebsite) QueryBindContext(ctx context.Context, data interface{}, query string, args ...interface{}) error {
	datas, err := ctl.QueryJSONContext(ctx, query, args...)
	if err != nil {
		return err
	}
	for k, v := range datas {
		eudore.Set(data, k, v)
	}
	return nil
}

type pairs struct {
	keys []string
	vals []interface{}
}

func NewPairs(keys []string) *pairs {
	return &pairs{
		keys: keys,
		vals: make([]interface{}, len(keys)),
	}
}

func (p *pairs) set(key string, val interface{}) {
	for i := range p.keys {
		if p.keys[i] == key {
			p.vals[i] = val
		}
	}
}

func (ctl *ControllerWebsite) ExecBodyWithJSON(sql string, keys ...string) error {
	pairs := NewPairs(keys)
	data := getJsonKeys(ctl.Body())

	var n = len(keys) * 2
	for i := 0; i < len(data); i += 2 {
		pairs.set(string(data[i]), string(data[i+1]))
		if (i+2)%n == 0 {
			// 执行sql
			// TODO: 未使用批量执行sql
			_, err := ctl.Exec(sql, pairs.vals...)
			if err != nil {
				ctl.Error(err)
				return err
			}
		}
	}

	return nil
}

func getJsonKeys(body []byte) [][]byte {
	var i, pos int
	var keys [][]byte
	for ; i < len(body); i++ {
		switch body[i] {
		case '{':
			// start
			pos = i + 1
		case ':':
			// key
			keys = append(keys, body[pos+1:i-1])
			pos = i + 1
		case '}', ',':
			// val
			if pos < i {
				if body[pos] == '"' && body[i-1] == '"' {
					keys = append(keys, body[pos+1:i-1])

				} else {
					keys = append(keys, body[pos:i])

				}
			}
			pos = i + 1
		case '\\':
			i++

		}

	}
	return keys
}

func unescape(data []byte) []byte {
	body := make([]byte, 0, len(data))
	for i := 0; i < len(data); i++ {
		if data[i] == '\\' {
			i++
			switch data[i] {
			case '\\':
				body = append(body, '\\')
			case '/':
				body = append(body, '/')
			case 'b':
				body = append(body, '\b')
			case 'f':
				body = append(body, '\f')
			case 'n':
				body = append(body, '\n')
			case 'r':
				body = append(body, '\r')
			case 't':
				body = append(body, '\t')
			case '"':
				body = append(body, '"')
			}
		} else {
			body = append(body, data[i])
		}

	}
	return body
}
