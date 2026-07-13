package issuer

import (
	"context"
	"log/slog"
	"time"
)

func LogWithTiming(ctx context.Context, level slog.Level, message string, args ...any) func() {
	slog.Log(ctx, level, message, args...)
	start := time.Now()

	return func() {
		duration := time.Since(start)
		durationMs := float64(duration) / float64(time.Millisecond)
		slog.Log(ctx, level, "timing:"+message, "duration", duration.String(), "durationMs", durationMs)
	}
}
