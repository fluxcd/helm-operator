package helm

import (
	"fmt"
	"io"

	"github.com/go-kit/kit/log"
)

// logWriter wraps a `log.Logger` so it can be used as an `io.Writer`
type logWriter struct {
	log.Logger
}

func NewLogWriter(logger log.Logger) io.Writer {
	return &logWriter{logger}
}

func (l *logWriter) Write(p []byte) (n int, err error) {
	origLen := len(p)
	if len(p) > 0 && p[len(p)-1] == '\n' {
		p = p[:len(p)-1] // Cut terminating newline
	}
	l.Log("info", fmt.Sprintf("%s", p))
	return origLen, nil
}
