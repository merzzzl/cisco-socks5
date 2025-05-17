package log

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
)

var (
	l     = setLoggerOutput(os.Stdout)
	level = zerolog.InfoLevel
)

type Event struct {
	e *zerolog.Event
}

// Debug logs a message at info level.
func Debug() *Event {
	return &Event{l.Debug()}
}

// Info logs a message at info level.
func Info() *Event {
	return &Event{l.Info()}
}

// Info logs a message at info level.
func Warn() *Event {
	return &Event{l.Warn()}
}

// Error logs a message at error level.
func Error() *Event {
	return &Event{l.Error()}
}

// Fatal logs a message at fatal level.
func Fatal() *Event {
	return &Event{l.Fatal()}
}

// Str logs a string with the given key and value.
func (e *Event) Str(key, val string) *Event {
	return &Event{e.e.Str(key, val)}
}

// Err logs an error.
func (e *Event) Err(err error) *Event {
	return &Event{e.e.Err(err)}
}

// Msg logs a message with the given tag and message.
func (e *Event) Msg(tag, msg string) {
	if e == nil {
		return
	}

	e.e.Msg(Colorize(tag, 14) + " " + msg)
}

// Msgf logs a message with the given tag and format string.
func (e *Event) Msgf(tag, format string, v ...any) {
	if e == nil {
		return
	}

	e.Msg(tag, fmt.Sprintf(format, v...))
}

// SetOutput sets the output for logging messages.
func SetOutput(out io.Writer) {
	l = setLoggerOutput(out)
}

func EnableDebug() {
	level = zerolog.DebugLevel

	SetOutput(os.Stdout)
}

func setLoggerOutput(out io.Writer) zerolog.Logger {
	return zlog.Output(zerolog.ConsoleWriter{
		Out: out,
		FormatFieldName: func(i any) string {
			str, ok := i.(string)
			if !ok {
				return ""
			}

			return Colorize(str+"=", 6)
		},
		FormatFieldValue: func(i any) string {
			str, ok := i.(string)
			if !ok {
				return ""
			}

			return Colorize(str, 6)
		},
		FormatErrFieldName: func(any) string {
			return ""
		},
		FormatErrFieldValue: func(i any) string {
			str, ok := i.(string)
			if !ok {
				return ""
			}

			return Colorize(str, 1)
		},
		FormatLevel: consoleDefaultFormatLevel(),
		FormatTimestamp: func(i any) string {
			str, ok := i.(string)
			if !ok {
				return ""
			}

			parse, err := time.Parse(time.RFC3339, str)
			if err != nil {
				return ""
			}

			return Colorize(parse.Format("15:04:05"), 7)
		},
	}).Level(level)
}

func consoleDefaultFormatLevel() zerolog.Formatter {
	return func(i any) string {
		var l string

		if ll, ok := i.(string); ok {
			switch ll {
			case zerolog.LevelTraceValue:
				l = Colorize("TRC", 10)
			case zerolog.LevelDebugValue:
				l = Colorize("DBG", 10)
			case zerolog.LevelInfoValue:
				l = Colorize("INF", 10)
			case zerolog.LevelWarnValue:
				l = Colorize("WRN", 11)
			case zerolog.LevelErrorValue:
				l = Colorize("ERR", 9)
			case zerolog.LevelFatalValue:
				l = Colorize("FTL", 9)
			case zerolog.LevelPanicValue:
				l = Colorize("PNC", 9)
			default:
				l = Colorize(ll, 11)
			}
		} else {
			if i == nil {
				l = Colorize("???", 11)
			} else {
				l = Colorize(strings.ToUpper(fmt.Sprintf("%s", i))[:3], 11)
			}
		}

		return l
	}
}

func Colorize(s string, c int) string {
	return fmt.Sprintf("\033[38;5;%dm%s\033[0m", c, s)
}
