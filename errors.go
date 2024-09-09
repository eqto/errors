package errors

import (
	"fmt"
	"io"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type stack []uintptr
type data interface{}

type Error struct {
	error
	stack
	data
}

func (e *Error) Cause() error {
	return e.error
}

func (e *Error) Error() string {
	return e.error.Error()
}

func (e *Error) Unwrap() error {
	return e.error
}

func formatFilename(file string, line int) string {
	dir, file := filepath.Split(file)
	split := strings.Split(strings.Trim(dir, `/`), `/`)
	dir = split[len(split)-1]
	return fmt.Sprintf(`%s/%s:%d`, dir, file, line)
}

func (e *Error) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		fmt.Fprintf(s, "%s %s\n", time.Now().Format(`2006-01-02 15:04:05`), e.Error())
		if s.Flag('+') {
			frames := runtime.CallersFrames(e.stack)
			var frame runtime.Frame
			next := true
			for next {
				frame, next = frames.Next()
				if !strings.HasPrefix(frame.Function, `runtime.`) &&
					!strings.HasPrefix(frame.Function, `reflect.Value.`) {
					fmt.Fprintf(s, "%19s %s (%s)\n", ``, frame.Function, formatFilename(frame.File, frame.Line))
				}
			}
		}
	case 's':
		io.WriteString(s, e.Error())
	case 'q':
		fmt.Fprintf(s, "%q", e.Error())
	}
}

func WithStack(err error) error {
	e := wrap(err)
	pc := make([]uintptr, 20)
	n := runtime.Callers(2, pc)
	if n == 0 {
		return e
	}
	e.stack = pc[:n]
	return e
}

func wrap(err error) *Error {
	return &Error{err, nil}
}

func WrapData(err error, data any) error {
	return &Error{err, nil, data}
}

func UnwrapData(err error) any {
	if e, ok := err.(*Error); ok {
		return e.data
	}
	return nil
}