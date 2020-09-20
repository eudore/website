package framework

import (
	"database/sql"
	"fmt"
	"github.com/eudore/eudore"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type (
	TableController struct {
		Table
		Pkg  string
		Name string
	}
	Table struct {
		TableName    string
		SelectFields string
		FieldNames   []string
		FieldTypes   []string
		FieldDefault []string
		FieldIsNull  []bool
		DataType     reflect.Type
	}
	ViewController struct {
		TableController
	}
	Condition struct {
		Field      string `json:"field"`
		Value      string `json:"value"`
		Expression string `json:"expression"`
	}
)

func NewTableController(pkg, name, tb string, object interface{}, db *sql.DB) *TableController {
	return &TableController{
		Table: newTable(db, object, tb),
		Pkg:   pkg,
		Name:  name,
	}
}

// Init 实现控制器初始方法。
func (ctl *TableController) Init(ctx eudore.Context) error {
	return nil
}

// Release 实现控制器释放方法。
func (ctl *TableController) Release(eudore.Context) error {
	return nil
}

// Inject 方法实现控制器注入到路由器的方法,调用ControllerInjectStateful方法注入。
func (ctl *TableController) Inject(controller eudore.Controller, router eudore.Router) error {
	router = router.Group("")
	params := router.Params()
	_, ok := controller.(*TableController)
	if ok && ctl.Name != "" && params.Get("controllergroup") == "" {
		params.Set("controllergroup", ctl.Name)
		params.Set("enable-route-extend", "1")
		params.Set("ignore-init", "1")
		params.Set("ignore-release", "1")
	}
	params.Set("resource-prefix", params.Get("route"))
	err := eudore.ControllerInjectSingleton(controller, router)
	if err != nil {
		return err
	}
	return nil
}

// ControllerRoute 方法返回默认路由信息。
func (ctl *TableController) ControllerRoute() map[string]string {
	return map[string]string{
		"Init":             "",
		"Release":          "",
		"Inject":           "",
		"ControllerRoute":  "",
		"GetRouteParam":    "",
		"QueryValue":       "",
		"QueryValues":      "",
		"SearchConditions": "",
	}
}

// GetRouteParam 方法添加路由参数信息。
func (ctl *TableController) GetRouteParam(pkg, name, method string) string {
	pos := strings.LastIndexByte(pkg, '/') + 1
	if pos != 0 {
		pkg = pkg[pos:]
	}
	if ctl.Pkg != "" {
		pkg = ctl.Pkg
	}
	if strings.HasSuffix(name, "Controller") {
		name = name[:len(name)-len("Controller")]
	}
	if name == "Table" || name == "View" {
		name = ctl.Name
	}
	return fmt.Sprintf("action=%s:%s:%s", pkg, name, method)
}

func (ctl *TableController) QueryValues(ctx Context, conds []string, args []interface{}) (interface{}, error) {
	var size = ctx.GetQueryInt("size", 20)
	var page = ctx.GetQueryInt("page") * size
	var order = ctx.GetQueryString("order", ctl.FieldNames[0])

	var cond string
	if len(conds) != 0 {
		cond = fmt.Sprintf("WHERE %s", strings.Join(conds, " OR "))
	}

	data := ctl.newSliceDataValue()
	err := ctx.QueryBind(data, fmt.Sprintf("SELECT %s FROM %s %s ORDER BY %s limit %d OFFSET %d", ctl.SelectFields, ctl.TableName, cond, order, size, page), args...)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"count": ctx.QueryCount(fmt.Sprintf("SELECT count(1) FROM %s %s", ctl.TableName, cond), args...),
		"field": ctl.FieldNames,
		"data":  data,
	}, nil
}

func (ctl *TableController) QueryValue(ctx Context, conds []string, args []interface{}) (interface{}, error) {
	var cond string
	if len(conds) != 0 {
		cond = fmt.Sprintf("WHERE %s", strings.Join(conds, " OR "))
	}

	data := ctl.newDataValue()
	err := ctx.QueryBind(data, fmt.Sprintf("SELECT * FROM %s %s", ctl.TableName, cond), args...)
	return data, err
}

func (ctl *TableController) GetFields() interface{} {
	return ctl.FieldNames
}

func (ctl *TableController) GetList(ctx Context) (interface{}, error) {
	return ctl.QueryValues(ctx, nil, nil)
}

func (ctl *TableController) GetByKey(ctx Context) (interface{}, error) {
	return ctl.QueryValue(ctx, []string{ctl.FieldNames[0] + "=$1"}, []interface{}{ctx.GetParam("key")})
}

func (ctl *TableController) GetIdById(ctx Context) (interface{}, error) {
	return ctl.QueryValue(ctx, []string{"id=$1"}, []interface{}{ctx.GetParam("id")})
}

func (ctl *TableController) GetNameByName(ctx Context) (interface{}, error) {
	return ctl.QueryValue(ctx, []string{"name=$1"}, []interface{}{ctx.GetParam("name")})
}

func (ctl *TableController) GetSearchByFieldByKey(ctx Context) (interface{}, error) {
	field := ctx.GetParam("field")
	if !ctl.checkField(field) {
		return nil, fmt.Errorf("TableController search field '%s' is invalid", field)
	}
	return ctl.SearchConditions(ctx, []Condition{{Field: field, Value: ctx.GetParam("key")}})

	// var size = ctl.GetQueryDefaultInt("size", 20)
	// var page = ctl.GetQueryInt("page") * size
	// var order = ctl.GetQueryDefaultString("order", ctl.FieldNames[0])
	// pagesql := fmt.Sprintf("SELECT * FROM %s WHERE %s=$1 ORDER BY %s limit %d OFFSET %d;", ctl.TableName, field, order, size, page)
	// fmt.Println(pagesql, ctl.GetParam("key"))
	// return ctl.QueryRows(pagesql, ctl.GetParam("key"))
}

func (ctl *TableController) GetSearchDefaultByKey(ctx Context) (interface{}, error) {
	key := ctx.GetParam("key")
	data := make([]Condition, len(ctl.FieldNames))
	for i := 0; i < len(data); i++ {
		data[i] = Condition{
			Field: ctl.FieldNames[i],
			Value: key,
		}
	}
	return ctl.SearchConditions(ctx, data)
}

/*
[]{"filed","value","expression"}
*/
func (ctl *TableController) PostSearch(ctx Context) (interface{}, error) {
	var data []Condition
	err := ctx.Bind(&data)
	fmt.Printf("%#v %v\n", data, err)
	if err != nil {
		return nil, err
	}
	return ctl.SearchConditions(ctx, data)
}

func (ctl *TableController) PostNew(ctx Context) error {
	data := ctl.newDataValue()
	err := ctx.Bind(&data)
	fmt.Printf("%#v %v\n", data, err)
	if err != nil {
		return err
	}
	return nil
}

func (ctl *TableController) PutNews(ctx Context) error {
	var data []map[string]interface{}
	err := ctx.Bind(&data)
	if err != nil {
		return err
	}
	for _, i := range data {
		ctl.PutNew(ctx, i)
	}
	return nil
}

func (ctl *TableController) PutNew(ctx Context, data map[string]interface{}) error {
	keys := make([]string, 0, len(data))
	vals := make([]interface{}, 0, len(data))
	for key, val := range data {
		if !ctl.checkField(key) {
			return fmt.Errorf("invalid key %s", key)
		}
		keys = append(keys, `"`+key+`"`)
		vals = append(vals, val)
	}

	var crlf string
	for i := range keys {
		crlf = fmt.Sprintf("%s,$%d", crlf, i+1)
	}

	sql := fmt.Sprintf("INSERT INTO %s(%s) VALUES(%s)", ctl.TableName, strings.Join(keys, ","), crlf[1:])
	_, err := ctx.Exec(sql, vals...)
	return err
}

func (ctl *TableController) PutKeyByKey(ctx Context, data map[string]interface{}) error {
	keys := make([]string, 0, len(data))
	args := make([]interface{}, 0, len(data)+1)
	for k, v := range data {
		if !ctl.checkField(k) {
			return fmt.Errorf("invalid key %s", k)
		}
		args = append(args, v)
		keys = append(keys, fmt.Sprintf("\"%s\"=$%d", k, len(args)))
	}
	args = append(args, ctx.GetParam("key"))

	sql := fmt.Sprintf("UPDATE %s SET %s WHERE %s=$%d", ctl.TableName, strings.Join(keys, ","), ctl.FieldNames[0], len(args))
	_, err := ctx.Exec(sql, args...)
	return err
}

func (ctl *TableController) Delete(ctx Context, data map[string]interface{}) error {
	if len(data) == 0 {
		return nil
	}
	i := 1
	keys := make([]string, 0, len(data))
	vals := make([]interface{}, 0, len(data))
	for key, val := range data {
		if !ctl.checkField(key) {
			return fmt.Errorf("invalid key %s", key)
		}
		keys = append(keys, fmt.Sprintf("\"%s\"=$%d", key, i))
		vals = append(vals, val)
		i++
	}

	sql := fmt.Sprintf("DELETE FROM %s WHERE %s", ctl.TableName, strings.Join(keys, " AND "))
	_, err := ctx.Exec(sql, vals...)
	return err
}

// DeleteIdById 方法根据id删除策略。
func (ctl *TableController) DeleteKeyByKey(ctx Context) (err error) {
	_, err = ctx.Exec(fmt.Sprintf("DELETE FROM %s WHERE \"%s\"=$1", ctl.TableName, ctl.FieldNames[0]), ctx.GetParam("key"))
	return
}

// DeleteIdById 方法根据id删除策略。
func (ctl *TableController) DeleteIdById(ctx Context) (err error) {
	_, err = ctx.Exec(fmt.Sprintf("DELETE FROM %s WHERE \"id\"=$1", ctl.TableName), ctx.GetParam("id"))
	return
}

// DeleteNameByName 方法根据名称删除策略。
func (ctl *TableController) DeleteNameByName(ctx Context) (err error) {
	_, err = ctx.Exec(fmt.Sprintf("DELETE FROM %s WHERE \"name\"=$1", ctl.TableName), ctx.GetParam("name"))
	return
}

func (ctl *TableController) SearchConditions(ctx Context, data []Condition) (interface{}, error) {
	var pos = 0
	conds := make([]string, 0, len(data))
	args := make([]interface{}, 0, len(data))
	for _, i := range data {
		switch ctl.getType(i.Field) {
		case "Bool":
			b, isb := strconv.ParseBool(i.Value)
			if isb == nil {
				pos++
				conds = append(conds, fmt.Sprintf("\"%s\"=$%d", i.Field, pos))
				args = append(args, b)
			}
		case "Int":
			num, isnum := strconv.ParseInt(i.Value, 10, 64)
			if isnum == nil && checkNumCompare(i.Expression) {
				if i.Expression == "" {
					i.Expression = "="
				}
				pos++
				conds = append(conds, fmt.Sprintf("\"%s\"%s$%d", i.Field, i.Expression, pos))
				args = append(args, num)
			}
		case "Float":
			num, isnum := strconv.ParseFloat(i.Value, 64)
			if isnum == nil && checkNumCompare(i.Expression) {
				if i.Expression == "" {
					i.Expression = "="
				}
				pos++
				conds = append(conds, fmt.Sprintf("\"%s\"=$%d", i.Field, pos))
				args = append(args, num)
			}
		case "String":
			if checkStringCompare(i.Expression) {
				if i.Expression == "" {
					i.Expression = "LIKE"
				}
				pos++
				conds = append(conds, fmt.Sprintf("\"%s\" %s $%d", i.Field, i.Expression, pos))
				args = append(args, `%`+i.Value+`%`)
			}
		case "DateTime", "Date", "Time":
			var t1 time.Time
			var add int
			switch ctl.getType(i.Field) {
			case "DateTime":
				t1, add = parseDateTime(i.Field, formatDateTimes, addDateTimes)
			case "Date":
				t1, add = parseDateTime(i.Field, formatDates, addDates)
			case "Time":
				t1, add = parseDateTime(i.Field, formatTimes, addTimes)
			}
			if add != 0 {
				pos += 2
				conds = append(conds, fmt.Sprintf("%s BETWEEN $%d AND $%d", i.Field, pos-1, pos))
				args = append(args, t1, timeAdd(t1, add))
			}
		}
	}

	return ctl.QueryValues(ctx, conds, args)
}

func checkNumCompare(flag string) bool {
	switch flag {
	case "=", ">", "<", "<>", "!=", ">=", "<=", "":
		return true
	default:
		return false
	}
}

func checkStringCompare(flag string) bool {
	switch flag {
	case "=", "!=", "LIKE", "NOT LIKE", "":
		return true
	default:
		return false
	}
}

func checkTimeCompare(flag string) bool {
	switch flag {
	case "=", ">", "<", "<>", "!=", ">=", "<=", "":
		return true
	default:
		return false
	}
}

var formatDates = []string{"2006", "2006-1", "2006-01", "2006-1-02", "2006-01-02"}
var addDates = []int{1, 2, 2, 3, 3}
var formatTimes = []string{"15", "15:4", "15:04", "15:04:05"}
var addTimes = []int{4, 5, 5, 6}
var formatDateTimes = []string{"2006", "2006-1", "2006-01", "2006-1-02", "2006-01-02", "2006-1-02 15",
	"2006-01-02 15", "2006-1-02 15:04", "2006-01-02 15:04", "2006-1-02 15:04:05", "2006-01-02 15:04:05",
	"15", "15:4", "15:04", "15:04:05"}
var addDateTimes = []int{1, 2, 2, 3, 3, 4, 4, 5, 5, 6, 6, 4, 5, 5, 6}

func parseDateTime(str string, formats []string, adds []int) (time.Time, int) {
	for i, f := range formats {
		t, err := time.Parse(f, str)
		if err == nil {
			return t, adds[i]
		}
	}
	return time.Unix(1, 0), 0
}

func timeAdd(t time.Time, s int) time.Time {
	t = t.Add(-1)
	switch s {
	case 1:
		return t.AddDate(1, 0, 0)
	case 2:
		return t.AddDate(0, 1, 0)
	case 3:
		return t.AddDate(0, 0, 1)
	case 4:
		return t.Add(time.Hour)
	case 5:
		return t.Add(time.Minute)
	case 6:
		return t.Add(time.Second)
	default:
		return t
	}
}

func (ctl Table) checkField(field string) bool {
	for _, i := range ctl.FieldNames {
		if i == field {
			return true
		}
	}
	return false
}

func (tb Table) getType(field string) string {
	for i := range tb.FieldNames {
		if tb.FieldNames[i] == field {
			return tb.FieldTypes[i]
		}
	}
	return ""
}

func (tb Table) newDataValue() interface{} {
	return reflect.New(tb.DataType).Interface()
}

func (tb Table) newSliceDataValue() interface{} {
	return reflect.New(reflect.SliceOf(tb.DataType)).Interface()
}

func newTable(db *sql.DB, object interface{}, name string) Table {
	// get fields
	rows, err := db.Query("SELECT COLUMN_NAME, data_type,column_default,is_nullable FROM information_schema.COLUMNS WHERE TABLE_NAME = '" + name + "' ORDER BY ordinal_position;")
	if err != nil {
		panic(err)
	}
	tb := Table{TableName: name}
	var field, typ, coldf, isnil string
	for rows.Next() {
		rows.Scan(&field, &typ, &coldf, &isnil)
		tb.FieldNames = append(tb.FieldNames, field)
		tb.FieldTypes = append(tb.FieldTypes, tabletypes[typ])
		tb.FieldDefault = append(tb.FieldDefault, coldf)
		tb.FieldIsNull = append(tb.FieldIsNull, isnil == "YES")
	}
	tb.SelectFields = `"` + strings.Join(tb.FieldNames, `","`) + `"`

	// get struct type
	iValue := reflect.ValueOf(object)
	for iValue.Kind() == reflect.Ptr || iValue.Kind() == reflect.Interface {
		iValue = iValue.Elem()
	}
	if iValue.Kind() == reflect.Struct || iValue.Kind() == reflect.Map {
		tb.DataType = iValue.Type()
	} else {
		var sf []reflect.StructField = make([]reflect.StructField, len(tb.FieldNames))
		for i := 0; i < len(tb.FieldNames); i++ {
			name := strings.ToLower(tb.FieldNames[i])
			sf[i] = reflect.StructField{
				Name: strings.ToTitle(name),
				Type: tableReflectTypes[tb.FieldTypes[i]],
				Tag:  reflect.StructTag(fmt.Sprintf("alias:\"%s\" json:\"%s\"", name, name)),
			}
		}
		tb.DataType = reflect.StructOf(sf)
	}
	return tb
}

var tabletypes = map[string]string{
	"boolean":                     "Bool",
	"integer":                     "Int",
	"real":                        "Float",
	"character":                   "Byte",
	"character varying":           "String",
	"text":                        "String",
	"date":                        "Date",
	"time without time zone":      "Time",
	"timestamp without time zone": "DateTime",
}

var tableReflectTypes = map[string]reflect.Type{
	"Bool":     reflect.TypeOf((*bool)(nil)).Elem(),
	"Int":      reflect.TypeOf((*int)(nil)).Elem(),
	"Float":    reflect.TypeOf((*float64)(nil)).Elem(),
	"Byte":     reflect.TypeOf((*byte)(nil)).Elem(),
	"String":   reflect.TypeOf((*string)(nil)).Elem(),
	"Date":     reflect.TypeOf((*time.Time)(nil)).Elem(),
	"Time":     reflect.TypeOf((*time.Time)(nil)).Elem(),
	"DateTime": reflect.TypeOf((*time.Time)(nil)).Elem(),
}

func NewTable2Struct(db *sql.DB, name string) {
	// get fields
	rows, err := db.Query("SELECT COLUMN_NAME, data_type FROM information_schema.COLUMNS WHERE TABLE_NAME = '" + name + "' ORDER BY ordinal_position;")
	if err != nil {
		panic(err)
	}
	var field, typ string
	fmt.Printf("\ntype %s struct{\n", name)
	for rows.Next() {
		rows.Scan(&field, &typ)
		fmt.Printf("\t%s \t%s \t`alias:\"%s\" json:\"%s\"`\n", strings.Title(field), tableReflectTypes[tabletypes[typ]], strings.ToLower(field), strings.ToLower(field))
	}
	fmt.Print("}\n\n")
}

func NewViewController(pkg, name, tb string, object interface{}, db *sql.DB) *ViewController {
	return &ViewController{
		TableController: TableController{
			Table: newTable(db, object, tb),
			Pkg:   pkg,
			Name:  name,
		},
	}
}

// ControllerRoute 方法返回默认路由信息。
func (ctl *ViewController) ControllerRoute() map[string]string {
	return map[string]string{
		"Init":             "",
		"Release":          "",
		"Inject":           "",
		"ControllerRoute":  "",
		"GetRouteParam":    "",
		"QueryValue":       "",
		"QueryValues":      "",
		"SearchConditions": "",
		// delete route method
		"PostNew":          "",
		"PostSearch":       "",
		"PutKeyByKey":      "",
		"PutNew":           "",
		"DeleteIdById":     "",
		"DeleteKeyByKey":   "",
		"DeleteNameByName": "",
	}
}

func (ctl *ViewController) Inject(controller eudore.Controller, router eudore.Router) error {
	router = router.Group("")
	params := router.Params()
	if ctl == controller && params.Get("controllergroup") == "" {
		params.Set("controllergroup", ctl.Name+"/view")
		params.Set("enable-route-extend", "1")
		params.Set("ignore-init", "1")
		params.Set("ignore-release", "1")
	}
	return ctl.TableController.Inject(controller, router)
}
