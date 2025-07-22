// Copyright (C) 2025 stencil contributors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package slogext is a small wrapper around the [log/slog] package
// focused on providing consistency in logging across the stencil
// codebase.
package slogext

import (
	"fmt"
	"io"
	"log/slog"
	"os"

	charmlog "github.com/charmbracelet/log"
)

// _ ensures that the logger struct satisfies the Logger interface.
var _ Logger = &logger{}

// Logger is a [slog.Logger] interface with extra functionality.
type Logger interface {
	Info(string, ...any)
	Infof(string, ...any)
	Debug(string, ...any)
	Debugf(string, ...any)
	Error(string, ...any)
	Errorf(string, ...any)
	Warn(string, ...any)
	Warnf(string, ...any)
	With(...any) Logger
	WithError(error) Logger
	SetLevel(charmlog.Level)
	GetHandler() slog.Handler
}

// Level is a logging level.
type Level = charmlog.Level

// Contains valid [Level] values.
const (
	// DebugLevel is for debug level logs.
	DebugLevel Level = charmlog.DebugLevel

	// InfoLevel is for info level logs.
	InfoLevel Level = charmlog.InfoLevel

	// WarnLevel is for warn level logs.
	WarnLevel Level = charmlog.WarnLevel

	// ErrorLevel is for error level logs
	ErrorLevel Level = charmlog.ErrorLevel

	// FatalLevel is for fatal level logs.
	FatalLevel Level = charmlog.FatalLevel
)

// New creates a new [Logger] using the slog package.
func New() Logger {
	return NewWithWriter(os.Stdout)
}

// NewWithWriter creates a new [Logger] using the slog package with a
// custom writer target.
func NewWithWriter(w io.Writer) Logger {
	handler := charmlog.New(w)
	return &logger{slog.New(handler), handler}
}

// NewWithHandler creates a new [Logger] using the slog package. Levels
// do not work since it is not a slog native concept.
func NewWithHandler(h slog.Handler) Logger {
	// TODO(jaredallard): Consider exposing a leveler function.
	return &logger{slog.New(h), charmlog.New(io.Discard)}
}

// logger is a simple wrapper around the slog.Logger interface. Use
// [Logger] when passing around loggers in the stencil codebase.
type logger struct {
	*slog.Logger
	handler *charmlog.Logger
}

// With wraps the slog.With method to return a new logger with the
// provided arguments while satisfying the Logger interface.
func (l *logger) With(args ...any) Logger {
	return &logger{l.Logger.With(args...), l.handler}
}

// WithError wraps the slog.With method using a consistent key for
// errors, "error".
func (l *logger) WithError(err error) Logger {
	return &logger{l.Logger.With("error", err), l.handler}
}

// SetLevel updates the level of the current logger to the provided
// level.
func (l *logger) SetLevel(level Level) {
	l.handler.SetLevel(level)
}

// Infof wraps Info with a formatted message.
func (l *logger) Infof(format string, args ...any) {
	l.Info(fmt.Sprintf(format, args...))
}

// Debugf wraps Debug with a formatted message.
func (l *logger) Debugf(format string, args ...any) {
	l.Debug(fmt.Sprintf(format, args...))
}

// Errorf wraps Error with a formatted message.
func (l *logger) Errorf(format string, args ...any) {
	l.Error(fmt.Sprintf(format, args...))
}

// Warnf wraps Warn with a formatted message.
func (l *logger) Warnf(format string, args ...any) {
	l.Warn(fmt.Sprintf(format, args...))
}

// GetHandler returns the [slog.Handler] used by this logger, for
// interoperability purposes.
func (l *logger) GetHandler() slog.Handler {
	return l.handler
}
