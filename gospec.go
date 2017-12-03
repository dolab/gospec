package gospec

import (
	"bytes"
	"fmt"
	"runtime"
	"strings"
	"testing"
)

// gospec implements TestingT and TestingR.
//
// Its implementation of Errorf writes the output that would be produced by
// testing.T.Errorf to an internal bytes.Buffer.
//
// Its implementation of Run runs specs using testing.T.Run if defined, or using custom func.
type gospec struct {
	t   *testing.T
	buf bytes.Buffer
}

func (spec *gospec) Errorf(format string, args ...interface{}) {
	// implementation of decorate is copied from testing.T
	decorate := func(s string) string {
		_, file, line, ok := runtime.Caller(3) // decorate + log + public function.
		if ok {
			// Truncate file name at last file name separator.
			if index := strings.LastIndex(file, "/"); index >= 0 {
				file = file[index+1:]
			} else if index = strings.LastIndex(file, "\\"); index >= 0 {
				file = file[index+1:]
			}
		} else {
			file = "???"
			line = 1
		}

		buf := new(bytes.Buffer)

		// Every line is indented at least one tab.
		buf.WriteByte('\t')
		fmt.Fprintf(buf, "%s:%d: ", file, line)

		lines := strings.Split(s, "\n")
		if l := len(lines); l > 1 && lines[l-1] == "" {
			lines = lines[:l-1]
		}
		for i, line := range lines {
			if i > 0 {
				// Second and subsequent lines are indented an extra tab.
				buf.WriteString("\n\t\t")
			}
			buf.WriteString(line)
		}
		buf.WriteByte('\n')
		return buf.String()
	}

	spec.buf.WriteString(decorate(fmt.Sprintf(format, args...)))
}

func (spec *gospec) Run(title string, runner func(t *testing.T)) bool {
	(func(iface interface{}) {

		switch iface.(type) {
		case TestingR:
			iface.(TestingR).Run(title, func(t *testing.T) {
				runner(t)
			})

		default:
			println(gray.Paint(title))
			runner(spec.t)

		}

	})(spec.t)

	return true
}

func (t *gospec) String() string {
	return t.buf.String()
}
