package framework

import (
	"reflect"
	"strings"
)

var typeInterface = reflect.TypeOf((*interface{})(nil)).Elem()
var typeString = reflect.TypeOf((*string)(nil)).Elem()

type BindValue struct {
	reflect.Value
}

func NewBindValue(i interface{}) BindValue {
	return BindValue{bindinit(reflect.ValueOf(i))}
}

func (bv BindValue) IsSlice() bool {
	return bv.Kind() == reflect.Slice
}
func (bv BindValue) IsBind() bool {
	if bv.Kind() == reflect.Struct || bv.Kind() == reflect.Map {
		return true
	}
	v2 := bindinit(reflect.New(bv.Type().Elem()).Elem())
	if v2.Kind() == reflect.Struct || v2.Kind() == reflect.Map {
		return true
	}
	return false
}

func (bv BindValue) Bind(keys []string, vals []interface{}) {
	if bv.Kind() != reflect.Slice {
		bv.binddata(bv.Value, keys, vals)
		return
	}
	newValue := reflect.New(bv.Type().Elem()).Elem()
	bv.binddata(newValue, keys, vals)
	bv.Set(reflect.Append(bv.Value, newValue))

}
func (bv BindValue) binddata(iValue reflect.Value, keys []string, vals []interface{}) {
	switch iValue.Kind() {
	case reflect.Ptr:
		if iValue.IsNil() {
			iValue.Set(reflect.New(iValue.Type().Elem()))
		}
		bv.binddata(iValue.Elem(), keys, vals)
	case reflect.Interface:
		if iValue.IsNil() && iValue.Type() == typeInterface {
			iValue.Set(reflect.ValueOf(make(map[string]interface{})))
			bv.binddata(iValue.Elem(), keys, vals)
		}
	case reflect.Map:
		iType := iValue.Type()
		if iValue.IsNil() {
			iValue.Set(reflect.MakeMap(iType))
		}
		if !typeString.ConvertibleTo(iType.Key()) {
			return
		}
		for i := 0; i < len(keys); i++ {
			if vals[i] == nil {
				continue
			}
			mapKey := reflect.New(iType.Key()).Elem()
			if mapKey.Type() == typeString {
				mapKey.SetString(keys[i])
			} else {
				mapKey.Set(reflect.ValueOf(keys[i]).Convert(iType.Key()))
			}

			mapValue := reflect.ValueOf(vals[i])
			if mapValue.Type() == iType.Elem() {
				iValue.SetMapIndex(mapKey, mapValue)
			} else if mapValue.Type().ConvertibleTo(iType.Elem()) {
				iValue.SetMapIndex(mapKey, mapValue.Convert(iType.Elem()))
			}
		}
	case reflect.Struct:
		iType := iValue.Type()
		for i := 0; i < len(keys); i++ {
			index := getStructIndexOfTags(iType, keys[i], []string{"alias"})
			if index == -1 || vals[i] == nil {
				continue
			}
			bv.binddataStruct(iType.Field(i), iValue.Field(i), vals[i])
			// field := iValue.Field(index)
			// sValue := reflect.ValueOf(vals[i])
			// if field.Type() == sValue.Type() {
			// 	if field.Type() == typeString && iType.Field(i).Tag.Get("masking") != "" {
			// 		field.SetString(passwordMasking(sValue.String()))
			// 	} else {
			// 		field.Set(sValue)
			// 	}

			// } else if sValue.Type().ConvertibleTo(field.Type()) {
			// 	field.Set(sValue.Convert(field.Type()))
			// }
		}
	}
}

func (bv BindValue) binddataStruct(iField reflect.StructField, iValue reflect.Value, val interface{}) {
	rval := reflect.ValueOf(val)
	// 类型相同
	if iValue.Type() == rval.Type() {
		// 字符串脱敏
		if iField.Type == typeString && iField.Tag.Get("masking") != "" {
			iValue.SetString(passwordMasking(val.(string)))
		} else {
			iValue.Set(rval)
		}
		return
	}

	// 数组类型
	char := iField.Tag.Get("splitchar")
	if char != "" && rval.Type() == typeString && val.(string) != "" && iField.Type.Kind() == reflect.Slice {
		strs := strings.Split(val.(string), char)
		if iField.Type.Kind() == reflect.Slice {
			iValue.Set(reflect.MakeSlice(reflect.SliceOf(typeString), len(strs), len(strs)))
		}
		for i, sub := range strs {
			iValue.Index(i).SetString(sub)
		}
		return
	}

	// 尝试类型转换
	if rval.Type().ConvertibleTo(iField.Type) {
		iValue.Set(rval.Convert(iField.Type))
	}
}

func passwordMasking(str string) string {
	slen := len(str)
	if slen < 5 {
		return "***"
	}
	if slen < 16 {
		return str[:2] + "****" + str[slen-2:]
	}
	return str[:4] + "*****" + str[slen-4:]
}

// 通过字符串获取结构体属性的索引
func getStructIndexOfTags(iType reflect.Type, name string, tags []string) int {
	// 遍历匹配
	for i := 0; i < iType.NumField(); i++ {
		typeField := iType.Field(i)
		// 字符串为结构体名称或结构体属性标签的值，则匹配返回索引。
		if typeField.Name == name {
			return i
		}
		for _, tag := range tags {
			if typeField.Tag.Get(tag) == name {
				return i
			}
		}
	}
	return -1
}

func bindinit(iValue reflect.Value) reflect.Value {
	switch iValue.Kind() {
	case reflect.Ptr:
		if iValue.IsNil() {
			iValue.Set(reflect.New(iValue.Type().Elem()))
		}
		return bindinit(iValue.Elem())
	case reflect.Interface:
		if iValue.IsNil() {
			if iValue.Type() != typeInterface {
				return iValue
			}
			iValue.Set(reflect.ValueOf(make(map[string]interface{})))
		}
		return bindinit(iValue.Elem())
	case reflect.Slice, reflect.Struct:
		return iValue
	case reflect.Map:
		if iValue.IsNil() {
			iValue.Set(reflect.MakeMap(iValue.Type()))
		}
		return iValue
	}
	return iValue
}
