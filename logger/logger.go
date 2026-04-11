package logger

import (
	"log/slog"
	"os"
)

var Log *slog.Logger

func Init(profile string) {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	handler := slog.NewJSONHandler(os.Stdout, opts).WithAttrs([]slog.Attr{
		slog.String("profile", profile),
	})
	Log = slog.New(handler)
	slog.SetDefault(Log)
}
