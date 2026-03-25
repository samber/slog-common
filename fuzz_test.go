package slogcommon

import (
	"log/slog"
	"testing"
	"time"
)

func FuzzAttrsToMap(f *testing.F) {
	f.Add("key1", "value1", "key2", "value2")
	f.Add("", "", "k", "v")
	f.Add("same", "first", "same", "second")

	f.Fuzz(func(t *testing.T, k1, v1, k2, v2 string) {
		attrs := []slog.Attr{
			slog.String(k1, v1),
			slog.String(k2, v2),
		}
		// Should not panic
		_ = AttrsToMap(attrs...)
	})
}

func FuzzValueToString(f *testing.F) {
	f.Add("hello")
	f.Add("")
	f.Add("special chars: <>&\"'\\")

	f.Fuzz(func(t *testing.T, s string) {
		// String values
		v := slog.StringValue(s)
		_ = ValueToString(v)

		// Any values
		v = slog.AnyValue(s)
		_ = ValueToString(v)
	})
}

func FuzzAnyValueToString(f *testing.F) {
	f.Add("test")
	f.Add("")
	f.Add("unicode: 日本語")

	f.Fuzz(func(t *testing.T, s string) {
		v := slog.AnyValue(s)
		_ = AnyValueToString(v)
	})
}

func FuzzFindAttribute(f *testing.F) {
	f.Add("target", "key1", "key2", "key3")
	f.Add("missing", "a", "b", "c")
	f.Add("", "", "", "")

	f.Fuzz(func(t *testing.T, target, k1, k2, k3 string) {
		attrs := []slog.Attr{
			slog.String(k1, "v1"),
			slog.String(k2, "v2"),
			slog.String(k3, "v3"),
		}
		// Should not panic
		_, _ = FindAttribute(attrs, []string{}, target)
	})
}

func FuzzRemoveEmptyAttrs(f *testing.F) {
	f.Add("key1", "val1", "", "val2", "key3", "")

	f.Fuzz(func(t *testing.T, k1, v1, k2, v2, k3, v3 string) {
		attrs := []slog.Attr{
			slog.String(k1, v1),
			slog.String(k2, v2),
			slog.String(k3, v3),
		}
		// Should not panic
		_ = RemoveEmptyAttrs(attrs)
	})
}

func FuzzUniqAttrs(f *testing.F) {
	f.Add("a", "b", "a", "c")
	f.Add("", "", "", "")
	f.Add("x", "x", "x", "x")

	f.Fuzz(func(t *testing.T, k1, k2, k3, k4 string) {
		attrs := []slog.Attr{
			slog.String(k1, "v1"),
			slog.String(k2, "v2"),
			slog.String(k3, "v3"),
			slog.String(k4, "v4"),
		}
		result := UniqAttrs(attrs)
		// Result should never be longer than input
		if len(result) > len(attrs) {
			t.Errorf("UniqAttrs returned %d attrs from %d input", len(result), len(attrs))
		}
	})
}

func FuzzAttrsToString(f *testing.F) {
	f.Add("key", "value")
	f.Add("", "")

	f.Fuzz(func(t *testing.T, k, v string) {
		attrs := []slog.Attr{
			slog.String(k, v),
			slog.Int(k+"_int", 42),
			slog.Duration(k+"_dur", time.Second),
		}
		// Should not panic
		_ = AttrsToString(attrs...)
	})
}
