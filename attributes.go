package slogcommon

import (
	"encoding"
	"fmt"
	"log/slog"
	"net/http"
	"reflect"
	"runtime"
	"strconv"
	"strings"

	"github.com/samber/lo"
)

type ReplaceAttrFn = func(groups []string, a slog.Attr) slog.Attr

func AppendRecordAttrsToAttrs(attrs []slog.Attr, groups []string, record *slog.Record) []slog.Attr {
	output := append(attrs) // copy

	groups = reverse(groups)
	record.Attrs(func(attr slog.Attr) bool {
		for i := range groups {
			attr = slog.Group(groups[i], attr)
		}
		output = append(output, attr)
		return true
	})

	return output
}

func ReplaceAttrs(fn ReplaceAttrFn, groups []string, attrs ...slog.Attr) []slog.Attr {
	if fn == nil {
		return attrs
	}

	for i := range attrs {
		attr := attrs[i]
		value := attr.Value.Resolve()
		if value.Kind() == slog.KindGroup {
			attr.Value = slog.GroupValue(ReplaceAttrs(fn, append(groups, attr.Key), value.Group()...)...)
		} else {
			attrs[i] = fn(groups, attr)
		}
	}

	return attrs
}

func AttrsToMap(attrs ...slog.Attr) map[string]any {
	output := map[string]any{}

	attrsByKey := groupValuesByKey(attrs)
	for k, values := range attrsByKey {
		v := mergeAttrValues(values...)
		if v.Kind() == slog.KindGroup {
			output[k] = AttrsToMap(v.Group()...)
		} else {
			output[k] = v.Any()
		}
	}

	return output
}

func groupValuesByKey(attrs []slog.Attr) map[string][]slog.Value {
	result := map[string][]slog.Value{}

	for _, item := range attrs {
		key := item.Key
		result[key] = append(result[key], item.Value)
	}

	return result
}

func mergeAttrValues(values ...slog.Value) slog.Value {
	v := values[0]

	for i := 1; i < len(values); i++ {
		if v.Kind() != slog.KindGroup || values[i].Kind() != slog.KindGroup {
			v = values[i]
			continue
		}

		v = slog.GroupValue(append(v.Group(), values[i].Group()...)...)
	}

	return v
}

func mergeNestedMap(m1, m2 map[string]any) map[string]any {
	for k, v1 := range m1 {
		if v1, ok := v1.(map[string]any); ok {
			if v2, ok := m1[k]; ok {
				if v2, ok := v2.(map[string]any); ok {
					m1[k] = mergeNestedMap(v2, v1)
					continue
				}
			}
		}
		m1[k] = v1
	}

	return m1
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
		return k, AttrsToMap(v.Group()...)
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

func Source(sourceKey string, r *slog.Record) slog.Attr {
	fs := runtime.CallersFrames([]uintptr{r.PC})
	f, _ := fs.Next()
	return slog.Any(
		sourceKey,
		&slog.Source{
			Function: f.Function,
			File:     f.File,
			Line:     f.Line,
		},
	)
}
