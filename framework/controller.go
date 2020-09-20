package framework

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/eudore/eudore"
	"strings"
)

type (
	ControllerWebsite struct {
		Context
	}
)

func NewControllerWebsite(db *sql.DB) *ControllerWebsite {
	return &ControllerWebsite{
		Context: Context{
			DB: db,
		},
	}
}

// Init 实现控制器初始方法。
func (ctl *ControllerWebsite) Init(ctx eudore.Context) error {
	ctl.Context.ContextData.Context = ctx
	return nil
}

// Release 实现控制器释放方法。
func (ctl *ControllerWebsite) Release(eudore.Context) error {
	return nil
}

// Inject 方法实现控制器注入到路由器的方法,调用ControllerInjectStateful方法注入。
func (ctl *ControllerWebsite) Inject(controller eudore.Controller, router eudore.Router) error {
	params := router.Params()
	params.Set("resource-prefix", params.Get("route"))
	return eudore.ControllerInjectStateful(controller, router)
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
	return ctl.QueryCountContext(ctl.Context.GetContext(), query, args...)
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
	var size = ctl.GetQueryInt("size", 20)
	var page = ctl.GetQueryInt("page") * size
	pagesql := fmt.Sprintf(" limit %d OFFSET %d", size, page)
	return ctl.QueryRows(query+pagesql, args...)
}

// QueryContextPages 方法对查询结构分页，使用url参数size和page作为分页参数，对查询sql后附加sql: " limit %d OFFSET %d"。
func (ctl *ControllerWebsite) QueryPagesContext(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
	var size = ctl.GetQueryInt("size", 20)
	var page = ctl.GetQueryInt("page") * size
	pagesql := fmt.Sprintf(" limit %d OFFSET %d", size, page)
	var datas []map[string]interface{}
	err := ctl.QueryBindContext(ctx, &datas, query+pagesql, args...)
	return datas, err
}

// Exec executes a query without returning any rows.
// The args are for any placeholder parameters in the query.
func (ctl *ControllerWebsite) Exec(query string, args ...interface{}) (sql.Result, error) {
	return ctl.ExecContext(ctl.Context.GetContext(), query, args...)
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
