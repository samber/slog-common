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
