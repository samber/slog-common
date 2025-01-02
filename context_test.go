package slogcommon

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

type ctxKey string

func TestContextExtractor(t *testing.T) {
	tests := map[string]struct {
		ctx      context.Context
		fns      []func(ctx context.Context) []slog.Attr
		expected []slog.Attr
	}{
		"NoFunctions": {
			ctx:      context.Background(),
			fns:      []func(ctx context.Context) []slog.Attr{},
			expected: []slog.Attr{},
		},
		"SingleFunction": {
			ctx: context.Background(),
			fns: []func(ctx context.Context) []slog.Attr{
				func(ctx context.Context) []slog.Attr {
					return []slog.Attr{slog.String("key1", "value1")}
				},
			},
			expected: []slog.Attr{slog.String("key1", "value1")},
		},
		"MultipleFunctions": {
			ctx: context.Background(),
			fns: []func(ctx context.Context) []slog.Attr{
				func(ctx context.Context) []slog.Attr {
					return []slog.Attr{slog.String("key1", "value1")}
				},
				func(ctx context.Context) []slog.Attr {
					return []slog.Attr{slog.String("key2", "value2")}
				},
			},
			expected: []slog.Attr{slog.String("key1", "value1"), slog.String("key2", "value2")},
		},
		"FunctionWithContext": {
			ctx: context.WithValue(context.Background(), ctxKey("userID"), "1234"),
			fns: []func(ctx context.Context) []slog.Attr{
				func(ctx context.Context) []slog.Attr {
					if userID, ok := ctx.Value(ctxKey("userID")).(string); ok {
						return []slog.Attr{slog.String("userID", userID)}
					}
					return []slog.Attr{}
				},
			},
			expected: []slog.Attr{slog.String("userID", "1234")},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			actual := ContextExtractor(tc.ctx, tc.fns)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
