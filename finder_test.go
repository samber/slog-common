package slogcommon

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindAttrByKey(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		attrs    []slog.Attr
		key      string
		expected slog.Attr
		found    bool
	}{
		"Found": {
			attrs:    []slog.Attr{slog.String("a", "1"), slog.Int("b", 2)},
			key:      "b",
			expected: slog.Int("b", 2),
			found:    true,
		},
		"NotFound": {
			attrs:    []slog.Attr{slog.String("a", "1"), slog.Int("b", 2)},
			key:      "c",
			expected: slog.Attr{},
			found:    false,
		},
		"EmptySlice": {
			attrs:    []slog.Attr{},
			key:      "a",
			expected: slog.Attr{},
			found:    false,
		},
		"FirstMatch": {
			attrs:    []slog.Attr{slog.String("a", "first"), slog.String("a", "second")},
			key:      "a",
			expected: slog.String("a", "first"),
			found:    true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			is := assert.New(t)
			attr, ok := FindAttrByKey(tt.attrs, tt.key)
			is.Equal(tt.found, ok)
			is.Equal(tt.expected, attr)
		})
	}
}

func TestFindAttrByGroupAndKey(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		attrs    []slog.Attr
		groups   []string
		key      string
		expected slog.Attr
		found    bool
	}{
		"NoGroups_Found": {
			attrs:    []slog.Attr{slog.String("a", "1"), slog.Int("b", 2)},
			groups:   []string{},
			key:      "a",
			expected: slog.String("a", "1"),
			found:    true,
		},
		"NoGroups_NotFound": {
			attrs:    []slog.Attr{slog.String("a", "1")},
			groups:   []string{},
			key:      "z",
			expected: slog.Attr{},
			found:    false,
		},
		"WithGroup_Found": {
			attrs: []slog.Attr{
				slog.Group("grp", slog.String("inner", "val")),
			},
			groups:   []string{"grp"},
			key:      "inner",
			expected: slog.String("inner", "val"),
			found:    true,
		},
		"WithGroup_NotFound": {
			attrs: []slog.Attr{
				slog.Group("grp", slog.String("inner", "val")),
			},
			groups:   []string{"grp"},
			key:      "missing",
			expected: slog.Attr{},
			found:    false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			is := assert.New(t)
			attr, ok := FindAttrByGroupAndKey(tt.attrs, tt.groups, tt.key)
			is.Equal(tt.found, ok)
			is.Equal(tt.expected, attr)
		})
	}
}
