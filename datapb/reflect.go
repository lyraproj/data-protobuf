package datapb

import (
	"reflect"
	"fmt"
)

func ToData(v reflect.Value) (value *Data, err error) {
	if !v.IsValid() {
		value = &Data{Kind: &Data_UndefValue{}}
		return
	}

	switch v.Kind() {
	case reflect.Bool:
		value = &Data{Kind: &Data_BooleanValue{v.Bool()}}
	case reflect.Float32, reflect.Float64:
		value = &Data{Kind: &Data_FloatValue{v.Float()}}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		value = &Data{Kind: &Data_IntegerValue{v.Int()}}
	case reflect.String:
		value = &Data{Kind: &Data_StringValue{v.String()}}
	case reflect.Slice, reflect.Array:
		cnt := v.Len()
		els := make([]*Data, cnt)
		for i := 0; i < cnt; i++ {
			els[i], err = ToData(v.Index(i))
			if err != nil {
				return
			}
		}
		value = &Data{Kind: &Data_ArrayValue{&DataArray{els}}}
	case reflect.Map:
		keys := v.MapKeys()
		cnt := len(keys)
		els := make([]*DataEntry, cnt)
		var dev *Data
		for i, k := range keys {
			dev, err = ToData(v.MapIndex(k))
			if err != nil {
				return
			}
			if !(k.IsValid() && k.Kind() == reflect.String) {
				err = fmt.Errorf(`expected hash key to be 'string', got '%s'`, k.Type())
				return
			}
			els[i] = &DataEntry{k.String(), dev}
		}
		value = &Data{Kind: &Data_HashValue{&DataHash{els}}}
	default:
		err = fmt.Errorf(`unable to convert a value of type '%s' to Data`, v.Type())
	}
	return
}

var interfaceType = reflect.TypeOf([]interface{}{}).Elem()
var stringType = reflect.TypeOf(``)

func FromData(v *Data) (value reflect.Value, err error) {
	switch v.Kind.(type) {
	case *Data_BooleanValue:
		value = reflect.ValueOf(v.GetBooleanValue())
	case *Data_FloatValue:
		value = reflect.ValueOf(v.GetFloatValue())
	case *Data_IntegerValue:
		value = reflect.ValueOf(v.GetIntegerValue())
	case *Data_StringValue:
		value = reflect.ValueOf(v.GetStringValue())
	case *Data_UndefValue:
		value = reflect.ValueOf(nil)
	case *Data_ArrayValue:
		av := v.GetArrayValue().GetValues()
		vals := make([]reflect.Value, len(av))
		var et reflect.Type = nil
		var rv reflect.Value
		for i, elem := range av {
			rv, err = FromData(elem)
			rt := rv.Type()
			if et == nil {
				et = rt
			} else if et != rt {
				et = nil
			}
			vals[i] = rv
		}
		if et == nil {
			et = interfaceType
		}
		value = reflect.Append(reflect.MakeSlice(reflect.SliceOf(et), 0, len(vals)), vals...)
	case *Data_HashValue:
		av := v.GetHashValue().Entries
		vals := make([]reflect.Value, len(av))
		keys := make([]reflect.Value, len(av))
		var et reflect.Type = nil
		var rv reflect.Value
		for i, elem := range av {
			keys[i] = reflect.ValueOf(elem.Key)
			rv, err = FromData(elem.Value)
			rt := rv.Type()
			if et == nil {
				et = rt
			} else if et != rt {
				et = nil
			}
		}
		if et == nil {
			et = interfaceType
		}
		value = reflect.MakeMapWithSize(reflect.MapOf(stringType, et), len(vals))
		for i, k := range keys {
			value.SetMapIndex(k, vals[i])
		}
	default:
		err = fmt.Errorf(`unable to convert a value of type '%T' to reflect.Value`, v.Kind)
	}
	return
}
