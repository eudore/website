package status

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/eudore/eudore"
	"github.com/eudore/website/config"
)

var typeInterface = reflect.TypeOf((*interface{})(nil)).Elem()

type pair struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func getConfig(conf *config.Config) eudore.HandlerFunc {
	prefixs := []string{
		"keys.configdata",
		"keys.ram.",
		"mods.",
		"auth.sender.mail.password",
		"auth.secrets.",
		"component.pprof.basicauth.",
	}
	return func(ctx eudore.Context) {
		ctx.WriteJSON(filter(listKey("", reflect.ValueOf(conf), make([]pair, 0, 20)), prefixs))
	}
}

func filter(vals []pair, prefixs []string) []pair {
	for i, val := range vals {
		vals[i].Name = strings.ToLower(val.Name[1:])
		for _, prefix := range prefixs {
			if strings.HasPrefix(vals[i].Name, prefix) {
				vals[i].Name = ""
				vals[i].Value = ""
			}
		}
	}
	return vals
}

func listKey(prefix string, iValue reflect.Value, vals []pair) []pair {
	switch iValue.Type().Kind() {
	case reflect.Ptr, reflect.Interface:
		if !iValue.IsNil() {
			return listKey(prefix, iValue.Elem(), vals)
		}
	case reflect.Map:
		return listMap(prefix, iValue, vals)
	case reflect.Struct:
		return listStruct(prefix, iValue, vals)
	default:
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
		if iValue.Field(i).CanSet() {
			vals = listKey(prefix+"."+iType.Field(i).Name, iValue.Field(i), vals)
		}
	}
	return vals
}
