package logging

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"
)

// ANSI escape codes.
const (
	ansiReset  = "\033[0m"
	ansiBold   = "\033[1m"
	ansiDim    = "\033[2m"
	ansiRed    = "\033[31m"
	ansiYellow = "\033[33m"
	ansiCyan   = "\033[36m"
	ansiGray   = "\033[90m"
)

// ColorHandler is a slog.Handler that writes compact, colored, human-readable lines:
//
//	15:04:05.000 INFO  message  key=value  key=value
//
// Colors are auto-detected (tty + NO_COLOR). JSON format always omits colors.
type ColorHandler struct {
	mu          sync.Mutex
	out         io.Writer
	opts        slog.HandlerOptions
	color       bool
	preAttrs    []slog.Attr
	groupPrefix string
}

// NewColorHandler creates a ColorHandler. Colors are enabled when out is a tty
// and the NO_COLOR environment variable is not set.
func NewColorHandler(out io.Writer, opts *slog.HandlerOptions) *ColorHandler {
	h := &ColorHandler{out: out, color: isTTY(out)}
	if opts != nil {
		h.opts = *opts
	}
	return h
}

func (h *ColorHandler) clone() *ColorHandler {
	c := *h
	c.preAttrs = append([]slog.Attr(nil), h.preAttrs...)
	return &c
}

func (h *ColorHandler) Enabled(_ context.Context, level slog.Level) bool {
	min := slog.LevelInfo
	if h.opts.Level != nil {
		min = h.opts.Level.Level()
	}
	return level >= min
}

func (h *ColorHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	c := h.clone()
	c.preAttrs = append(c.preAttrs, attrs...)
	return c
}

func (h *ColorHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	c := h.clone()
	if c.groupPrefix != "" {
		c.groupPrefix += "." + name
	} else {
		c.groupPrefix = name
	}
	return c
}

func (h *ColorHandler) Handle(_ context.Context, r slog.Record) error {
	var buf bytes.Buffer

	if h.color {
		// Timestamp — dim gray, short
		buf.WriteString(ansiGray)
		buf.WriteString(r.Time.Format("15:04:05.000"))
		buf.WriteString(ansiReset)
		buf.WriteByte(' ')

		// Level — colored, 5-char fixed width
		clr, lbl := levelStyle(r.Level)
		buf.WriteString(ansiBold + clr)
		buf.WriteString(lbl)
		buf.WriteString(ansiReset)
		buf.WriteByte(' ')

		// Message — bold white
		buf.WriteString(ansiBold)
		buf.WriteString(r.Message)
		buf.WriteString(ansiReset)
	} else {
		buf.WriteString(r.Time.Format(time.RFC3339Nano))
		buf.WriteByte(' ')
		_, lbl := levelStyle(r.Level)
		buf.WriteString(strings.TrimRight(lbl, " "))
		buf.WriteByte(' ')
		buf.WriteString(r.Message)
	}

	for _, a := range h.preAttrs {
		h.appendAttr(&buf, a)
	}
	r.Attrs(func(a slog.Attr) bool {
		h.appendAttr(&buf, a)
		return true
	})

	buf.WriteByte('\n')

	h.mu.Lock()
	defer h.mu.Unlock()
	_, err := h.out.Write(buf.Bytes())
	return err
}

func (h *ColorHandler) appendAttr(buf *bytes.Buffer, a slog.Attr) {
	key := a.Key
	if h.groupPrefix != "" {
		key = h.groupPrefix + "." + key
	}
	val := formatValue(a.Value)

	if h.color {
		buf.WriteString("  ")
		buf.WriteString(ansiCyan)
		buf.WriteString(key)
		buf.WriteString(ansiReset + "=")
		buf.WriteString(val)
	} else {
		buf.WriteByte(' ')
		buf.WriteString(key)
		buf.WriteByte('=')
		buf.WriteString(val)
	}
}

func levelStyle(l slog.Level) (color, label string) {
	switch {
	case l >= slog.LevelError:
		return ansiRed, "ERR  "
	case l >= slog.LevelWarn:
		return ansiYellow, "WARN "
	case l >= slog.LevelInfo:
		return ansiCyan, "INFO "
	default:
		return ansiDim, "DBG  "
	}
}

func formatValue(v slog.Value) string {
	switch v.Kind() {
	case slog.KindString:
		s := v.String()
		if s == "" || strings.ContainsAny(s, " \t\n\"=") {
			return fmt.Sprintf("%q", s)
		}
		return s
	case slog.KindTime:
		return v.Time().Format(time.RFC3339)
	default:
		return fmt.Sprintf("%v", v.Any())
	}
}

// isTTY returns true when w is a real terminal and NO_COLOR is not set.
func isTTY(w io.Writer) bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}
