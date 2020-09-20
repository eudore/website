package framework

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/eudore/eudore"
)

type Context struct {
	eudore.ContextData
	*sql.DB
}

func NewContextExtend(db *sql.DB) []interface{} {
	return []interface{}{
		NewContextExtendContext(db),
		NewContextExtendContextError(db),
		NewContextExtendContextInterface(db),
		NewContextExtendContextInterfaceError(db),
		NewContextExtendContextSqlResultError(db),
		NewContextExtendContextMapStringError(db),
		NewContextExtendContextMapStringInterfaceError(db),
	}
}

func NewContextExtendContext(db *sql.DB) interface{} {
	return func(fn func(Context)) eudore.HandlerFunc {
		return func(ctx eudore.Context) {
			fn(Context{eudore.ContextData{Context: ctx}, db})
		}
	}
}

func NewContextExtendContextError(db *sql.DB) interface{} {
	return func(fn func(Context) error) eudore.HandlerFunc {
		return func(ctx eudore.Context) {
			err := fn(Context{eudore.ContextData{Context: ctx}, db})
			if err != nil {
				ctx.Fatal(err)
			}
		}
	}
}

func NewContextExtendContextInterface(db *sql.DB) interface{} {
	return func(fn func(Context) interface{}) eudore.HandlerFunc {
		return func(ctx eudore.Context) {
			data := fn(Context{eudore.ContextData{Context: ctx}, db})
			if ctx.Response().Size() == 0 {
				err := ctx.Render(data)
				if err != nil {
					ctx.Fatal(err)
				}
			}
		}
	}
}

func NewContextExtendContextInterfaceError(db *sql.DB) interface{} {
	return func(fn func(Context) (interface{}, error)) eudore.HandlerFunc {
		return func(ctx eudore.Context) {
			data, err := fn(Context{eudore.ContextData{Context: ctx}, db})
			if err == nil && ctx.Response().Size() == 0 {
				err = ctx.Render(data)
			}
			if err != nil {
				ctx.Fatal(err)
			}
		}
	}
}

func NewContextExtendContextSqlResultError(db *sql.DB) interface{} {
	return func(fn func(Context) (sql.Result, error)) eudore.HandlerFunc {
		return func(ctx eudore.Context) {
			result, err := fn(Context{eudore.ContextData{Context: ctx}, db})
			if err == nil && ctx.Response().Size() == 0 {
				rows, _ := result.RowsAffected()
				err = ctx.Render(map[string]int64{
					"rowsaffected": rows,
				})
			}
			if err != nil {
				ctx.Fatal(err)
			}
		}
	}
}

func NewContextExtendContextMapStringError(db *sql.DB) interface{} {
	return func(fn func(Context, map[string]interface{}) error) eudore.HandlerFunc {
		return func(ctx eudore.Context) {
			data := make(map[string]interface{})
			err := ctx.Bind(&data)
			if err != nil {
				ctx.Fatal(err)
				return
			}
			err = fn(Context{eudore.ContextData{Context: ctx}, db}, data)
			if err != nil {
				ctx.Fatal(err)
			}
		}
	}
}

func NewContextExtendContextMapStringInterfaceError(db *sql.DB) interface{} {
	return func(fn func(Context, map[string]interface{}) (interface{}, error)) eudore.HandlerFunc {
		return func(ctx eudore.Context) {
			data := make(map[string]interface{})
			err := ctx.Bind(&data)
			if err != nil {
				ctx.Fatal(err)
				return
			}
			result, err := fn(Context{eudore.ContextData{Context: ctx}, db}, data)
			if err != nil {
				ctx.Fatal(err)
				return
			}
			if ctx.Response().Size() == 0 {
				err = ctx.Render(result)
				if err != nil {
					ctx.Fatal(err)
				}
			}
		}
	}
}

func (ctx Context) Debugf(format string, args ...interface{}) {
	if format == "json" {
		body, err := json.Marshal(args)
		ctx.Debug(string(body), err)
		return
	}
	ctx.ContextData.Debugf(format, args...)
}

func (ctx Context) QueryMap(query string, args ...interface{}) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	err := ctx.QueryBindContext(ctx.GetContext(), &data, query, args...)
	return data, err
}

func (ctx Context) QueryRows(query string, args ...interface{}) ([]map[string]interface{}, error) {
	var datas []map[string]interface{}
	err := ctx.QueryBindContext(ctx.GetContext(), &datas, query, args...)
	return datas, err
}

func (ctx Context) QueryBind(dest interface{}, query string, args ...interface{}) error {
	return ctx.QueryBindContext(ctx.GetContext(), dest, query, args...)
}

func (ctx Context) QueryBindContext(cctx context.Context, dest interface{}, query string, args ...interface{}) error {
	ctx.SetParam("sql", fmt.Sprintf("%s %v", query, args))
	iValue := NewBindValue(dest)
	if !iValue.IsBind() {
		return fmt.Errorf("bind input type is invlide '%s'", iValue.Type())
	}
	rows, err := ctx.QueryContext(cctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return err
	}
	length := len(columns)
	value := make([]interface{}, length)
	columnPointers := make([]interface{}, length)
	for i := 0; i < length; i++ {
		columnPointers[i] = &value[i]
	}
	if rows.Next() {
		rows.Scan(columnPointers...)
		iValue.Bind(columns, value)
	}
	if iValue.IsSlice() {
		for rows.Next() {
			rows.Scan(columnPointers...)
			iValue.Bind(columns, value)
		}
	}
	return rows.Err()
}

// QueryCount
func (ctx *Context) QueryCount(query string, args ...interface{}) int {
	return ctx.QueryCountContext(ctx.GetContext(), query, args...)
}

// QueryCountContext
func (ctx *Context) QueryCountContext(cctx context.Context, query string, args ...interface{}) int {
	var count int
	err := ctx.QueryRowContext(cctx, query, args...).Scan(&count)
	if err != nil {
		ctx.Error(err)
		return 0
	}
	return count
}
