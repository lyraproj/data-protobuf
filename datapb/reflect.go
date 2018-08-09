package datapb

import (
	"reflect"
	"fmt"
)

func ToDataHash(h map[string]reflect.Value) (*DataHash, error) {
	cnt := len(h)
	els := make([]*DataEntry, 0, cnt)
	for k, v := range h {
		dev, err := ToData(v)
		if err != nil {
			return nil, err
		}
		els = append(els, &DataEntry{k, dev})
	}
	return &DataHash{els}, nil
}

func ToData(v reflect.Value) (*Data, error) {
	if !v.IsValid() {
		return &Data{Kind: &Data_UndefValue{}}, nil
	}

	switch v.Kind() {
	case reflect.Bool:
		return &Data{Kind: &Data_BooleanValue{v.Bool()}}, nil
	case reflect.Float32, reflect.Float64:
		return &Data{Kind: &Data_FloatValue{v.Float()}}, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return &Data{Kind: &Data_IntegerValue{v.Int()}}, nil
	case reflect.String:
		return &Data{Kind: &Data_StringValue{v.String()}}, nil
	case reflect.Slice, reflect.Array:
		cnt := v.Len()
		els := make([]*Data, cnt)
		for i := 0; i < cnt; i++ {
			elem, err := ToData(v.Index(i))
			if err != nil {
				return nil, err
			}
			els[i] = elem
		}
		return &Data{Kind: &Data_ArrayValue{&DataArray{els}}}, nil
	case reflect.Map:
		keys := v.MapKeys()
		hash := make(map[string]reflect.Value, len(keys))
		for _, k := range keys {
			if !(k.IsValid() && k.Kind() == reflect.String) {
				return nil, fmt.Errorf(`expected hash key to be 'string', got '%s'`, k.Type())
			}
			hash[k.String()] = v.MapIndex(k)
		}
		dh, err := ToDataHash(hash)
		if err != nil {
			return nil, err
		}
		return &Data{Kind: &Data_HashValue{dh}}, nil
	case reflect.Interface:
		if v.Type() == interfaceType && v.Interface() == nil {
			// The interface{} nil value represents a generic nil
			return &Data{Kind: &Data_UndefValue{}}, nil
		}
		// Other interfaces/values cannot be converted
		fallthrough
	default:
		return nil, fmt.Errorf(`unable to convert a value of type '%s' to Data`, v.Type())
	}
}

var interfaceType = reflect.TypeOf([]interface{}{}).Elem()
var stringType = reflect.TypeOf(``)
var GenericNilValue = reflect.Zero(interfaceType)
var InvalidValue = reflect.ValueOf(nil)

func FromDataHash(h *DataHash) (map[string]reflect.Value, error) {
	av := h.Entries
	hash := make(map[string]reflect.Value, len(av))
	for _, elem := range av {
		rv, err := FromData(elem.Value)
		if err != nil {
			return nil, err
		}
		hash[elem.Key] = rv
	}
	return hash, nil
}

func FromData(v *Data) (reflect.Value, error) {
	if v.Kind == nil {
		return GenericNilValue, nil
	}

	switch v.Kind.(type) {
	case *Data_BooleanValue:
		return reflect.ValueOf(v.GetBooleanValue()), nil
	case *Data_FloatValue:
		return reflect.ValueOf(v.GetFloatValue()), nil
	case *Data_IntegerValue:
		return reflect.ValueOf(v.GetIntegerValue()), nil
	case *Data_StringValue:
		return reflect.ValueOf(v.GetStringValue()), nil
	case *Data_UndefValue:
		return GenericNilValue, nil
	case *Data_ArrayValue:
		av := v.GetArrayValue().GetValues()
		vals := make([]reflect.Value, len(av))
		var et reflect.Type = nil
		for i, elem := range av {
			rv, err := FromData(elem)
			if err != nil {
				return InvalidValue, err
			}
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
		return reflect.Append(reflect.MakeSlice(reflect.SliceOf(et), 0, len(vals)), vals...), nil
	case *Data_HashValue:
		av := v.GetHashValue().Entries
		vals := make([]reflect.Value, len(av))
		keys := make([]reflect.Value, len(av))
		var et reflect.Type = nil
		for i, elem := range av {
			keys[i] = reflect.ValueOf(elem.Key)
			rv, err := FromData(elem.Value)
			if err != nil {
				return InvalidValue, err
			}
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
		hash := reflect.MakeMapWithSize(reflect.MapOf(stringType, et), len(vals))
		for i, k := range keys {
			hash.SetMapIndex(k, vals[i])
		}
		return hash, nil
	default:
		return InvalidValue, fmt.Errorf(`unable to convert a value of type '%T' to reflect.Value`, v.Kind)
	}
}
