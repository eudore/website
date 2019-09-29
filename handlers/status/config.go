package status

import (
	"fmt"
	"github.com/eudore/eudore"
	"reflect"
)

type pair struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func getConfig(app *eudore.App) eudore.HandlerFunc {
	conf := app.Config
	return func(ctx eudore.Context) {
		ctx.WriteJSON(filter(listKey("", reflect.ValueOf(conf), make([]pair, 0, 20))))
	}
}

func filter(vals []pair) []pair {
	for i, val := range vals {
		vals[i].Name = val.Name[1:]
	}
	return vals
}

func listKey(prefix string, iValue reflect.Value, vals []pair) []pair {
	switch iValue.Type().Kind() {
	case reflect.Ptr:
		if !iValue.IsNil() {
			return listKey(prefix, iValue.Elem(), vals)
		}
	case reflect.Map:
		return listMap(prefix, iValue, vals)
	case reflect.Struct:
		return listStruct(prefix, iValue, vals)
	default:
		if iValue.Type().Kind() == reflect.Interface {
			switch iValue.Elem().Kind() {
			case reflect.Ptr, reflect.Map, reflect.Struct, reflect.Interface, reflect.Array, reflect.Slice:
				return vals
			}
		}
		vals = append(vals, pair{prefix, fmt.Sprint(iValue.Interface())})
	}
	return vals
}

func listMap(prefix string, iValue reflect.Value, vals []pair) []pair {
	for _, key := range iValue.MapKeys() {
		vals = listKey(prefix+"."+fmt.Sprint(key.Interface()), iValue.MapIndex(key), vals)
	}

	return vals
}

func listStruct(prefix string, iValue reflect.Value, vals []pair) []pair {
	iType := iValue.Type()
	for i := 0; i < iType.NumField(); i++ {
		vals = listKey(prefix+"."+iType.Field(i).Name, iValue.Field(i), vals)
	}
	return vals
}
