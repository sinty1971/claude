package core

import (
	"errors"
	"reflect"
	"strconv"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// GetFieldAs は protobuf メッセージのフィールドを指定型 V で返します。
// サポート: 文字列・整数・符号無し整数・浮動小数点・真偽値・メッセージ型の直接アサーション。
// 文字列フィールドを数値型に要求した場合はパースを試みます。
func GetFieldAs[V any](message proto.Message, fieldName string) (V, error) {
	var zero V

	if message == nil {
		return zero, errors.New("message is nil")
	}

	msgRef := message.ProtoReflect()
	fields := msgRef.Descriptor().Fields()
	fd := fields.ByName(protoreflect.Name(fieldName))
	if fd == nil {
		return zero, errors.New("field " + fieldName + " not found")
	}
	if !msgRef.Has(fd) {
		return zero, errors.New("field '" + fieldName + "' is not set")
	}

	val := msgRef.Get(fd)
	var rval reflect.Value
	typ := reflect.TypeOf(zero)

	switch fd.Kind() {
	case protoreflect.StringKind:
		s := val.String()
		switch typ.Kind() {
		case reflect.String:
			return any(s).(V), nil
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			bits := typ.Bits()
			i, err := strconv.ParseInt(s, 10, bits)
			if err != nil {
				return zero, err
			}
			rval = reflect.New(typ).Elem()
			rval.SetInt(i)
			return rval.Interface().(V), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			bits := typ.Bits()
			u, err := strconv.ParseUint(s, 10, bits)
			if err != nil {
				return zero, err
			}
			rval = reflect.New(typ).Elem()
			rval.SetUint(u)
			return rval.Interface().(V), nil
		case reflect.Float32, reflect.Float64:
			bits := typ.Bits()
			f, err := strconv.ParseFloat(s, bits)
			if err != nil {
				return zero, err
			}
			rval = reflect.New(typ).Elem()
			rval.SetFloat(f)
			return rval.Interface().(V), nil
		case reflect.Bool:
			b, err := strconv.ParseBool(s)
			if err != nil {
				return zero, err
			}
			return any(b).(V), nil
		default:
			return zero, errors.New("unsupported target type for string field")
		}

	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind,
		protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		i := val.Int()
		switch typ.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			rval = reflect.New(typ).Elem()
			rval.SetInt(i)
			return rval.Interface().(V), nil
		case reflect.String:
			return any(strconv.FormatInt(i, 10)).(V), nil
		default:
			return zero, errors.New("unsupported target type for int field")
		}

	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind, protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		u := val.Uint()
		switch typ.Kind() {
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			rval = reflect.New(typ).Elem()
			rval.SetUint(u)
			return rval.Interface().(V), nil
		case reflect.String:
			return any(strconv.FormatUint(u, 10)).(V), nil
		default:
			return zero, errors.New("unsupported target type for uint field")
		}

	case protoreflect.FloatKind, protoreflect.DoubleKind:
		f := val.Float()
		switch typ.Kind() {
		case reflect.Float32, reflect.Float64:
			rval = reflect.New(typ).Elem()
			rval.SetFloat(f)
			return rval.Interface().(V), nil
		case reflect.String:
			return any(strconv.FormatFloat(f, 'f', -1, typ.Bits())).(V), nil
		default:
			return zero, errors.New("unsupported target type for float field")
		}

	case protoreflect.BoolKind:
		b := val.Bool()
		if typ.Kind() == reflect.Bool {
			return any(b).(V), nil
		}
		if typ.Kind() == reflect.String {
			return any(strconv.FormatBool(b)).(V), nil
		}
		return zero, errors.New("unsupported target type for bool field")

	case protoreflect.MessageKind, protoreflect.GroupKind:
		// Try to assert the underlying proto message to V
		if msg := val.Message().Interface(); msg != nil {
			if res, ok := msg.(V); ok {
				return res, nil
			}
			// allow returning as interface{} when V is interface type
			rv := reflect.ValueOf(msg)
			if rv.Type().AssignableTo(typ) {
				return rv.Interface().(V), nil
			}
		}
		return zero, errors.New("message field cannot be converted to requested type")

	default:
		return zero, errors.New("unsupported field kind")
	}
}
