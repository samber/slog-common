package slogcommon

import (
	"encoding"
	"fmt"
	"log/slog"
	"reflect"
	"strconv"
)

func AttrsToValue(attrs ...slog.Attr) map[string]any {
	log := map[string]any{}

	for i := range attrs {
		k, v := AttrToValue(attrs[i])
		log[k] = v
	}

	return log
}

func AttrToValue(attr slog.Attr) (string, any) {
	k := attr.Key
	v := attr.Value
	kind := v.Kind()

	switch kind {
	case slog.KindAny:
		return k, v.Any()
	case slog.KindLogValuer:
		return k, v.Any()
	case slog.KindGroup:
		return k, AttrsToValue(v.Group()...)
	case slog.KindInt64:
		return k, v.Int64()
	case slog.KindUint64:
		return k, v.Uint64()
	case slog.KindFloat64:
		return k, v.Float64()
	case slog.KindString:
		return k, v.String()
	case slog.KindBool:
		return k, v.Bool()
	case slog.KindDuration:
		return k, v.Duration()
	case slog.KindTime:
		return k, v.Time().UTC()
	default:
		return k, AnyValueToString(v)
	}
}

func AnyValueToString(v slog.Value) string {
	if tm, ok := v.Any().(encoding.TextMarshaler); ok {
		data, err := tm.MarshalText()
		if err != nil {
			return ""
		}

		return string(data)
	}

	return fmt.Sprintf("%+v", v.Any())
}

func ValueToString(v slog.Value) string {
	switch v.Kind() {
	case slog.KindAny:
		return AnyValueToString(v)
	case slog.KindLogValuer:
		return AnyValueToString(v)
	case slog.KindGroup:
		return fmt.Sprint(v)
	case slog.KindInt64:
		return fmt.Sprintf("%d", v.Int64())
	case slog.KindUint64:
		return fmt.Sprintf("%d", v.Uint64())
	case slog.KindFloat64:
		return fmt.Sprintf("%f", v.Float64())
	case slog.KindString:
		return v.String()
	case slog.KindBool:
		return strconv.FormatBool(v.Bool())
	case slog.KindDuration:
		return v.Duration().String()
	case slog.KindTime:
		return v.Time().UTC().String()
	default:
		return AnyValueToString(v)
	}
}

func FormatErrorKey(values map[string]any, errorKey string) map[string]any {
	if err, ok := values["error"]; ok {
		if e, ok := err.(error); ok {
			values["error"] = FormatError(e)
		}
	} else if err, ok := values["err"]; ok {
		if e, ok := err.(error); ok {
			values["err"] = FormatError(e)
		}
	} else if err, ok := values[errorKey]; ok {
		if e, ok := err.(error); ok {
			values[errorKey] = FormatError(e)
		}
	}

	return values
}

func FormatError(err error) map[string]any {
	return map[string]any{
		"kind":  reflect.TypeOf(err).String(),
		"error": err.Error(),
		"stack": nil, // @TODO
	}
}
