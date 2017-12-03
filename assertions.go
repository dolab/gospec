package gospec

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/buger/jsonparser"
)

// IsType asserts that the specified values are of the same type.
//
// 	assert.IsType(t, int32, int32(123), "int32(123) should to be of type int32")
//
// Returns whether the assertion was successful (true) or not (false).
func IsType(t TestingT, expected, actual interface{}, extras ...interface{}) bool {
	if !DeepEqual(reflect.TypeOf(expected), reflect.TypeOf(actual)) {
		exps, acts := toString(expected, actual)

		return Errorf(t, "Expect to be of the same type", []labeledOutput{
			{
				label:   labelMessages,
				content: formatExtras(extras...),
			},
			{
				label:   "+expected",
				content: exps,
			},
			{
				label:   "-received",
				content: acts,
			},
		})
	}

	return true
}

// Implements asserts that the specified value implements the specified interface.
//
//    assert.Implements(t, (*MyInterface)(nil), new(MyObject), "MyObject should implement MyInterface")
//
// Returns whether the assertion was successful (true) or not (false).
func Implements(t TestingT, expectedIface, actual interface{}, extras ...interface{}) bool {
	if expectedIface == nil || actual == nil {
		iface, value := toString(expectedIface, actual)

		return Errorf(t, "Expect to implement interface", []labeledOutput{
			{
				label:   labelMessages,
				content: formatExtras(extras...),
			},
			{
				label:   "+interface",
				content: strings.TrimPrefix(iface, "*"),
			},
			{
				label:   "+value",
				content: value,
			},
		})
	}

	ifaceType := reflect.TypeOf(expectedIface).Elem()
	if !reflect.TypeOf(actual).Implements(ifaceType) {
		iface, value := toString(expectedIface, actual)

		return Errorf(t, "Expect to implement interface", []labeledOutput{
			{
				label:   labelMessages,
				content: formatExtras(extras...),
			},
			{
				label:   "+interface",
				content: strings.TrimPrefix(iface, "*"),
			},
			{
				label:   "+value",
				content: value,
			},
		})
	}

	return true
}

// Equal asserts that two values are equal.
//
//    assert.Equal(t, 123, 123, "123 and 123 should be equal")
//
// Returns whether the assertion was successful (true) or not (false).
//
// Pointer variable equality is determined based on the equality of the
// referenced values (as opposed to the memory addresses).
func Equal(t TestingT, expected, actual interface{}, extras ...interface{}) bool {
	if !DeepEqual(expected, actual) {
		return Errorf(t, "Expect to be equal", []labeledOutput{
			{
				label:   labelMessages,
				content: formatExtras(extras...),
			},
			{
				label:   "Diff",
				content: diff(expected, actual),
			},
		})
	}

	return true
}

// NotEqual asserts that the specified values are NOT equal.
//
//    assert.NotEqual(t, obj1, obj2, "two objects shouldn't be equal")
//
// Returns whether the assertion was successful (true) or not (false).
//
// Pointer variable equality is determined based on the equality of the
// referenced values (as opposed to the memory addresses).
func NotEqual(t TestingT, expected, actual interface{}, extras ...interface{}) bool {
	if DeepEqual(expected, actual) {
		exps, acts := toString(expected, actual)

		return Errorf(t, "Expect to be NOT equal", []labeledOutput{
			{
				label:   labelMessages,
				content: formatExtras(extras...),
			},
			{
				label:   "-expected",
				content: exps,
			},
			{
				label:   "+received",
				content: acts,
			},
		})
	}

	return true
}

// EqualValues asserts that two objects are equal or convertable to the same types
// and equal.
//
//    assert.EqualValues(t, uint32(123), int32(123), "uint32(123) and int32(123) should be equal values")
//
// Returns whether the assertion was successful (true) or not (false).
func EqualValues(t TestingT, expected, actual interface{}, extras ...interface{}) bool {
	if !DeepEqualValues(expected, actual) {
		return Errorf(t, "Expect to be equal in values", []labeledOutput{
			{
				label:   labelMessages,
				content: formatExtras(extras...),
			},
			{
				label:   "Diff",
				content: diff(expected, actual),
			},
		})
	}

	return true
}

// EqualJSON asserts that two JSON strings are equivalent.
//
//  assert.EqualJSON(t, `{"hello": "world", "foo": "bar"}`, `{"foo": "bar", "hello": "world"}`)
//
// Returns whether the assertion was successful (true) or not (false).
func EqualJSON(t TestingT, expected, actual string, extras ...interface{}) bool {
	var expectedValue, actualValue interface{}

	if err := json.Unmarshal([]byte(expected), &expectedValue); err != nil {
		exps, _ := toString(expected, nil)

		return Errorf(t, "Expect value should be valid json.", []labeledOutput{
			{
				label:   labelMessages,
				content: formatExtras(extras...),
			},
			{
				label:   "+expected:",
				content: exps,
			},
			{
				label:   "+JSON Parse:",
				content: err.Error(),
			},
		})
	}

	if err := json.Unmarshal([]byte(actual), &actualValue); err != nil {
		_, acts := toString(nil, actual)

		return Errorf(t, "Actual value should be valid json.", []labeledOutput{
			{
				label:   labelMessages,
				content: formatExtras(extras...),
			},
			{
				label:   "+actual:",
				content: acts,
			},
			{
				label:   "+JSON Parse:",
				content: err.Error(),
			},
		})
	}

	return Equal(t, expectedValue, actualValue, extras...)
}

// Exactly asserts that two values are equal, both value and type.
//
//    assert.Exactly(t, int32(123), int64(123), "int32(123) and int64(123) should NOT be equal")
//
// Returns whether the assertion was successful (true) or not (false).
func Exactly(t TestingT, expected, actual interface{}, extras ...interface{}) bool {
	if !reflect.DeepEqual(expected, actual) {
		return Errorf(t, "Expect to be equal in deep, both types and values", []labeledOutput{
			{
				label:   labelMessages,
				content: formatExtras(extras...),
			},
			{
				label:   "Diff",
				content: diff(expected, actual),
			},
		})
	}

	return true
}

// Nil asserts that the specified value is nil.
//
//    assert.Nil(t, err, "err should be nothing")
//
// Returns whether the assertion was successful (true) or not (false).
func Nil(t TestingT, v interface{}, extras ...interface{}) bool {
	if !IsNil(v) {
		var nilval interface{}

		rval := reflect.ValueOf(v)
		if rval.IsValid() {
			switch rval.Kind() {
			case reflect.Slice, reflect.Map, reflect.Ptr, reflect.Chan, reflect.Func, reflect.Interface:
				nilval = reflect.Zero(rval.Type()).Interface()
			}
		}

		exps, acts := toString(nilval, v)

		return Errorf(t, "Expect to be nil", []labeledOutput{
			{
				label:   labelMessages,
				content: formatExtras(extras...),
			},
			{
				label:   "-expected",
				content: exps,
			},
			{
				label:   "+received",
				content: acts,
			},
		})
	}

	return true
}

// NotNil asserts that the specified value is not nil.
//
//    assert.NotNil(t, err, "err should be something")
//
// Returns whether the assertion was successful (true) or not (false).
func NotNil(t TestingT, v interface{}, extras ...interface{}) bool {
	if IsNil(v) {
		var nilval interface{}

		rval := reflect.ValueOf(v)
		if rval.IsValid() {
			switch rval.Kind() {
			case reflect.Slice, reflect.Map, reflect.Ptr, reflect.Chan, reflect.Func, reflect.Interface:
				nilval = reflect.Zero(rval.Type()).Interface()
			}
		}

		exps, acts := toString(nilval, v)

		return Errorf(t, "Expect to be NOT nil", []labeledOutput{
			{
				label:   labelMessages,
				content: formatExtras(extras...),
			},
			{
				label:   "-expected",
				content: exps,
			},
			{
				label:   "+received",
				content: acts,
			},
		})
	}

	return true
}

// True asserts that the specified value is true.
//
//    assert.True(t, myBool, "myBool should be true")
//
// Returns whether the assertion was successful (true) or not (false).
func True(t TestingT, v interface{}, extras ...interface{}) bool {
	val, ok := v.(bool)
	if !ok || val != true {
		exps, acts := toString(true, v)

		return Errorf(t, "Expect to be true", []labeledOutput{
			{
				label:   labelMessages,
				content: formatExtras(extras...),
			},
			{
				label:   "-expected",
				content: exps,
			},
			{
				label:   "+received",
				content: acts,
			},
		})
	}

	return true
}

// False asserts that the specified value is false.
//
//    assert.False(t, myBool, "myBool should be false")
//
// Returns whether the assertion was successful (true) or not (false).
func False(t TestingT, v interface{}, extras ...interface{}) bool {
	val, ok := v.(bool)
	if !ok || val != false {
		exps, acts := toString(false, v)

		return Errorf(t, "Expect to be false", []labeledOutput{
			{
				label:   labelMessages,
				content: formatExtras(extras...),
			},
			{
				label:   "-expected",
				content: exps,
			},
			{
				label:   "+received",
				content: acts,
			},
		})
	}

	return true
}

// Zero asserts that v is the zero value for its type and returns the truth.
func Zero(t TestingT, v interface{}, extras ...interface{}) bool {
	if v != nil && !reflect.DeepEqual(v, reflect.Zero(reflect.TypeOf(v)).Interface()) {
		exps, acts := toString(reflect.Zero(reflect.TypeOf(v)).Interface(), v)

		return Errorf(t, "Expect to be zero", []labeledOutput{
			{
				label:   labelMessages,
				content: formatExtras(extras...),
			},
			{
				label:   "-expected",
				content: exps,
			},
			{
				label:   "+received",
				content: acts,
			},
		})
	}

	return true
}

// NotZero asserts that v is not the zero value for its type and returns the truth.
func NotZero(t TestingT, v interface{}, extras ...interface{}) bool {
	if v == nil || reflect.DeepEqual(v, reflect.Zero(reflect.TypeOf(v)).Interface()) {
		var acts = "<nil>"
		if v != nil {
			_, acts = toString(reflect.Zero(reflect.TypeOf(v)).Interface(), v)
		}

		return Errorf(t, "Expect to be NOT zero", []labeledOutput{
			{
				label:   labelMessages,
				content: formatExtras(extras...),
			},
			{
				label:   "-expected",
				content: fmt.Sprintf("(%T)(???)", v),
			},
			{
				label:   "+received",
				content: acts,
			},
		})
	}

	return true
}

// Empty asserts that the specified value is empty.  I.e. nil, "", false, 0 or either
// a slice or a channel with len == 0.
//
//  assert.Empty(t, obj)
//
// Returns whether the assertion was successful (true) or not (false).
func Empty(t TestingT, v interface{}, extras ...interface{}) bool {
	if !IsEmpty(v) {
		var acts = "<nil>"
		if v != nil {
			_, acts = toString(reflect.Zero(reflect.TypeOf(v)).Interface(), v)
		}

		return Errorf(t, "Expect to be empty", []labeledOutput{
			{
				label:   labelMessages,
				content: formatExtras(extras...),
			},
			{
				label:   "-expected",
				content: fmt.Sprintf("(%T)()", v),
			},
			{
				label:   "+received",
				content: acts,
			},
		})
	}

	return true
}

// NotEmpty asserts that the specified value is NOT empty.  I.e. not nil, "", false, 0 or either
// a slice or a channel with len == 0.
//
//  assert.NotEmpty(t, obj)
//
// Returns whether the assertion was successful (true) or not (false).
func NotEmpty(t TestingT, v interface{}, extras ...interface{}) bool {
	if IsEmpty(v) {
		var acts = "<nil>"
		if v != nil {
			_, acts = toString(reflect.Zero(reflect.TypeOf(v)).Interface(), v)
		}

		return Errorf(t, "Expect to be NOT empty", []labeledOutput{
			{
				label:   labelMessages,
				content: formatExtras(extras...),
			},
			{
				label:   "-expected",
				content: fmt.Sprintf("(%T)(???)", v),
			},
			{
				label:   "+received",
				content: acts,
			},
		})
	}

	return true
}

// Contains asserts that the specified string, list(array, slice, channel...) or map contains the
// specified substring or element.
//
//    assert.Contains(t, "Hello World", "World", "But 'Hello World' does contain 'World'")
//    assert.Contains(t, ["Hello", "World"], "World", "But ["Hello", "World"] does contain 'World'")
//    assert.Contains(t, {"Hello": "World"}, "Hello", "But {'Hello': 'World'} does contain 'Hello'")
//
// Returns whether the assertion was successful (true) or not (false).
func Contains(t TestingT, v, element interface{}, extras ...interface{}) bool {
	if !ContainsElement(v, element) {
		return Errorf(t, "Expect to include substring or element", []labeledOutput{
			{
				label:   labelMessages,
				content: formatExtras(extras...),
			},
			{
				label:   "Diff",
				content: diff(v, element),
			},
		})
	}

	return true
}

// NotContains asserts that the specified string, list(array, slice, channel...) or map does NOT contain the
// specified substring or element.
//
//    assert.NotContains(t, "Hello World", "Earth", "But 'Hello World' does NOT contain 'Earth'")
//    assert.NotContains(t, ["Hello", "World"], "Earth", "But ['Hello', 'World'] does NOT contain 'Earth'")
//    assert.NotContains(t, {"Hello": "World"}, "Earth", "But {'Hello': 'World'} does NOT contain 'Earth'")
//
// Returns whether the assertion was successful (true) or not (false).
func NotContains(t TestingT, v, element interface{}, extras ...interface{}) bool {
	if ContainsElement(v, element) {
		exps, acts := toString(v, element)

		return Errorf(t, "Expect to NOT include substring or element", []labeledOutput{
			{
				label:   labelMessages,
				content: formatExtras(extras...),
			},
			{
				label:   "-value",
				content: exps,
			},
			{
				label:   "+element",
				content: acts,
			},
		})
	}

	return true
}

// Match asserts that a specified regexp matches the given value.
//
//  assert.Match(t, regexp.MustCompile("start"), "it's starting")
//  assert.Match(t, "start...$", "it's not starting")
//
// Returns whether the assertion was successful (true) or not (false).
func Match(t TestingT, r, v interface{}, extras ...interface{}) bool {
	reg, ok := tryMatch(r, v)
	if !ok {
		_, acts := toString(nil, v)

		Errorf(t, "Expect to match regexp", []labeledOutput{
			{
				label:   labelMessages,
				content: formatExtras(extras...),
			},
			{
				label:   "-regexp",
				content: fmt.Sprintf("%#v", reg.String()),
			},
			{
				label:   "+value",
				content: fmt.Sprintf("%#v", acts),
			},
		})
	}

	return ok
}

// NotMatch asserts that a specified regexp does not match a stringify of the given value.
//
//  assert.NotMatch(t, regexp.MustCompile("starts"), "it's starting")
//  assert.NotMatch(t, "^start", "it's not starting")
//
// Returns whether the assertion was successful (true) or not (false).
func NotMatch(t TestingT, r, v interface{}, extras ...interface{}) bool {
	reg, ok := tryMatch(r, v)
	if ok {
		_, acts := toString(nil, v)

		Errorf(t, "Expect to NOT match regexp", []labeledOutput{
			{
				label:   labelMessages,
				content: formatExtras(extras...),
			},
			{
				label:   "-regexp",
				content: fmt.Sprintf("%#v", reg.String()),
			},
			{
				label:   "+value",
				content: fmt.Sprintf("%#v", acts),
			},
		})
	}

	return !ok
}

// Condition uses custom Comparison to assert a complex condition.
func Condition(t TestingT, comp Comparison, extras ...interface{}) bool {
	ok := comp()
	if !ok {
		exps, acts := toString(true, ok)

		return Errorf(t, "Expect to return true", []labeledOutput{
			{
				label:   labelMessages,
				content: formatExtras(extras...),
			},
			{
				label:   "-expected",
				content: exps,
			},
			{
				label:   "+received",
				content: acts,
			},
		})
	}

	return ok
}

// Len asserts that the specified value has specific length.
// It will fail if the value has a type that len() not accept.
//
//    assert.Len(t, mySlice, 3, "The size of slice is not 3")
//
// Returns whether the assertion was successful (true) or not (false).
func Len(t TestingT, v interface{}, length int, extras ...interface{}) bool {
	n, ok := tryLen(v)
	if !ok {
		_, acts := toString(nil, v)

		return Errorf(t, fmt.Sprintf("Expect to apply buildin len() on %s", acts), []labeledOutput{
			{
				label:   labelMessages,
				content: formatExtras(extras...),
			},
		})
	}

	if n != length {
		_, acts := toString(nil, v)

		return Errorf(t, fmt.Sprintf("Expect %s to have %d item(s)", acts, length), []labeledOutput{
			{
				label:   labelMessages,
				content: formatExtras(extras...),
			},
			{
				label:   "-expected",
				content: strconv.Itoa(length),
			},
			{
				label:   "+received",
				content: strconv.Itoa(n),
			},
		})
	}

	return true
}

// InDelta asserts that the two numerals are within delta of each other.
//
// 	 assert.InDelta(t, math.Pi, (22 / 7.0), 0.01)
//
// Returns whether the assertion was successful (true) or not (false).
func InDelta(t TestingT, expected, actual interface{}, delta float64, extras ...interface{}) bool {
	expf, expok := toFloat(expected)
	actf, actok := toFloat(actual)

	if !expok || !actok {
		exps, acts := toString(expected, actual)

		return Errorf(t, "Parameters must be numerical", []labeledOutput{
			{
				label:   labelMessages,
				content: formatExtras(extras...),
			},
			{
				label:   "-expected",
				content: exps,
			},
			{
				label:   "+received",
				content: acts,
			},
		})
	}

	if math.IsNaN(expf) || math.IsNaN(actf) {
		exps, acts := toString(expf, actf)

		return Errorf(t, "Both expected and actual values must NOT be NaN", []labeledOutput{
			{
				label:   labelMessages,
				content: formatExtras(extras...),
			},
			{
				label:   "-expected",
				content: exps,
			},
			{
				label:   "+received",
				content: acts,
			},
		})
	}

	value := expf - actf
	if value < -delta || value > delta {
		exps, acts := toString(expected, actual)

		return Errorf(t, fmt.Sprintf("Expect the delta between two numbers within %v", delta), []labeledOutput{
			{
				label:   labelMessages,
				content: formatExtras(extras...),
			},
			{
				label:   "+calculated:",
				content: fmt.Sprintf("%s - %s = %v", exps, acts, value),
			},
		})
	}

	return true
}

// WithinDuration asserts that the two times are within duration delta of each other.
//
//   assert.WithinDuration(t, time.Now(), time.Now(), 10*time.Second, "The difference should not be more than 10s")
//
// Returns whether the assertion was successful (true) or not (false).
func WithinDuration(t TestingT, expected, actual time.Time, delta time.Duration, extras ...interface{}) bool {
	value := expected.Sub(actual)
	if value < -delta || value > delta {
		exps, acts := toString(expected, actual)

		return Errorf(t, fmt.Sprintf("Expect the deviation between two times within %v", delta), []labeledOutput{
			{
				label:   labelMessages,
				content: formatExtras(extras...),
			},
			{
				label:   "+calculated:",
				content: fmt.Sprintf("%s.Sub(%s) = %v", exps, acts, value),
			},
		})
	}

	return true
}

// Error asserts that a value is an error (i.e. `errors.New("some message")`).
//
//   _, err := Func()
//   assert.Error(t, err, "An error was expected")
//
// Returns whether the assertion was successful (true) or not (false).
func Error(t TestingT, v interface{}, extras ...interface{}) bool {
	_, ok := v.(error)
	if !ok {
		_, acts := toString(nil, v)

		return Errorf(t, "Expect to be an error", []labeledOutput{
			{
				label:   labelMessages,
				content: formatExtras(extras...),
			},
			{
				label:   "-expected",
				content: fmt.Sprintf("_, ok := %s.(error); ok == true", acts),
			},
			{
				label:   "+received",
				content: fmt.Sprintf("_, ok := %s.(error); ok == false", acts),
			},
		})
	}

	return true
}

// NotError asserts that a value is not an error (i.e. `nil`).
//
//   _, err := Func()
//   assert.NotError(t, err)
//
// Returns whether the assertion was successful (true) or not (false).
func NotError(t TestingT, v interface{}, extras ...interface{}) bool {
	_, ok := v.(error)
	if ok {
		_, acts := toString(nil, v)

		return Errorf(t, "Expect to be NOT an error", []labeledOutput{
			{
				label:   labelMessages,
				content: formatExtras(extras...),
			},
			{
				label:   "-expected",
				content: fmt.Sprintf("_, ok := %s.(error); ok == false", acts),
			},
			{
				label:   "+received",
				content: fmt.Sprintf("_, ok := %s.(error); ok == true", acts),
			},
		})
	}

	return true
}

// EqualErrors asserts that a value is an error (i.e. not `nil`)
// and it is equal to the provided error string.
//
//   _, err := Func()
//   assert.EqualErrors(t, err,  expectedErrorString, "An error was expected")
//
// Returns whether the assertion was successful (true) or not (false).
func EqualErrors(t TestingT, actualErr, expectedErr interface{}, extras ...interface{}) bool {
	if !Error(t, actualErr, extras...) {
		return false
	}

	expected := ""
	switch expectedErr.(type) {
	case error:
		expected = expectedErr.(error).Error()

	case string:
		expected = expectedErr.(string)

	case []byte:
		expected = string(expectedErr.([]byte))

	default:
		expected = fmt.Sprintf("%v", expectedErr)
	}

	// don't need to use deep equals here, we know they are both strings
	actual := actualErr.(error).Error()
	if expected != actual {
		return Errorf(t, "Expect to be error with the same message", []labeledOutput{
			{
				label:   labelMessages,
				content: formatExtras(extras...),
			},
			{
				label:   "Diff",
				content: diff(expected, actual),
			},
		})
	}

	return true
}

// Panics asserts that the code inside the specified PanicRecover panics.
//
//   assert.Panics(t, func(){
//     GoCrazy()
//   }, "Calling GoCrazy() should panic")
//
// Returns whether the assertion was successful (true) or not (false).
func Panics(t TestingT, f PanicRecover, extras ...interface{}) bool {
	paniced, _ := recovery(f)
	if !paniced {
		return Errorf(t, "Expect to panic with invocation", []labeledOutput{
			{
				label:   labelMessages,
				content: formatExtras(extras...),
			},
			// {
			// 	label:   "Panic Value",
			// 	content: fmt.Sprintf("%v", err),
			// },
		})
	}

	return true
}

// NotPanics asserts that the code inside the specified PanicRecover does NOT panic.
//
//   assert.NotPanics(t, func(){
//     RemainCalm()
//   }, "Calling RemainCalm() should NOT panic")
//
// Returns whether the assertion was successful (true) or not (false).
func NotPanics(t TestingT, f PanicRecover, extras ...interface{}) bool {
	paniced, err := recovery(f)
	if paniced {
		return Errorf(t, "Expect to NOT panic with invocation", []labeledOutput{
			{
				label:   labelMessages,
				content: formatExtras(extras...),
			},
			{
				label:   "Panic Value",
				content: fmt.Sprintf("%v", err),
			},
		})
	}

	return true
}

// JSONContains asserts that JSON strings contains specified key.
//
//  assert.JSONContains(t, `{"hello": "world", "foo": "bar"}`, "hello")
//
// Returns whether the assertion was successful (true) or not (false).
func JSONContains(t TestingT, jsonData, searchKeyPath string, extras ...interface{}) bool {
	var jsonValue interface{}

	if err := json.Unmarshal([]byte(jsonData), &jsonValue); err != nil {
		exps, _ := toString(jsonData, nil)

		return Errorf(t, "Expect data should be valid json", []labeledOutput{
			{
				label:   labelMessages,
				content: formatExtras(extras...),
			},
			{
				label:   "+JSON",
				content: exps,
			},
			{
				label:   "+JSON Parse",
				content: err.Error(),
			},
		})
	}

	var (
		buf  = []byte(jsonData)
		data []byte
		err  error
	)

	for _, yek := range strings.Split(searchKeyPath, ".") {
		data, _, _, err = jsonparser.Get(buf, yek)
		if err == nil {
			buf = data

			continue
		}

		// is the yek an array subscript?
		n, e := strconv.ParseInt(yek, 10, 32)
		if e != nil {
			break
		}

		var i int64
		jsonparser.ArrayEach(buf, func(arrBuf []byte, arrType jsonparser.ValueType, arrOffset int, arrErr error) {
			if i == n {
				buf = arrBuf
				err = arrErr
			}

			i++
		})
		if err != nil {
			break
		}
	}
	if err != nil {
		exps, _ := toString(jsonData, nil)

		return Errorf(t, fmt.Sprintf("Expect data should contain json key %s", searchKeyPath), []labeledOutput{
			{
				label:   labelMessages,
				content: formatExtras(extras...),
			},
			{
				label:   "+JSON",
				content: exps,
			},
		})
	}

	return true
}

// JSONEqualValues asserts that JSON strings contains value with specified key.
//
//  assert.JSONContains(t, `{"hello": "world", "foo": "bar"}`, "hello", "world")
//
// Returns whether the assertion was successful (true) or not (false).
func JSONEqualValues(t TestingT, jsonData, searchKeyPath string, expected interface{}, extras ...interface{}) bool {
	var jsonValue interface{}

	if err := json.Unmarshal([]byte(jsonData), &jsonValue); err != nil {
		exps, _ := toString(jsonData, nil)

		return Errorf(t, "Expect data should be valid json", []labeledOutput{
			{
				label:   labelMessages,
				content: formatExtras(extras...),
			},
			{
				label:   "+JSON",
				content: exps,
			},
			{
				label:   "+JSON Parse",
				content: err.Error(),
			},
		})
	}

	var (
		actual = []byte(jsonData)
		data   []byte
		err    error
	)

	for _, yek := range strings.Split(searchKeyPath, ".") {
		data, _, _, err = jsonparser.Get(actual, yek)
		if err == nil {
			actual = data

			continue
		}

		// is the yek an array subscript?
		n, e := strconv.ParseInt(yek, 10, 32)
		if e != nil {
			break
		}

		var i int64
		jsonparser.ArrayEach(actual, func(arrBuf []byte, arrType jsonparser.ValueType, arrOffset int, arrErr error) {
			if i == n {
				actual = arrBuf
				err = arrErr
			}

			i++
		})
		if err != nil {
			break
		}
	}
	if err != nil {
		exps, _ := toString(jsonData, nil)

		return Errorf(t, fmt.Sprintf("Expect data should contain json key %s", searchKeyPath), []labeledOutput{
			{
				label:   labelMessages,
				content: formatExtras(extras...),
			},
			{
				label:   "+JSON",
				content: exps,
			},
		})
	}

	var (
		actualValue interface{}
		actualErr   error
	)

	tmpValue := string(actual)
	switch expected.(type) {
	case int, int8, int16, int32, int64:
		actualValue, actualErr = strconv.ParseInt(tmpValue, 10, 64)

	case uint, uint8, uint16, uint32, uint64:
		actualValue, actualErr = strconv.ParseUint(tmpValue, 10, 64)

	case float32, float64:
		actualValue, actualErr = strconv.ParseFloat(tmpValue, 64)

	case bool:
		actualValue = false
		switch tmpValue {
		case "true":
			actualValue = true
		}

	default:
		actualValue = tmpValue

	}

	if actualErr != nil || !DeepEqualValues(expected, actualValue) {
		return Errorf(t, fmt.Sprintf("Expect data should contain json key %s", searchKeyPath), []labeledOutput{
			{
				label:   labelMessages,
				content: formatExtras(extras...),
			},
			{
				label:   "Diff",
				content: diff(expected, string(actual)),
			},
		})
	}

	return true
}
