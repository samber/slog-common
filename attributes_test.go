package slogcommon

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReplaceAttrs(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	// no ReplaceAttr func
	is.Equal(
		[]slog.Attr{slog.Bool("bool", true), slog.Int("int", 42)},
		ReplaceAttrs(
			nil,
			[]string{"foobar"},
			slog.Bool("bool", true), slog.Int("int", 42),
		),
	)

	// ReplaceAttr func, but returns the same attributes
	is.Equal(
		[]slog.Attr{slog.Bool("bool", true), slog.Int("int", 42)},
		ReplaceAttrs(
			func(groups []string, a slog.Attr) slog.Attr {
				is.Equal("foobar", groups[0])
				return a
			},
			[]string{"foobar"},
			slog.Bool("bool", true), slog.Int("int", 42),
		),
	)

	// Replace int and divide by 2
	is.Equal(
		[]slog.Attr{slog.Bool("bool", true), slog.Int("int", 21)},
		ReplaceAttrs(
			func(groups []string, a slog.Attr) slog.Attr {
				is.Equal("foobar", groups[0])
				if a.Value.Kind() == slog.KindInt64 {
					a.Value = slog.Int64Value(a.Value.Int64() / 2)
				}
				return a
			},
			[]string{"foobar"},
			slog.Bool("bool", true), slog.Int("int", 42),
		),
	)

	// Remove int attr
	is.Equal(
		[]slog.Attr{slog.Bool("bool", true), slog.Any("int", nil)},
		ReplaceAttrs(
			func(groups []string, a slog.Attr) slog.Attr {
				is.Equal("foobar", groups[0])
				if a.Value.Kind() == slog.KindInt64 {
					return slog.Any("int", nil)
				}
				return a
			},
			[]string{"foobar"},
			slog.Bool("bool", true), slog.Int("int", 42),
		),
	)

	// Rename int attr
	is.Equal(
		[]slog.Attr{slog.Bool("bool", true), slog.Int("int2", 21)},
		ReplaceAttrs(
			func(groups []string, a slog.Attr) slog.Attr {
				is.Equal("foobar", groups[0])
				if a.Value.Kind() == slog.KindInt64 {
					return slog.Int("int2", 21)
				}
				return a
			},
			[]string{"foobar"},
			slog.Bool("bool", true), slog.Int("int", 42),
		),
	)

	// Rename attr in groups
	is.Equal(
		[]slog.Attr{slog.Bool("bool", true), slog.Group("group1", slog.Group("group2", slog.Int("int", 21)))},
		ReplaceAttrs(
			func(groups []string, a slog.Attr) slog.Attr {
				is.Equal("foobar", groups[0])
				if len(groups) > 1 {
					is.Equal([]string{"foobar", "group1", "group2"}, groups)
					return slog.Int("int", 21)
				}
				return a
			},
			[]string{"foobar"},
			slog.Bool("bool", true), slog.Group("group1", slog.Group("group2", slog.String("string", "foobar"))),
		),
	)
}

func TestAttrsToMap(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	// simple
	is.EqualValues(
		map[string]any{"key": "value"},
		AttrsToMap(slog.Any("key", "value")),
	)

	// nested
	is.EqualValues(
		map[string]any{"key": "value", "key1": map[string]any{"key2": "value2"}},
		AttrsToMap(slog.Any("key", "value"), slog.Group("key1", slog.Any("key2", "value2"))),
	)

	// merge
	is.EqualValues(
		map[string]any{"key": "value", "key1": map[string]any{"key2": "value2", "key3": "value3"}},
		AttrsToMap(
			slog.Any("key", "value"),
			slog.Group("key1", slog.Any("key2", "value2")),
			slog.Group("key1", slog.Any("key3", "value3")),
		),
	)

}

func TestFindAttribute(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	// simple
	attr, ok := FindAttribute(
		[]slog.Attr{
			slog.Any("key", "value"),
			slog.Group("key1", slog.Any("key2", "value2")),
		},
		[]string{},
		"key1",
	)
	is.True(ok)
	is.EqualValues(slog.Group("key1", slog.Any("key2", "value2")), attr)

	// return first
	attr, ok = FindAttribute(
		[]slog.Attr{
			slog.Any("key", "value"),
			slog.Group("key1", slog.Any("key2", "value2")),
			slog.Group("key1", slog.Any("key3", "value3")),
		},
		[]string{},
		"key1",
	)
	is.True(ok)
	is.EqualValues(slog.Group("key1", slog.Any("key2", "value2")), attr)

	// not found
	attr, ok = FindAttribute(
		[]slog.Attr{
			slog.Any("key", "value"),
			slog.Group("key1", slog.Any("key2", "value2")),
			slog.Group("key1", slog.Any("key3", "value3")),
		},
		[]string{},
		"key2",
	)
	is.False(ok)
	is.EqualValues(slog.Attr{}, attr)

	// nested
	attr, ok = FindAttribute(
		[]slog.Attr{
			slog.Any("key", "value"),
			slog.Group("key1", slog.Any("key2", "value2")),
			slog.Group("key1", slog.Any("key3", "value3")),
		},
		[]string{"key1"},
		"key2",
	)
	is.True(ok)
	is.EqualValues(slog.Any("key2", "value2"), attr)

	// nested second
	attr, ok = FindAttribute(
		[]slog.Attr{
			slog.Any("key", "value"),
			slog.Group("key1", slog.Any("key2", "value2")),
			slog.Group("key1", slog.Any("key3", "value3")),
		},
		[]string{"key1"},
		"key3",
	)
	is.True(ok)
	is.EqualValues(slog.Any("key3", "value3"), attr)

	// too nested
	attr, ok = FindAttribute(
		[]slog.Attr{
			slog.Any("key", "value"),
			slog.Group("key1", slog.Any("key2", "value2")),
			slog.Group("key1", slog.Any("key3", "value3")),
		},
		[]string{"key1", "key2"},
		"key3",
	)
	is.False(ok)
	is.EqualValues(slog.Attr{}, attr)

	// nested not found
	attr, ok = FindAttribute(
		[]slog.Attr{
			slog.Any("key", "value"),
			slog.Group("key1", slog.Any("key2", "value2")),
			slog.Group("key1", slog.Any("key3", "value3")),
		},
		[]string{"key4"},
		"key1",
	)
	is.False(ok)
	is.EqualValues(slog.Attr{}, attr)
}
