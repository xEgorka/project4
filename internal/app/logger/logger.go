// Package logger implements logger singleton based on zap.
package logger

import "go.uber.org/zap"

// Log sets up default no-op-logger which prints no messages.
var Log *zap.Logger = zap.NewNop()

// Initialize creates logger singleton on specified logging level.
func Initialize(level string) error {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}
	cfg := zap.NewDevelopmentConfig()
	cfg.Level = lvl
	zl, err := cfg.Build()
	if err != nil {
		return err
	}
	// set singleton
	Log = zl
	return nil
}
