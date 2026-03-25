package slogcommon

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppendAttrsToGroup(t *testing.T) {
	tests := map[string]struct {
		groups      []string
		actualAttrs []slog.Attr
		newAttrs    []slog.Attr
		expected    []slog.Attr
	}{
		"NoGroups": {
			groups:      []string{},
			actualAttrs: []slog.Attr{slog.String("key1", "value1")},
			newAttrs:    []slog.Attr{slog.String("key2", "value2")},
			expected:    []slog.Attr{slog.String("key1", "value1"), slog.String("key2", "value2")},
		},
		"SingleGroup": {
			groups: []string{"group1"},
			actualAttrs: []slog.Attr{
				slog.Group("group1", slog.String("key1", "value1")),
			},
			newAttrs: []slog.Attr{slog.String("key2", "value2")},
			expected: []slog.Attr{
				slog.Group("group1", slog.String("key1", "value1"), slog.String("key2", "value2")),
			},
		},
		"NestedGroups": {
			groups: []string{"group1", "group2"},
			actualAttrs: []slog.Attr{
				slog.Group("group1", slog.Group("group2", slog.String("key1", "value1"))),
			},
			newAttrs: []slog.Attr{slog.String("key2", "value2")},
			expected: []slog.Attr{
				slog.Group("group1", slog.Group("group2", slog.String("key1", "value1"), slog.String("key2", "value2"))),
			},
		},
		"NewGroup": {
			groups:      []string{"group1"},
			actualAttrs: []slog.Attr{slog.String("key1", "value1")},
			newAttrs:    []slog.Attr{slog.String("key2", "value2")},
			expected: []slog.Attr{
				slog.String("key1", "value1"),
				slog.Group("group1", slog.String("key2", "value2")),
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			actual := AppendAttrsToGroup(tc.groups, tc.actualAttrs, tc.newAttrs...)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestUniqAttrs(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input    []slog.Attr
		expected []slog.Attr
	}{
		"NoDuplicates": {
			input:    []slog.Attr{slog.String("a", "1"), slog.Int("b", 2)},
			expected: []slog.Attr{slog.String("a", "1"), slog.Int("b", 2)},
		},
		"WithDuplicates_KeepsLast": {
			input:    []slog.Attr{slog.String("a", "first"), slog.Int("b", 2), slog.String("a", "last")},
			expected: []slog.Attr{slog.String("a", "last"), slog.Int("b", 2)},
		},
		"Empty": {
			input:    []slog.Attr{},
			expected: []slog.Attr{},
		},
		"AllSameKey": {
			input:    []slog.Attr{slog.String("a", "1"), slog.String("a", "2"), slog.String("a", "3")},
			expected: []slog.Attr{slog.String("a", "3")},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			is := assert.New(t)
			result := UniqAttrs(tt.input)
			is.Equal(tt.expected, result)
		})
	}
}
