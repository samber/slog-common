package slogcommon

import (
	"encoding"
	"fmt"
	"log/slog"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/samber/lo"
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

func AttrsToString(attrs ...slog.Attr) map[string]string {
	output := map[string]string{}

	for i := range attrs {
		attr := attrs[i]
		k, v := attr.Key, attr.Value
		output[k] = ValueToString(v)
	}

	return output
}

func ValueToString(v slog.Value) string {
	switch v.Kind() {
	case slog.KindAny:
		return AnyValueToString(v)
	case slog.KindLogValuer:
		return AnyValueToString(v)
	case slog.KindGroup:
		return AnyValueToString(v)
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

func FormatErrorKey(values map[string]any, errorKeys ...string) map[string]any {
	for _, errorKey := range errorKeys {
		if err, ok := values[errorKey]; ok {
			if e, ok := err.(error); ok {
				values[errorKey] = FormatError(e)
				break
			}
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

func FormatRequest(req *http.Request, ignoreHeaders bool) map[string]any {
	output := map[string]any{
		"host":   req.Host,
		"method": req.Method,
		"url": map[string]any{
			"url":       req.URL.String(),
			"scheme":    req.URL.Scheme,
			"host":      req.URL.Host,
			"path":      req.URL.Path,
			"raw_query": req.URL.RawQuery,
			"fragment":  req.URL.Fragment,
			"query": lo.MapEntries(req.URL.Query(), func(key string, values []string) (string, string) {
				return key, strings.Join(values, ",")
			}),
		},
	}

	if !ignoreHeaders {
		output["headers"] = lo.MapEntries(req.Header, func(key string, values []string) (string, string) {
			return key, strings.Join(values, ",")
		})
	}

	return output
}
