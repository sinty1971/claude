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
	fd := msgRef.Descriptor().Fields().ByName(protoreflect.Name(fieldName))
	if fd == nil {
		return zero, errors.New("field " + fieldName + " not found")
	}
	if !msgRef.Has(fd) {
		return zero, errors.New("field '" + fieldName + "' is not set")
	}

	val := msgRef.Get(fd)
	return convertValue[V](val, fd.Kind())
}

// convertValue はフィールド値を指定型 V に変換します
func convertValue[V any](val protoreflect.Value, kind protoreflect.Kind) (V, error) {
	var zero V
	targetType := reflect.TypeOf(zero)

	switch kind {
	case protoreflect.StringKind:
		return convertFromString[V](val.String(), targetType)

	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind,
		protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return convertFromInt[V](val.Int(), targetType)

	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind,
		protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return convertFromUint[V](val.Uint(), targetType)

	case protoreflect.FloatKind, protoreflect.DoubleKind:
		return convertFromFloat[V](val.Float(), targetType)

	case protoreflect.BoolKind:
		return convertFromBool[V](val.Bool(), targetType)

	case protoreflect.MessageKind, protoreflect.GroupKind:
		if msg := val.Message().Interface(); msg != nil {
			if res, ok := msg.(V); ok {
				return res, nil
			}
			if reflect.ValueOf(msg).Type().AssignableTo(targetType) {
				return reflect.ValueOf(msg).Interface().(V), nil
			}
		}
		return zero, errors.New("message field cannot be converted to requested type")

	default:
		return zero, errors.New("unsupported field kind")
	}
}

// convertFromString は文字列を指定型に変換します
func convertFromString[V any](s string, targetType reflect.Type) (V, error) {
	var zero V
	switch targetType.Kind() {
	case reflect.String:
		return any(s).(V), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(s, 10, targetType.Bits())
		return setReflectValue[V](targetType, func(rv reflect.Value) { rv.SetInt(i) }), err
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u, err := strconv.ParseUint(s, 10, targetType.Bits())
		return setReflectValue[V](targetType, func(rv reflect.Value) { rv.SetUint(u) }), err
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(s, targetType.Bits())
		return setReflectValue[V](targetType, func(rv reflect.Value) { rv.SetFloat(f) }), err
	case reflect.Bool:
		b, err := strconv.ParseBool(s)
		return any(b).(V), err
	default:
		return zero, errors.New("unsupported target type for string field")
	}
}

// convertFromInt は整数を指定型に変換します
func convertFromInt[V any](i int64, targetType reflect.Type) (V, error) {
	var zero V
	switch targetType.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return setReflectValue[V](targetType, func(rv reflect.Value) { rv.SetInt(i) }), nil
	case reflect.String:
		return any(strconv.FormatInt(i, 10)).(V), nil
	default:
		return zero, errors.New("unsupported target type for int field")
	}
}

// convertFromUint は符号なし整数を指定型に変換します
func convertFromUint[V any](u uint64, targetType reflect.Type) (V, error) {
	var zero V
	switch targetType.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return setReflectValue[V](targetType, func(rv reflect.Value) { rv.SetUint(u) }), nil
	case reflect.String:
		return any(strconv.FormatUint(u, 10)).(V), nil
	default:
		return zero, errors.New("unsupported target type for uint field")
	}
}

// convertFromFloat は浮動小数点数を指定型に変換します
func convertFromFloat[V any](f float64, targetType reflect.Type) (V, error) {
	var zero V
	switch targetType.Kind() {
	case reflect.Float32, reflect.Float64:
		return setReflectValue[V](targetType, func(rv reflect.Value) { rv.SetFloat(f) }), nil
	case reflect.String:
		return any(strconv.FormatFloat(f, 'f', -1, targetType.Bits())).(V), nil
	default:
		return zero, errors.New("unsupported target type for float field")
	}
}

// convertFromBool は真偽値を指定型に変換します
func convertFromBool[V any](b bool, targetType reflect.Type) (V, error) {
	var zero V
	switch targetType.Kind() {
	case reflect.Bool:
		return any(b).(V), nil
	case reflect.String:
		return any(strconv.FormatBool(b)).(V), nil
	default:
		return zero, errors.New("unsupported target type for bool field")
	}
}

// setReflectValue は reflect 経由で値を設定するヘルパー
func setReflectValue[V any](typ reflect.Type, setter func(reflect.Value)) V {
	rv := reflect.New(typ).Elem()
	setter(rv)
	return rv.Interface().(V)
}
