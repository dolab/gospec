package gospec

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path"
	"runtime"
	"strings"
	"testing"
)

type grammar struct {
	t      *testing.T
	n      int32
	actual interface{}
}

func (g *grammar) Errorf(format string, args ...interface{}) {
	decorate := func() string {
		buf := bytes.NewBuffer(nil)

		var (
			name   = "???"
			callee = "???"
			lino   = 1
		)

		pc, file, lino, ok := runtime.Caller(5) // decorate + log + public function.
		if ok {
			data, err := ioutil.ReadFile(file)
			if err != nil {
				fmt.Fprintf(buf, magenta.Paint("Error: ioutil.ReadFile(%s): %v"), file, err)
			} else {
				for i, line := range strings.Split(string(data), "\n") {
					if i+1 == lino {
						fmt.Fprintf(buf, magenta.Paint("Failure/Error:")+" %s", keywordRe.ReplaceAllStringFunc(strings.TrimSpace(line), func(key string) string {
							if key[len(key)-1] == '(' {
								return blue.Paint(key[:len(key)-1]) + "("
							}

							return blue.Paint(key)
						}))
					}
				}
			}

			// resolve file name with parent dir if exists
			if i := strings.LastIndex(file, "/src/"); i >= 0 {
				name = file[i+1:]
			} else {
				dir, filename := path.Split(file)

				dir = strings.TrimSuffix(dir, "/")
				if i := strings.LastIndex(dir, "/"); i >= 0 {
					name = dir[i+1:] + "/" + filename
				} else if i = strings.LastIndex(dir, "\\"); i >= 0 {
					name = dir[i+1:] + "/" + filename
				} else {
					name = filename
				}
			}

			callee = runtime.FuncForPC(pc).Name()
		}

		buf.WriteString("\n\n")

		content := []string{}

		switch args[0].(type) {
		case *testingOutput:
			for _, label := range args[0].(*testingOutput).labels {
				switch label.label {
				case labelError, labelErrorTrace, labelMessages:
					// ignore

				default:
					if label.label != "Diff" {
						content = append(content, fmt.Sprintf("%s: %s\n", label.label, strings.TrimSpace(label.content)))
					} else {
						content = append(content, fmt.Sprintf("%s:\n%s\n", label.label, strings.TrimSpace(label.content)))
					}

				}
			}
		}

		if len(content) > 0 {
			buf.WriteString(magenta.Paint(strings.Join(content, "\n")))
		} else {
			buf.WriteByte('\r')
		}

		fmt.Fprintf(buf, cyan.Paint("\r\n\n// ./%s:%d:in %s\n"), name, lino, callee)

		fmt.Printf(">>> %#v\n", buf.String())
		return buf.String()
	}

	g.t.Logf("\r\t%s", green.Paint(fmt.Sprintf("%d) ", g.n))+decorate())
	g.t.Fail()
}
