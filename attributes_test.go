package slogcommon

import (
	"log/slog"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSource(t *testing.T) {
	// Simulate a runtime frame
	pc, file, _, _ := runtime.Caller(0)
	record := &slog.Record{PC: pc}

	// Call the source function
	attr := Source("sourceKey", record)

	// Assert the attributes
	assert.Equal(t, "sourceKey", attr.Key)
	assert.Equal(t, slog.KindGroup, attr.Value.Kind())

	groupAttrs := attr.Value.Group()

	expectedAttrs := map[string]any{
		"function": "github.com/samber/slog-common.TestSource",
		"file":     file,
		"line":     int64(14),
	}

	for _, a := range groupAttrs {
		expectedValue, ok := expectedAttrs[a.Key]
		if assert.True(t, ok, "unexpected attribute key: %s", a.Key) {
			assert.Equal(t, expectedValue, a.Value.Any())
		}
	}
}

func TestAppendRecordAttrsToAttrs(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	a := slog.String("a", "1")
	b := slog.Int("b", 2)
	c := slog.Bool("c", true)
	d := slog.Float64("d", 3.14)
	e := slog.Duration("e", time.Second)
	r := slog.NewRecord(time.Now(), slog.LevelError, assert.AnError.Error(), 0)
	r.AddAttrs(d, e)

	output := AppendRecordAttrsToAttrs([]slog.Attr{a, b, c}, []string{}, &r)
	is.Len(output, 5)
	is.Equal([]slog.Attr{a, b, c, d, e}, output)

	output = AppendRecordAttrsToAttrs([]slog.Attr{a, b, c}, []string{"foo", "bar"}, &r)
	is.Len(output, 5)
	is.Equal([]slog.Attr{a, b, c, slog.Group("foo", slog.Group("bar", d)), slog.Group("foo", slog.Group("bar", e))}, output)

	r = slog.NewRecord(time.Now(), slog.LevelError, assert.AnError.Error(), 0)

	output = AppendRecordAttrsToAttrs([]slog.Attr{a, b, c}, []string{}, &r)
	is.Len(output, 3)
	is.Equal([]slog.Attr{a, b, c}, output)
}

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

type testMaskedLogValuer string

func (t testMaskedLogValuer) LogValue() slog.Value {
	return slog.StringValue("********") // LogValue does not always return GroupValue
}

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
	is.Equal(
		[]slog.Attr{slog.String("password", "********")},
		ReplaceAttrs(
			nil,
			[]string{"foobar"},
			slog.Any("password", testMaskedLogValuer("password")),
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
	is.Equal(
		[]slog.Attr{slog.Bool("bool", false), slog.Int("int", 42)},
		RemoveEmptyAttrs(
			[]slog.Attr{slog.Bool("bool", false), slog.Int("int", 42)},
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
	is.Equal(
		[]slog.Attr{slog.Group("test", slog.Int("int", 42))},
		RemoveEmptyAttrs(
			[]slog.Attr{slog.Group("test", slog.Int("int", 42), slog.Any("foobar", nil))},
		),
	)

	// do not remove empty group name with non-empty values
	is.Equal(
		[]slog.Attr{slog.Int("int", 42)},
		RemoveEmptyAttrs(
			[]slog.Attr{slog.Group("", slog.Int("int", 42))},
		),
	)
}

type textMarshalerExample struct {
	Data string
}

func (t textMarshalerExample) MarshalText() (text []byte, err error) {
	return []byte(t.Data), nil
}

type nonTextMarshalerExample struct {
	Data string
}

func TestAnyValueToString(t *testing.T) {
	tests := map[string]struct {
		input    slog.Attr
		expected string
	}{
		"TextMarshaler implementation": {
			input:    slog.Any("key", textMarshalerExample{Data: "example"}),
			expected: "example",
		},
		"Non-TextMarshaler implementation": {
			input:    slog.Any("key", nonTextMarshalerExample{Data: "example"}),
			expected: "{Data:example}",
		},
		"String value": {
			input:    slog.String("key", "example string"),
			expected: "example string",
		},
		"Integer value": {
			input:    slog.Int("key", 42),
			expected: "42",
		},
		"Boolean value": {
			input:    slog.Bool("key", true),
			expected: "true",
		},
		"Nil value": {
			input:    slog.Any("key", nil),
			expected: "<nil>",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			output := AnyValueToString(tt.input.Value)
			if output != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, output)
			}
		})
	}
}

func TestValueToString(t *testing.T) {
	tests := map[string]struct {
		input    slog.Attr
		expected string
	}{
		"KindInt64": {
			input:    slog.Int64("key", 42),
			expected: "42",
		},
		"KindUint64": {
			input:    slog.Uint64("key", 42),
			expected: "42",
		},
		"KindFloat64": {
			input:    slog.Float64("key", 3.14),
			expected: "3.14",
		},
		"KindString": {
			input:    slog.String("key", "test"),
			expected: "test",
		},
		"KindBool": {
			input:    slog.Bool("key", true),
			expected: "true",
		},
		"KindDuration": {
			input:    slog.Duration("key", time.Second*42),
			expected: "42s",
		},
		"KindTime": {
			input:    slog.Time("key", time.Date(2023, time.July, 30, 12, 0, 0, 0, time.UTC)),
			expected: "2023-07-30 12:00:00 +0000 UTC",
		},
		"KindAny": {
			input:    slog.Any("key", "any value"),
			expected: "any value",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			actual := ValueToString(tc.input.Value)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
