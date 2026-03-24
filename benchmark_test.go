package slogcommon

import (
	"fmt"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func BenchmarkAttrsToMap(b *testing.B) {
	b.Run("small/3-attrs", func(b *testing.B) {
		attrs := []slog.Attr{
			slog.String("key1", "value1"),
			slog.Int("key2", 42),
			slog.Bool("key3", true),
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			AttrsToMap(attrs...)
		}
	})

	b.Run("medium/20-attrs", func(b *testing.B) {
		attrs := make([]slog.Attr, 20)
		for i := range attrs {
			attrs[i] = slog.String(fmt.Sprintf("key%d", i), fmt.Sprintf("value%d", i))
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			AttrsToMap(attrs...)
		}
	})

	b.Run("with-groups", func(b *testing.B) {
		attrs := []slog.Attr{
			slog.String("key1", "value1"),
			slog.Group("group1",
				slog.String("nested1", "val1"),
				slog.String("nested2", "val2"),
			),
			slog.Group("group1",
				slog.String("nested3", "val3"),
			),
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			AttrsToMap(attrs...)
		}
	})
}

func BenchmarkReplaceAttrs(b *testing.B) {
	replaceFn := func(groups []string, a slog.Attr) slog.Attr {
		return a
	}

	b.Run("flat/5-attrs", func(b *testing.B) {
		attrs := []slog.Attr{
			slog.String("a", "1"), slog.Int("b", 2), slog.Bool("c", true),
			slog.Float64("d", 3.14), slog.Duration("e", time.Second),
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ReplaceAttrs(replaceFn, []string{"group"}, attrs...)
		}
	})

	b.Run("nested-groups", func(b *testing.B) {
		attrs := []slog.Attr{
			slog.String("a", "1"),
			slog.Group("g1", slog.Group("g2", slog.String("deep", "val"))),
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ReplaceAttrs(replaceFn, []string{"root"}, attrs...)
		}
	})
}

func BenchmarkAppendRecordAttrsToAttrs(b *testing.B) {
	for _, n := range []int{1, 5, 20} {
		b.Run(fmt.Sprintf("record-%d-attrs", n), func(b *testing.B) {
			baseAttrs := []slog.Attr{slog.String("base", "val")}
			r := slog.NewRecord(time.Now(), slog.LevelInfo, "msg", 0)
			for i := 0; i < n; i++ {
				r.AddAttrs(slog.String(fmt.Sprintf("key%d", i), fmt.Sprintf("val%d", i)))
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				AppendRecordAttrsToAttrs(baseAttrs, []string{}, &r)
			}
		})
	}
}

func BenchmarkRemoveEmptyAttrs(b *testing.B) {
	b.Run("no-empty", func(b *testing.B) {
		attrs := []slog.Attr{
			slog.String("a", "1"), slog.Int("b", 2), slog.Bool("c", true),
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			RemoveEmptyAttrs(attrs)
		}
	})

	b.Run("with-empty", func(b *testing.B) {
		attrs := []slog.Attr{
			slog.String("a", "1"),
			slog.Any("empty", nil),
			slog.Group("g", slog.Any("nested_empty", nil)),
			slog.Int("b", 2),
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			RemoveEmptyAttrs(attrs)
		}
	})
}

func BenchmarkFindAttribute(b *testing.B) {
	attrs := []slog.Attr{
		slog.String("a", "1"),
		slog.String("b", "2"),
		slog.Group("g1", slog.String("c", "3"), slog.Group("g2", slog.String("d", "4"))),
		slog.String("e", "5"),
	}

	b.Run("flat-found", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			FindAttribute(attrs, []string{}, "e")
		}
	})

	b.Run("flat-not-found", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			FindAttribute(attrs, []string{}, "z")
		}
	})

	b.Run("nested", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			FindAttribute(attrs, []string{"g1", "g2"}, "d")
		}
	})
}

func BenchmarkAppendAttrsToGroup(b *testing.B) {
	b.Run("no-groups", func(b *testing.B) {
		actual := []slog.Attr{slog.String("a", "1")}
		newAttrs := []slog.Attr{slog.String("b", "2")}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			AppendAttrsToGroup([]string{}, actual, newAttrs...)
		}
	})

	b.Run("nested-groups", func(b *testing.B) {
		actual := []slog.Attr{
			slog.Group("g1", slog.Group("g2", slog.String("a", "1"))),
		}
		newAttrs := []slog.Attr{slog.String("b", "2")}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			AppendAttrsToGroup([]string{"g1", "g2"}, actual, newAttrs...)
		}
	})
}

func BenchmarkUniqAttrs(b *testing.B) {
	b.Run("no-duplicates", func(b *testing.B) {
		attrs := []slog.Attr{
			slog.String("a", "1"), slog.String("b", "2"), slog.String("c", "3"),
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			UniqAttrs(attrs)
		}
	})

	b.Run("with-duplicates", func(b *testing.B) {
		attrs := []slog.Attr{
			slog.String("a", "1"), slog.String("b", "2"), slog.String("a", "3"),
			slog.String("c", "4"), slog.String("b", "5"),
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			UniqAttrs(attrs)
		}
	})
}

func BenchmarkFormatError(b *testing.B) {
	b.Run("standard-error", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			FormatError(assert.AnError)
		}
	})

	b.Run("log-valuer-error", func(b *testing.B) {
		err := &testError{}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			FormatError(err)
		}
	})
}

func BenchmarkAttrsToString(b *testing.B) {
	attrs := []slog.Attr{
		slog.String("name", "alice"),
		slog.Int("age", 30),
		slog.Bool("active", true),
		slog.Duration("elapsed", 5*time.Second),
		slog.Time("created", time.Now()),
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		AttrsToString(attrs...)
	}
}

func BenchmarkFormatRequest(b *testing.B) {
	req, _ := http.NewRequest("GET", "https://example.com/path?foo=bar&baz=qux", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer token")

	b.Run("with-headers", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			FormatRequest(req, false)
		}
	})

	b.Run("without-headers", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			FormatRequest(req, true)
		}
	})
}
