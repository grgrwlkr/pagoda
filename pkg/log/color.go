package log

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"runtime"
)

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorWhite  = "\033[37m"
)

// ColorHandler is a custom slog handler that adds color to log output
type ColorHandler struct {
	w                 io.Writer
	opts              *slog.HandlerOptions
	preformattedAttrs []slog.Attr
	groups            []string
}

// NewColorHandler creates a new color handler
func NewColorHandler(w io.Writer, opts *slog.HandlerOptions) *ColorHandler {
	if opts == nil {
		opts = &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}
	}
	return &ColorHandler{
		w:    w,
		opts: opts,
	}
}

// Enabled reports whether the handler handles records at the given level
func (h *ColorHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.opts.Level.Level()
}

// Handle handles the record
func (h *ColorHandler) Handle(ctx context.Context, r slog.Record) error {
	// Format time
	timeStr := r.Time.Format("2006-01-02 15:04:05")

	// Get level string and color
	levelStr := r.Level.String()
	var levelColor string
	switch r.Level {
	case slog.LevelError:
		levelColor = colorRed
	case slog.LevelWarn:
		levelColor = colorYellow
	default:
		levelColor = colorWhite
	}

	// Write time and colored level
	fmt.Fprintf(h.w, "%s %s%s%s", timeStr, levelColor, levelStr, colorReset)

	// Write message in white
	if r.Message != "" {
		fmt.Fprintf(h.w, " %s%s%s", colorWhite, r.Message, colorReset)
	}

	// Write preformatted attributes (from WithAttrs)
	for _, a := range h.preformattedAttrs {
		h.writeAttr(a)
	}

	// Write record attributes with yellow keys and white values
	r.Attrs(func(a slog.Attr) bool {
		h.writeAttr(a)
		return true
	})

	// Write source if enabled
	if h.opts.AddSource && r.PC != 0 {
		fs := runtime.CallersFrames([]uintptr{r.PC})
		f, _ := fs.Next()
		fmt.Fprintf(h.w, " %s%s%s=%s%s%s",
			colorYellow, "source", colorReset,
			colorWhite, fmt.Sprintf("%s:%d", f.File, f.Line), colorReset)
	}

	fmt.Fprintf(h.w, "\n")
	return nil
}

// writeAttr writes a single attribute with colors
func (h *ColorHandler) writeAttr(a slog.Attr) {
	key := a.Key
	if len(h.groups) > 0 {
		// Prepend group names
		for _, g := range h.groups {
			key = g + "." + key
		}
	}

	valueStr := a.Value.String()
	// Format: yellow_key=white_value
	fmt.Fprintf(h.w, " %s%s%s=%s%s%s",
		colorYellow, key, colorReset,
		colorWhite, valueStr, colorReset)
}

// WithAttrs returns a new handler with the given attributes
func (h *ColorHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	// Create a new handler with preformatted attributes
	h2 := *h
	// Store attributes to be written with colors
	h2.preformattedAttrs = make([]slog.Attr, len(attrs))
	copy(h2.preformattedAttrs, attrs)
	return &h2
}

// WithGroup returns a new handler with the given group
func (h *ColorHandler) WithGroup(name string) slog.Handler {
	h2 := *h
	h2.groups = append(h2.groups, name)
	return &h2
}

// InitColorLogger initializes the default logger with color support
func InitColorLogger() {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	handler := NewColorHandler(os.Stdout, opts)
	logger := slog.New(handler)
	slog.SetDefault(logger)
}
