package slogcommon

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testLogValuer struct {
	name string
	pass string
}

func (t testLogValuer) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("name", t.name),
		slog.String("password", "********"),
	)
}

var stubLogValuer = testLogValuer{"userName", "password"}

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

	// no ReplaceAttr func, but convert struct with interface slog.LogValue in slog.Group
	is.Equal(
		[]slog.Attr{slog.Group("user", slog.String("name", stubLogValuer.name), slog.String("password", "********"))},
		ReplaceAttrs(
			nil,
			[]string{"foobar"},
			slog.Any("user", stubLogValuer),
		),
	)

	// ReplaceAttr func, but convert struct with interface slog.LogValue in slog.Group
	is.Equal(
		[]slog.Attr{slog.Group("user", slog.String("name", stubLogValuer.name), slog.String("password", "********"))},
		ReplaceAttrs(
			func(groups []string, a slog.Attr) slog.Attr {
				is.Equal([]string{"foobar", "user"}, groups)
				return a
			},
			[]string{"foobar"},
			slog.Any("user", stubLogValuer),
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

func TestExtractError(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	// not found
	attrs, err := ExtractError(
		[]slog.Attr{
			slog.Any("key", "value"),
			slog.Group("key1", slog.Any("key2", "value2")),
			slog.String("foo", "bar"),
		},
		"kerrorey1",
	)
	is.Len(attrs, 3)
	is.Nil(err)

	// found key but wrong type
	attrs, err = ExtractError(
		[]slog.Attr{
			slog.Any("key", "value"),
			slog.Group("key1", slog.Any("key2", "value2")),
			slog.String("error", "bar"),
		},
		"error",
	)
	is.Len(attrs, 3)
	is.Nil(err)

	// found start first key
	attrs, err = ExtractError(
		[]slog.Attr{
			slog.Any("error", assert.AnError),
			slog.Any("key", "value"),
			slog.Group("key1", slog.Any("key2", "value2")),
			slog.String("foo", "bar"),
		},
		"error",
		"err",
	)
	is.Len(attrs, 3)
	is.EqualError(err, assert.AnError.Error())

	// found start second key
	attrs, err = ExtractError(
		[]slog.Attr{
			slog.Any("err", assert.AnError),
			slog.Any("key", "value"),
			slog.Group("key1", slog.Any("key2", "value2")),
			slog.String("foo", "bar"),
		},
		"error",
		"err",
	)
	is.Len(attrs, 3)
	is.EqualError(err, assert.AnError.Error())

	// found middle
	attrs, err = ExtractError(
		[]slog.Attr{
			slog.Any("key", "value"),
			slog.Any("error", assert.AnError),
			slog.Group("key1", slog.Any("key2", "value2")),
			slog.String("foo", "bar"),
		},
		"error",
	)
	is.Len(attrs, 3)
	is.EqualError(err, assert.AnError.Error())

	// found end
	attrs, err = ExtractError(
		[]slog.Attr{
			slog.Any("key", "value"),
			slog.Group("key1", slog.Any("key2", "value2")),
			slog.String("foo", "bar"),
			slog.Any("error", assert.AnError),
		},
		"error",
	)
	is.Len(attrs, 3)
	is.EqualError(err, assert.AnError.Error())
}

func TestRemoveEmptyAttrs(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	// do not remove anything
	is.Equal(
		[]slog.Attr{slog.Bool("bool", true), slog.Int("int", 42)},
		RemoveEmptyAttrs(
			[]slog.Attr{slog.Bool("bool", true), slog.Int("int", 42)},
		),
	)

	// remove if missing keys
	is.Equal(
		[]slog.Attr{slog.Int("int", 42)},
		RemoveEmptyAttrs(
			[]slog.Attr{slog.Bool("", true), slog.Int("int", 42)},
		),
	)

	// remove if missing value
	is.Equal(
		[]slog.Attr{slog.Int("int", 42)},
		RemoveEmptyAttrs(
			[]slog.Attr{slog.Any("test", nil), slog.Int("int", 42)},
		),
	)
	is.Equal(
		[]slog.Attr{slog.Int("int", 42)},
		RemoveEmptyAttrs(
			[]slog.Attr{slog.Group("test"), slog.Int("int", 42)},
		),
	)

	// remove nested
	is.Equal(
		[]slog.Attr{slog.Int("int", 42)},
		RemoveEmptyAttrs(
			[]slog.Attr{slog.Any("test", nil), slog.Int("int", 42)},
		),
	)
	is.Equal(
		[]slog.Attr{slog.Int("int", 42)},
		RemoveEmptyAttrs(
			[]slog.Attr{slog.Group("test", slog.Any("foobar", nil)), slog.Int("int", 42)},
		),
	)
}
