package issuer

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/cfn"
	"github.com/aws/aws-lambda-go/lambdacontext"
)

func HandlerWithLambdaLogging[E, R any](handler func(context.Context, E) (R, error)) func(context.Context, E) (R, error) {
	var level slog.Level
	switch os.Getenv("AWS_LAMBDA_LOG_LEVEL") {
	case "DEBUG":
		level = slog.LevelDebug
	case "INFO":
		level = slog.LevelInfo
	case "WARN":
		level = slog.LevelWarn
	case "ERROR":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	return func(ctx context.Context, event E) (R, error) {
		lc, _ := lambdacontext.FromContext(ctx)
		logHandler := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})).With("requestId", lc.AwsRequestID)
		slog.SetDefault(logHandler)

		return handler(ctx, event)
	}
}

func CloudFormationHandlerWithLambdaLogging(handler func(context.Context, cfn.Event) (string, map[string]any, error)) func(context.Context, cfn.Event) (string, map[string]any, error) {
	var level slog.Level
	switch os.Getenv("AWS_LAMBDA_LOG_LEVEL") {
	case "DEBUG":
		level = slog.LevelDebug
	case "INFO":
		level = slog.LevelInfo
	case "WARN":
		level = slog.LevelWarn
	case "ERROR":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	return func(ctx context.Context, event cfn.Event) (string, map[string]any, error) {
		lc, _ := lambdacontext.FromContext(ctx)

		alwaysLogAttrs := []any{
			"requestId", lc.AwsRequestID,
			"functionName", os.Getenv("AWS_LAMBDA_FUNCTION_NAME"),
		}
		logHandler := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})).With(alwaysLogAttrs...)
		slog.SetDefault(logHandler)

		return handler(ctx, event)
	}
}

func LogWithTiming(ctx context.Context, level slog.Level, message string, args ...any) func() {
	slog.Log(ctx, level, message, args...)
	start := time.Now()

	return func() {
		duration := time.Since(start)
		durationMs := float64(duration) / float64(time.Millisecond)
		slog.Log(ctx, level, "timing:"+message, "duration", duration.String(), "durationMs", durationMs)
	}
}
