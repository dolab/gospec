package gospec

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"path"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/davecgh/go-spew/spew"
	"github.com/pmezard/go-difflib/difflib"
)

var (
	spewConfig = spew.ConfigState{
		Indent:                  " ",
		DisablePointerAddresses: true,
		DisableCapacities:       true,
		SortKeys:                true,
	}
)

// Errorf reports a failure through and return false
func Errorf(t TestingT, err string, extras ...interface{}) bool {
	traces, padding := getBacktrace()

	output := &testingOutput{}
	output.Add(labeledOutput{
		label:   labelErrorTrace,
		content: strings.Join(traces, labelNewLine+strings.Repeat(" ", padding+1)),
	}).Add(labeledOutput{
		label:   labelError,
		content: err,
	})

	message := ""
	for _, extra := range extras {
		switch extra.(type) {
		case labeledOutput:
			output.Add(extra.(labeledOutput))

		case *labeledOutput:
			output.Add(*(extra.(*labeledOutput)))

		case []labeledOutput:
			for _, label := range extra.([]labeledOutput) {
				output.Add(label)
			}

		case []*labeledOutput:
			for _, label := range extra.([]*labeledOutput) {
				output.Add(*label)
			}

		default:
			message = formatExtras(extras...)

		}

		// break on formatted message
		if message != "" {
			break
		}
	}

	if message != "" {
		output.Add(labeledOutput{
			label:   labelMessages,
			content: message,
		})
	}

	t.Errorf("%s", output)

	return false
}

// DeepEqual determines if two objects are considered equal.
//
// This function does no assertion of any kind.
func DeepEqual(expected, actual interface{}) bool {
	if expected == nil || actual == nil {
		return expected == actual
	}

	expectedValue := reflect.ValueOf(expected)
	actualValue := reflect.ValueOf(actual)
	if !expectedValue.IsValid() || !actualValue.IsValid() {
		return false
	}

	if expectedValue.Kind() == reflect.Func &&
		actualValue.Kind() == reflect.Func &&
		expectedValue.Pointer() == actualValue.Pointer() {
		return true
	}

	return reflect.DeepEqual(expected, actual)
}

// DeepEqualValues gets whether two objects are equal, or if their
// values are equal.
//
// NOTE: it returns true if two values are func with the same addr.
func DeepEqualValues(expected, actual interface{}) bool {
	if DeepEqual(expected, actual) {
		return true
	}

	// try comparison after type conversion
	expectedType := reflect.TypeOf(expected)
	actualType := reflect.TypeOf(actual)
	if expectedType == nil || actualType == nil {
		return false
	}

	expectedValue := reflect.ValueOf(expected)
	actualValue := reflect.ValueOf(actual)
	if !expectedValue.IsValid() || !actualValue.IsValid() {
		return false
	}

	if actualValue.Type().ConvertibleTo(expectedType) {
		return reflect.DeepEqual(expected, actualValue.Convert(expectedType).Interface())
	}

	if expectedValue.Type().ConvertibleTo(actualType) {
		return reflect.DeepEqual(expectedValue.Convert(actualType).Interface(), actual)
	}

	return false
}

// IsNil checks if a specified value is nil or not, without Failing.
func IsNil(v interface{}) bool {
	if v == nil {
		return true
	}

	rval := reflect.ValueOf(v)
	switch rval.Kind() {
	case reflect.Slice, reflect.Map, reflect.Ptr, reflect.Chan, reflect.Func, reflect.Interface:
		return rval.IsNil()
	}

	return false
}

// IsEmpty gets whether the specified value is considered empty or not.
// 	returns true if value is nil or nil pointer
// 	returns true if value is empty string
// 	returns true if value is false
// 	returns true if member of slice/map/channel is empty
// 	returns true if value is equal to zero value of it's type
func IsEmpty(v interface{}) bool {
	switch v {
	case nil:
		return true
	case "":
		return true
	case false:
		return true
	}

	rval := reflect.ValueOf(v)
	switch rval.Kind() {
	case reflect.Map, reflect.Slice, reflect.Chan:
		return rval.Len() == 0

	case reflect.Ptr:
		if rval.IsNil() {
			return true
		}

		rval = rval.Elem()
	}

	if rval.CanInterface() {
		zero := reflect.New(rval.Type()).Elem()

		return reflect.DeepEqual(rval.Interface(), zero.Interface())
	}

	return false
}

// ContainsElement try loop over the list checking if the list includes the element.
// 	return false if impossible.
// 	return true if element was found, false otherwise.
func ContainsElement(v, element interface{}) (ok bool) {
	defer func() {
		if e := recover(); e != nil {
			log.Printf("[PANIC] ContainsElement(%v, %v): %v", v, element, e)
			ok = false
		}
	}()

	listValue := reflect.ValueOf(v)
	elementValue := reflect.ValueOf(element)

	switch listValue.Type().Kind() {
	case reflect.String:
		return strings.Contains(listValue.String(), elementValue.String())

	case reflect.Map: // compare by key
		keys := listValue.MapKeys()
		for i := 0; i < len(keys); i++ {
			if DeepEqual(element, keys[i].Interface()) {
				return true
			}
		}

	case reflect.Slice, reflect.Array: // compare by element
		for i := 0; i < listValue.Len(); i++ {
			if DeepEqual(element, listValue.Index(i).Interface()) {
				return true
			}
		}

	case reflect.Chan: // compare by element
		for i := 0; i < listValue.Len(); i++ {
			value, ok := listValue.Recv()
			if !ok {
				continue
			}

			listValue.Send(value)

			if DeepEqual(element, value.Interface()) {
				return true
			}
		}

	}

	var found bool

	switch v.(type) {
	case io.Reader:
		r := v.(io.Reader)

		b, err := ioutil.ReadAll(r)
		if err != nil {
			return false
		}

		found = strings.Contains(string(b), elementValue.String())

		switch v.(type) {
		case io.Seeker:
			seeker := v.(io.Seeker)
			seeker.Seek(0, 0)

		case io.Writer:
			writer := v.(io.Writer)
			writer.Write(b)

		default:
			log.Printf("[WARN] ContainsElement(%T, %v): CANNOT reset io.Reader for next read\n", v, element)

		}
	}

	return found
}

// Stolen from the `go test` tool.
// isTest tells whether name looks like a test (or benchmark, according to prefix).
// It is a Test (say) if there is a character after Test that is not a lower-case letter.
// We don't want TesticularCancer.
func isTest(name, prefix string) bool {
	if !strings.HasPrefix(name, prefix) {
		return false
	}

	if len(name) == len(prefix) { // "Test" is ok
		return true
	}

	rune, _ := utf8.DecodeRuneInString(name[len(prefix):])
	return !unicode.IsLower(rune)
}

// recovery returns true if the func passed to it panics. Otherwise, it returns false.
func recovery(f PanicRecover) (isPanic bool, err interface{}) {
	func() {
		defer func() {
			err = recover()
			if err != nil {
				isPanic = true
			}
		}()

		// call the target func
		f()
	}()

	return
}

// diff returns a diff of values of the same type and it's type MUST be a struct, map, slice or array.
// Otherwise it returns an empty string.
func diff(expected, actual interface{}) string {
	if expected == nil || actual == nil {
		if expected == actual {
			return ""
		}

		return fmt.Sprintf("--- %T(%v)\n+++ %T(%v)\n\n", expected, expected, actual, actual)
	}

	et, ek := getTypeAndKind(expected)
	at, _ := getTypeAndKind(actual)
	if et != at {
		return fmt.Sprintf("--- %T(%v)\n+++ %T(%v)\n\n", expected, expected, actual, actual)
	}

	switch ek {
	case reflect.String, reflect.Struct, reflect.Slice, reflect.Array, reflect.Map:
		// ignore

	default:
		if reflect.DeepEqual(expected, actual) {
			return ""
		}

		return fmt.Sprintf("--- %T(%v)\n+++ %T(%v)\n\n", expected, expected, actual, actual)
	}

	exps := spewConfig.Sdump(expected)
	acts := spewConfig.Sdump(actual)

	diff, _ := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
		A:        difflib.SplitLines(exps),
		B:        difflib.SplitLines(acts),
		FromFile: "Expected",
		FromDate: "",
		ToFile:   "Actual",
		ToDate:   "",
		Context:  1,
	})

	return diff
}

// tryMatch return *regexp.Regexp instance of r and true if a specified regexp matches a stringify of the given value.
func tryMatch(r, v interface{}) (reg *regexp.Regexp, ok bool) {
	defer func() {
		if e := recover(); e != nil {
			log.Printf("[PANIC] tryMatch(%v, %v): %v\n", r, v, e)
			ok = false
		}
	}()

	reg, ok = r.(*regexp.Regexp)
	if !ok {
		reg = regexp.MustCompile(fmt.Sprint(r))
	}

	ok = reg.FindStringIndex(fmt.Sprint(v)) != nil
	return
}

// tryLen try to get length of value.
// return (false, 0) if impossible.
func tryLen(v interface{}) (length int, ok bool) {
	defer func() {
		if e := recover(); e != nil {
			log.Printf("[PANIC] tryLen(%v): %v\n", v, e)
			ok = false
		}
	}()

	rval := reflect.ValueOf(v)
	length = rval.Len()
	ok = true
	return
}

func toFloat(v interface{}) (float64, bool) {
	var f64 float64
	ok := true

	switch t := v.(type) {
	case uint8:
		f64 = float64(t)
	case uint16:
		f64 = float64(t)
	case uint32:
		f64 = float64(t)
	case uint64:
		f64 = float64(t)
	case int:
		f64 = float64(t)
	case int8:
		f64 = float64(t)
	case int16:
		f64 = float64(t)
	case int32:
		f64 = float64(t)
	case int64:
		f64 = float64(t)
	case float32:
		f64 = float64(t)
	case float64:
		f64 = float64(t)
	default:
		ok = false
	}

	return f64, ok
}

// toString takes two values of arbitrary types and returns string
// representations appropriate to be presented to the user.
//
// If the values are not of the same type, the returned strings will be prefixed
// with the type name, and the value will be enclosed in parenthesis similar
// to a type conversion in the Go grammar.
func toString(expected, actual interface{}) (exps, acts string) {
	if reflect.TypeOf(expected) == reflect.TypeOf(actual) {
		exps = fmt.Sprintf("%#v", expected)
		acts = fmt.Sprintf("%#v", actual)

		return
	}

	expval := reflect.ValueOf(expected)
	actval := reflect.ValueOf(actual)

	switch expected {
	case nil:
		exps = fmt.Sprintf("%T", expected)

	default:
		if expval.Kind() == reflect.Ptr {
			expval = expval.Elem()
		}

		switch expval.Kind() {
		case reflect.Chan, reflect.Func, reflect.Map, reflect.Slice, reflect.Interface:
			if expval.IsNil() {
				exps = fmt.Sprintf("%T", expected)
			} else {
				exps = expval.String()
				if exps == "" {
					exps = "<string Value>"
				}
			}

		default:
			if expval.IsValid() {
				if actval.IsValid() {
					exps = fmt.Sprintf("%T(%#v)", expected, expected)
				} else {
					exps = expval.String()
					if exps == "" {
						exps = "<string Value>"
					}
				}
			} else {
				exps = fmt.Sprintf("%T", expected)
			}
		}
	}

	switch actual {
	case nil:
		acts = fmt.Sprintf("%T", actual)

	default:
		switch actval.Kind() {
		case reflect.Chan, reflect.Func, reflect.Map, reflect.Slice, reflect.Interface:
			if actval.IsNil() {
				acts = fmt.Sprintf("%T", expected)
			} else {
				if actval.Kind() == reflect.Ptr {
					actval = actval.Elem()
				}

				acts = actval.String()
				if acts == "" {
					acts = "<string Value>"
				}
			}

		default:
			if actval.IsValid() {
				if expval.IsValid() {
					acts = fmt.Sprintf("%T(%#v)", actual, actual)
				} else {
					acts = actval.String()
					if acts == "" {
						acts = "<string Value>"
					}
				}
			} else {
				acts = fmt.Sprintf("%T", actual)
			}
		}
	}

	return
}

func getTypeAndKind(v interface{}) (t reflect.Type, k reflect.Kind) {
	t = reflect.TypeOf(v)
	k = t.Kind()

	if k == reflect.Ptr {
		t = t.Elem()
		k = t.Kind()
	}

	return
}

// getWhitespace returns a string that is long enough to overwrite the default
// output from the go testing framework.
func getWhitespace() string {
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		return ""
	}

	_, filename := path.Split(file)

	return strings.Repeat(" ", len(fmt.Sprintf("%s:%d:        ", filename, line)))
}

// getBacktrace is necessary because the assert functions use the testing object
// internally, causing it to print the file:line of the assert method, rather than where
// the problem actually occurred in calling code.
//
// getBacktrace returns an array of strings containing the file and line number
// of each stack frame leading from the current test to the assert call that failed.
func getBacktrace() (callers []string, longestFile int) {
	for i := 1; ; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			// The breaks below failed to terminate the loop, and we ran off the
			// end of the call stack.
			break
		}

		// This is a huge edge case, but it will panic if this is the case, see #180
		if file == "<autogenerated>" {
			break
		}

		f := runtime.FuncForPC(pc)
		if f == nil {
			break
		}
		name := f.Name()

		// testing.tRunner is the standard library function that calls
		// tests. Subtests are called directly by tRunner, without going through
		// the Test/Benchmark/Example function that contains the t.Run calls, so
		// with subtests we should break when we hit tRunner, without adding it
		// to the list of callers.
		if name == "testing.tRunner" {
			break
		}

		parts := strings.Split(file, "/")
		dir := parts[len(parts)-2]
		file = parts[len(parts)-1]

		if !strings.HasSuffix(file, "_test.go") && len(file) > longestFile {
			longestFile = len(file)
		}

		if (dir != "assert" && dir != "mock" && dir != "require") ||
			file == "assertions_test.go" ||
			file == "expectations_test.go" {
			callers = append(callers, fmt.Sprintf("%s:%d", file, line))
		}

		// Drop the package
		segments := strings.Split(name, ".")
		name = segments[len(segments)-1]
		if isTest(name, "Test") ||
			isTest(name, "Benchmark") ||
			isTest(name, "Example") {
			break
		}
	}

	return
}

func formatExtras(extras ...interface{}) string {
	message := ""
	if len(extras) > 0 {
		switch extras[0].(type) {
		case string:
			message = extras[0].(string)

			if len(extras) > 1 {
				message = fmt.Sprintf(message, extras[1:]...)
			}

		case []byte:
			message = string(extras[0].([]byte))

			if len(extras) > 1 {
				message = fmt.Sprintf(message, extras[1:]...)
			}

		default:
			message = fmt.Sprintf("%v", extras)

		}
	}

	return message
}
