package gospec

import (
	"bytes"
	"errors"
	"io"
	"math"
	"regexp"
	"strings"
	"testing"
	"time"
)

var (
	i     interface{}
	zeros = []interface{}{
		false,
		byte(0),
		complex64(0),
		complex128(0),
		float32(0),
		float64(0),
		int(0),
		int8(0),
		int16(0),
		int32(0),
		int64(0),
		rune(0),
		uint(0),
		uint8(0),
		uint16(0),
		uint32(0),
		uint64(0),
		uintptr(0),
		"",
		[0]interface{}{},
		[]interface{}(nil),
		struct{ x int }{},
		(*interface{})(nil),
		(func())(nil),
		nil,
		interface{}(nil),
		map[interface{}]interface{}(nil),
		(chan interface{})(nil),
		(<-chan interface{})(nil),
		(chan<- interface{})(nil),
	}
	nonZeros = []interface{}{
		true,
		byte(1),
		complex64(1),
		complex128(1),
		float32(1),
		float64(1),
		int(1),
		int8(1),
		int16(1),
		int32(1),
		int64(1),
		rune(1),
		uint(1),
		uint8(1),
		uint16(1),
		uint32(1),
		uint64(1),
		uintptr(1),
		"s",
		[1]interface{}{1},
		[]interface{}{},
		struct{ x int }{1},
		(*interface{})(&i),
		(func())(func() {}),
		interface{}(1),
		map[interface{}]interface{}{},
		(chan interface{})(make(chan interface{})),
		(<-chan interface{})(make(chan interface{})),
		(chan<- interface{})(make(chan interface{})),
	}
)

func TestIsType(t *testing.T) {
	mockT := new(testing.T)

	trueCases := []struct {
		expected, actual interface{}
	}{
		{"Hello", "world"},
		{0, 0},
		{0.0, 0.0},
		{nil, nil},
		{new(bytes.Buffer), bytes.NewBuffer(nil)},
	}

	for i, tc := range trueCases {
		True(t, IsType(mockT, tc.expected, tc.actual), "IsType should return true for trueCases(%d)", i)
	}

	falseCases := []struct {
		expected, actual interface{}
	}{
		{new(bytes.Reader), new(bytes.Buffer)},
		{nil, 0},
		{0, nil},
		{0, 0.0},
		{0.0, 0},
	}

	for i, fc := range falseCases {
		False(t, IsType(mockT, fc.expected, fc.actual), "IsType should return false for false cashe(%d)", i)
	}
}

func TestImplements(t *testing.T) {
	mockT := new(testing.T)

	trueCases := []struct {
		iface, value interface{}
	}{
		{(*io.Reader)(nil), new(bytes.Buffer)},
	}

	for i, tc := range trueCases {
		True(t, Implements(mockT, tc.iface, tc.value), "Implements should return true for trueCases(%d)", i)
	}

	falseCases := []struct {
		iface, value interface{}
	}{
		{(*io.Reader)(nil), nil},
		{(*io.WriteCloser)(nil), new(bytes.Buffer)},
		{(*io.WriteCloser)(nil), nil},
		{nil, nil},
	}

	for i, fc := range falseCases {
		False(t, Implements(mockT, fc.iface, fc.value), "Implements should return false for falseCases(%d)", i)
	}
}

func TestEqual(t *testing.T) {
	mockT := new(testing.T)

	funcA := func() int { return 23 }
	funcB := func() int { return 23 }

	trueCases := []struct {
		expected, actual interface{}
	}{
		{"Hello World", "Hello World"},
		{[]byte("Hello World"), []byte("Hello World")},
		{0, 0},
		{0.0, 0.0},
		{nil, nil},
		{int32(0), int32(0)},
		{uint64(0), uint64(0)},
		{new(bytes.Buffer), new(bytes.Buffer)},
		{&struct{}{}, &struct{}{}},
		{funcA, funcA},
	}

	for i, tc := range trueCases {
		True(t, Equal(mockT, tc.expected, tc.actual), "Equal should return true for trueCases(%d)", i)
	}

	falseCases := []struct {
		expected, actual interface{}
	}{
		{"Hello, world!", ""},
		{"", "Hello, world!"},
		{"Hello World", []byte("Hello World")},
		{0, 0.0},
		{0.0, 0},
		{nil, 0},
		{0, nil},
		{map[string]string{}, "something"},
		{funcA, funcB},
		{map[string]string{"foo": "bar"}, "foo"},
		{map[string]string{"foo": "bar"}, "bar"},
	}

	for i, fc := range falseCases {
		False(t, Equal(mockT, fc.expected, fc.actual), "Equal should return false for falseCases(%d)", i)
	}
}

func TestEqualFormatting(t *testing.T) {
	for i, currCase := range []struct {
		expected string
		actual   string
		extras   []interface{}
		message  string
	}{
		{
			expected: "want",
			actual:   "got",
			message:  "\tassertions.go:[0-9]+:\\s+?Error Trace:\t(\\S+:[0-9]+\\s+)+?Error:\tExpect to be equal\\s+Diff:\\s+--- Expected\\s+\\+\\+\\+ Actual\\s+?",
		},
		{expected: "want", actual: "got", extras: []interface{}{"Hello, %s", "world!"}, message: "\tassertions.go:[0-9]+:\\s+Hello, world!\\s+Error Trace:\t(\\S+:[0-9]+\\s+)+?Error:\tExpect to be equal\\s+Diff:\\s+--- Expected\\s+\\+\\+\\+ Actual\\s+?"},
	} {
		mockT := &gospec{}

		Equal(mockT, currCase.expected, currCase.actual, currCase.extras...)
		Match(t, regexp.MustCompile(currCase.message), mockT.String(), "Case %d", i)
	}
}

func TestNotEqual(t *testing.T) {
	mockT := new(testing.T)

	funcA := func() int { return 23 }
	funcB := func() int { return 23 }

	trueCases := []struct {
		expected, actual interface{}
	}{
		{"Hello, world!", ""},
		{"", "Hello, world!"},
		{"Hello World", []byte("Hello World")},
		{0, 0.0},
		{0.0, 0},
		{nil, 0},
		{0, nil},
		{map[string]string{}, "something"},
		{funcA, funcB},
		{map[string]string{"foo": "bar"}, "foo"},
		{map[string]string{"foo": "bar"}, "bar"},
	}

	for i, tc := range trueCases {
		True(t, NotEqual(mockT, tc.expected, tc.actual), "NotEqual should return true for trueCases(%d)", i)
	}

	falseCases := []struct {
		expected, actual interface{}
	}{
		{"Hello World", "Hello World"},
		{[]byte("Hello World"), []byte("Hello World")},
		{0, 0},
		{0.0, 0.0},
		{nil, nil},
		{int32(0), int32(0)},
		{uint64(0), uint64(0)},
		{new(bytes.Buffer), new(bytes.Buffer)},
		{&struct{}{}, &struct{}{}},
		{funcA, funcA},
	}

	for i, fc := range falseCases {
		False(t, NotEqual(mockT, fc.expected, fc.actual), "NotEqual should return false for falseCases(%d)", i)
	}
}

func TestEqualValues(t *testing.T) {
	mockT := new(testing.T)

	funcA := func() int { return 23 }
	funcB := func() int { return 23 }

	trueCases := []struct {
		expected, actual interface{}
	}{
		{"Hello World", "Hello World"},
		{[]byte("Hello World"), []byte("Hello World")},
		{"Hello World", []byte("Hello World")},
		{[]byte("Hello World"), "Hello World"},
		{0, 0},
		{0.0, 0.0},
		{0, 0.0},
		{0.0, 0},
		{nil, nil},
		{int32(0), int32(0)},
		{uint64(0), uint64(0)},
		{new(bytes.Buffer), new(bytes.Buffer)},
		{&struct{}{}, &struct{}{}},
		{funcA, funcA},
	}

	for i, tc := range trueCases {
		True(t, EqualValues(mockT, tc.expected, tc.actual), "EqualValues should return true for trueCases(%d)", i)
	}

	falseCases := []struct {
		expected, actual interface{}
	}{
		{"Hello, world!", ""},
		{"", "Hello, world!"},
		{"Hello World", "Hello World!"},
		{"Hello World!", "Hello World"},
		{"Hello World", []byte("Hello World!")},
		{[]byte("Hello World!"), "Hello World"},
		{nil, 0},
		{0, nil},
		{map[string]string{}, "something"},
		{funcA, funcB},
		{map[string]string{"foo": "bar"}, "foo"},
		{map[string]string{"foo": "bar"}, "bar"},
	}

	for i, fc := range falseCases {
		False(t, EqualValues(mockT, fc.expected, fc.actual), "EqualValues should return false for falseCases(%d)", i)
	}
}

func TestEqualJSON(t *testing.T) {
	mockT := new(testing.T)

	trueCases := []struct {
		expected, actual string
	}{
		{`{"hello": "world", "foo": "bar"}`, `{"hello": "world", "foo": "bar"}`},
		{`{"hello": "world", "foo": "bar"}`, `{"foo": "bar", "hello": "world"}`},
		{`{"numeric": 1.5,"array": [{"foo": "bar"}, 1, "string", ["nested", "array", 5.5]],"hash": {"nested": "hash", "nested_slice": ["this", "is", "nested"]},"string": "foo"}`,
			`{"numeric": 1.5,"hash": {"nested": "hash", "nested_slice": ["this", "is", "nested"]},"string": "foo","array": [{"foo": "bar"}, 1, "string", ["nested", "array", 5.5]]}`},
		{`["foo", {"hello": "world", "nested": "hash"}]`, `["foo", {"nested": "hash", "hello": "world"}]`},
	}
	for i, tc := range trueCases {
		True(t, EqualJSON(mockT, tc.expected, tc.actual), "EqualJSON should return true for trueCases(%d)", i)
	}

	falseCases := []struct {
		expected, actual string
	}{
		{`{"hello": "bar", "foo": "world"}`, `{"hello": "world", "foo": "bar"}`},
		{`{"foo": "bar"}`, `{"foo": "bar", "hello": "world"}`},
		{`["foo", {"hello": "world", "nested": "hash"}]`, `{"foo": "bar", {"nested": "hash", "hello": "world"}}`},
		{`["foo", {"hello": "world", "nested": "hash"}]`, `[{ "hello": "world", "nested": "hash"}, "foo"]`},
		{`{"foo": "bar"}`, "Not JSON"},
		{"Not JSON", `{"foo": "bar", "hello": "world"}`},
		{"Not JSON", "Not JSON"},
	}
	for i, fc := range falseCases {
		False(t, EqualJSON(mockT, fc.expected, fc.actual), "EqualJSON should return false for falseCases(%d)", i)
	}
}

func TestExactly(t *testing.T) {
	mockT := new(testing.T)

	a := float32(1)
	b := float64(1)
	c := float32(1)
	d := float32(2)
	funcA := func() int { return 23 }
	funcB := func() int { return 23 }

	trueCases := []struct {
		expected, actual interface{}
	}{
		{"Hello World", "Hello World"},
		{[]byte("Hello World"), []byte("Hello World")},
		{a, c},
		{nil, nil},
		{0, 0},
		{0.0, 0.0},
	}

	for i, tc := range trueCases {
		True(t, Exactly(mockT, tc.expected, tc.actual), "Exactly should return true for trueCases(%d)", i)
	}

	falseCases := []struct {
		expected, actual interface{}
	}{
		{"Hello World", "Hello World!"},
		{"Hello World!", "Hello World"},
		{[]byte("Hello World"), []byte("Hello World!")},
		{[]byte("Hello World!"), []byte("Hello World")},
		{a, b},
		{a, d},
		{nil, a},
		{a, nil},
		{funcA, funcA},
		{funcA, funcB},
	}

	for i, fc := range falseCases {
		False(t, Exactly(mockT, fc.expected, fc.actual), "Exactly should return false for fase case(%d)", i)
	}
}

func TestNil(t *testing.T) {
	mockT := new(testing.T)

	trueCases := []struct {
		value interface{}
	}{
		{nil},
		{(*struct{})(nil)},
		{(*io.Reader)(nil)},
	}

	for i, tc := range trueCases {
		True(t, Nil(mockT, tc.value), "Nil should return true for trueCases(%d)", i)
	}

	falseCases := []struct {
		value interface{}
	}{
		{""},
		{0},
		{0.0},
		{new(bytes.Buffer)},
		{func() int { return 23 }},
	}

	for i, fc := range falseCases {
		False(t, Nil(mockT, fc.value), "Nil should return false for falseCases(%d)", i)
	}
}

func TestNotNil(t *testing.T) {
	mockT := new(testing.T)

	trueCases := []struct {
		value interface{}
	}{
		{""},
		{0},
		{0.0},
		{new(bytes.Buffer)},
		{func() int { return 23 }},
	}

	for i, tc := range trueCases {
		True(t, NotNil(mockT, tc.value), "NotNil should return false for trueCases(%d)", i)
	}

	falseCases := []struct {
		value interface{}
	}{
		{nil},
		{(*struct{})(nil)},
		{(*io.Reader)(nil)},
	}

	for i, fc := range falseCases {
		False(t, NotNil(mockT, fc.value), "NotNil should return true for falseCases(%d)", i)
	}
}

func TestTrue(t *testing.T) {
	mockT := new(testing.T)

	if !True(mockT, true) {
		t.Error("True should return true")
	}

	if True(mockT, false) {
		t.Error("True should return false")
	}
	if True(mockT, 0) {
		t.Error("True should return false")
	}
	if True(mockT, 0.0) {
		t.Error("True should return false")
	}
	if True(mockT, nil) {
		t.Error("True should return false")
	}
	if True(mockT, new(bytes.Buffer)) {
		t.Error("True should return false")
	}
	if True(mockT, func() int { return 23 }) {
		t.Error("True should return false")
	}
}

func TestFalse(t *testing.T) {
	mockT := new(testing.T)

	if !False(mockT, false) {
		t.Error("False should return true")
	}

	if False(mockT, true) {
		t.Error("False should return false")
	}
	if False(mockT, 0) {
		t.Error("False should return false")
	}
	if False(mockT, 0.0) {
		t.Error("False should return false")
	}
	if False(mockT, nil) {
		t.Error("False should return false")
	}
	if False(mockT, new(bytes.Buffer)) {
		t.Error("False should return false")
	}
	if False(mockT, func() int { return 23 }) {
		t.Error("False should return false")
	}
}

func TestZero(t *testing.T) {
	mockT := new(testing.T)

	for i, v := range zeros {
		True(t, Zero(mockT, v), "Zero should return true for zeros(%d)", i)
	}

	for i, v := range nonZeros {
		False(t, Zero(mockT, v), "Zero should return false for nonZeros(%d)", i)
	}
}

func TestNotZero(t *testing.T) {
	mockT := new(testing.T)

	for i, v := range nonZeros {
		True(t, NotZero(mockT, v), "NotZero should return true for nonZeros(%d)", i)
	}

	for i, v := range zeros {
		False(t, NotZero(mockT, v), "NotZero should return false for zeros(%d)", i)
	}
}

func TestEmpty(t *testing.T) {
	mockT := new(testing.T)

	chWithValue := make(chan struct{}, 1)
	chWithValue <- struct{}{}

	chWithoutValue := make(chan struct{}, 1)

	var (
		sp  *string
		tp  *time.Time
		tnp time.Time
		est = struct {
			Name string
			age  int
		}{}
		estp = &est
		st   = struct {
			Name string
			age  int
		}{"", 1}
		stp = &st
		f   func() int
	)

	True(t, Empty(mockT, ""), "Empty string is empty")
	True(t, Empty(mockT, nil), "Nil is empty")
	True(t, Empty(mockT, 0), "Zero int value is empty")
	True(t, Empty(mockT, false), "False value is empty")
	True(t, Empty(mockT, []string{}), "Empty string array is empty")
	True(t, Empty(mockT, chWithoutValue), "Channel without values is empty")
	True(t, Empty(mockT, sp), "Nil string pointer is empty")
	True(t, Empty(mockT, tp), "Nil time.Time pointer is empty")
	True(t, Empty(mockT, tnp), "time.Time is empty")
	True(t, Empty(mockT, est), "Empty struct is empty")
	True(t, Empty(mockT, estp), "Empty struct pointer is empty")
	True(t, Empty(mockT, f), "Empty func type is empty")

	False(t, Empty(mockT, "something"), "Non Empty string is not empty")
	False(t, Empty(mockT, 1), "Non-zero int value is not empty")
	False(t, Empty(mockT, true), "True value is not empty")
	False(t, Empty(mockT, errors.New("something")), "Non nil object is not empty")
	False(t, Empty(mockT, []string{"something"}), "Non empty string array is not empty")
	False(t, Empty(mockT, st), "Struct value is not empty")
	False(t, Empty(mockT, stp), "Struct pointer value is not empty")
	False(t, Empty(mockT, chWithValue), "Channel with values is not empty")
	False(t, Empty(mockT, func() int { return 23 }), "Func value is not empty")
}

func TestNotEmpty(t *testing.T) {
	mockT := new(testing.T)

	chWithValue := make(chan struct{}, 1)
	chWithValue <- struct{}{}

	chWithoutValue := make(chan struct{}, 1)

	var (
		sp  *string
		tp  *time.Time
		tnp time.Time
		est = struct {
			Name string
			age  int
		}{}
		estp = &est
		st   = struct {
			Name string
			age  int
		}{"", 1}
		stp = &st
		f   func() int
	)

	False(t, NotEmpty(mockT, ""), "Empty string is empty")
	False(t, NotEmpty(mockT, nil), "Nil is empty")
	False(t, NotEmpty(mockT, 0), "Zero int value is empty")
	False(t, NotEmpty(mockT, false), "False value is empty")
	False(t, NotEmpty(mockT, []string{}), "Empty string array is empty")
	False(t, NotEmpty(mockT, chWithoutValue), "Channel without values is empty")
	False(t, NotEmpty(mockT, sp), "Nil string pointer is empty")
	False(t, NotEmpty(mockT, tp), "Nil time.Time pointer is empty")
	False(t, NotEmpty(mockT, tnp), "time.Time is empty")
	False(t, NotEmpty(mockT, est), "Empty struct is empty")
	False(t, NotEmpty(mockT, estp), "Empty struct pointer is empty")
	False(t, NotEmpty(mockT, f), "Empty func type is empty")

	True(t, NotEmpty(mockT, "something"), "Non Empty string is not empty")
	True(t, NotEmpty(mockT, 1), "Non-zero int value is not empty")
	True(t, NotEmpty(mockT, true), "True value is not empty")
	True(t, NotEmpty(mockT, errors.New("something")), "Non nil object is not empty")
	True(t, NotEmpty(mockT, []string{"something"}), "Non empty string array is not empty")
	True(t, NotEmpty(mockT, st), "Struct value is not empty")
	True(t, NotEmpty(mockT, stp), "Struct pointer value is not empty")
	True(t, NotEmpty(mockT, chWithValue), "Channel with values is not empty")
	True(t, NotEmpty(mockT, func() int { return 23 }), "Func value is not empty")
}

type kv struct {
	Name, Value string
}

func TestContains(t *testing.T) {
	mockT := new(testing.T)

	list := []string{"Foo", "Bar"}
	complexList := []*kv{
		{"b", "c"},
		{"d", "e"},
		{"g", "h"},
		{"j", "k"},
	}
	simpleMap := map[interface{}]interface{}{"Foo": "Bar"}
	ch := make(chan string, 1)
	ch <- "Foo"

	trueCases := []struct {
		value, element interface{}
	}{
		{"Hello World", "Hello"},
		{list, "Bar"},
		{complexList, &kv{"g", "h"}},
		{simpleMap, "Foo"},
		{ch, "Foo"},
	}

	for i, tc := range trueCases {
		True(t, Contains(mockT, tc.value, tc.element), "Contains should return true for trueCases(%d)", i)
	}

	falseCases := []struct {
		value, element interface{}
	}{
		{"Hello World", "Salut"},
		{nil, 0},
		{0, nil},
		{nil, 0.0},
		{0.0, nil},
		{nil, nil},
		{0, 0.0},
		{0.0, 0},
		{list, "Salut"},
		{complexList, &kv{"g", "e"}},
		{simpleMap, "Bar"},
		{ch, "Bar"},
		{func() int { return 23 }, 23},
	}

	for i, fc := range falseCases {
		False(t, Contains(mockT, fc.value, fc.element), "Contains should return false for falseCases(%d)", i)
	}
}

func TestNotContains(t *testing.T) {
	mockT := new(testing.T)

	list := []string{"Foo", "Bar"}
	complexList := []*kv{
		{"b", "c"},
		{"d", "e"},
		{"g", "h"},
		{"j", "k"},
	}
	simpleMap := map[interface{}]interface{}{"Foo": "Bar"}
	ch := make(chan string, 1)
	ch <- "Foo"

	trueCases := []struct {
		value, element interface{}
	}{
		{"Hello World", "Salut"},
		{nil, 0},
		{0, nil},
		{nil, 0.0},
		{0.0, nil},
		{nil, nil},
		{0, 0.0},
		{0.0, 0},
		{list, "Salut"},
		{complexList, &kv{"g", "e"}},
		{simpleMap, "Bar"},
		{ch, "Bar"},
		{func() int { return 23 }, 23},
	}

	for i, tc := range trueCases {
		True(t, NotContains(mockT, tc.value, tc.element), "NotContains should return true for trueCases(%d)", i)
	}

	falseCases := []struct {
		value, element interface{}
	}{
		{"Hello World", "Hello"},
		{list, "Bar"},
		{complexList, &kv{"g", "h"}},
		{simpleMap, "Foo"},
		{ch, "Foo"},
	}

	for i, fc := range falseCases {
		False(t, NotContains(mockT, fc.value, fc.element), "NotContains should return false for falseCases(%d)", i)
	}
}

func TestContainsForReader(t *testing.T) {
	mockT := new(testing.T)

	reader := strings.NewReader("Hello, World")

	True(t, Contains(mockT, reader, "Hello"), "Contains should return true")
	False(t, Contains(mockT, reader, "Salut"), "Contains should return false")
}

func TestNotContainsForReader(t *testing.T) {
	mockT := new(testing.T)

	reader := strings.NewReader("Hello, World")

	True(t, NotContains(mockT, reader, "Salut"), "NotContains should return true")
	False(t, NotContains(mockT, reader, "Hello"), "NotContains should return false")
}

func TestMatch(t *testing.T) {
	mockT := new(testing.T)

	trueCases := []struct {
		reg    interface{}
		values []interface{}
	}{
		{`^\d+$`,
			[]interface{}{
				0, 0.0, 1, 2.0,
			}},
		{regexp.MustCompile("^Hello"),
			[]interface{}{
				"Hello", "Hello, world!",
			}},
	}

	for i, tc := range trueCases {
		for j, v := range tc.values {
			True(t, Match(mockT, tc.reg, v), "Match should return true for trueCases(%d, %d)", i, j)
		}
	}

	falseCases := []struct {
		reg    interface{}
		values []interface{}
	}{
		{`^\d+$`,
			[]interface{}{
				"", nil, true, false, new(bytes.Buffer),
			}},
		{regexp.MustCompile("^Hello"),
			[]interface{}{
				"hello, world!", "hi, Hello", nil, true, false, new(bytes.Buffer),
			}},
	}

	for i, fc := range falseCases {
		for j, v := range fc.values {
			False(t, Match(mockT, fc.reg, v), "Match should return false for falseCases(%d, %d)", i, j)
		}
	}
}

func TestNotMatch(t *testing.T) {
	mockT := new(testing.T)

	trueCases := []struct {
		reg    interface{}
		values []interface{}
	}{
		{`^\d+$`,
			[]interface{}{
				"", nil, true, false, new(bytes.Buffer),
			}},
		{regexp.MustCompile("^Hello"),
			[]interface{}{
				"hello, world!", "hi, Hello", nil, true, false, new(bytes.Buffer),
			}},
	}

	for i, tc := range trueCases {
		for j, v := range tc.values {
			True(t, NotMatch(mockT, tc.reg, v), "NotMatch should return true for trueCases(%d, %d)", i, j)
		}
	}

	falseCases := []struct {
		reg    interface{}
		values []interface{}
	}{
		{`^\d+$`,
			[]interface{}{
				0, 0.0, 1, 2.0,
			}},
		{regexp.MustCompile("^Hello"),
			[]interface{}{
				"Hello", "Hello, world!",
			}},
	}

	for i, fc := range falseCases {
		for j, v := range fc.values {
			False(t, NotMatch(mockT, fc.reg, v), "NotMatch should return flase for falseCases(%d, %d)", i, j)
		}
	}
}

func TestCondition(t *testing.T) {
	mockT := new(testing.T)

	if !Condition(mockT, func() bool { return true }, "Truth") {
		t.Error("Condition should return true")
	}

	if Condition(mockT, func() bool { return false }, "Lie") {
		t.Error("Condition should return false")
	}
}

func TestLen(t *testing.T) {
	mockT := new(testing.T)

	ch := make(chan int, 5)
	ch <- 1
	ch <- 2
	ch <- 3

	trueCases := []struct {
		v interface{}
		l int
	}{
		{[]int{1, 2, 3}, 3},
		{[...]int{1, 2, 3}, 3},
		{"ABC", 3},
		{map[int]int{1: 2, 2: 4, 3: 6}, 3},
		{ch, 3},

		{[]int{}, 0},
		{map[int]int{}, 0},
		{make(chan int), 0},

		{[]int(nil), 0},
		{map[int]int(nil), 0},
		{(chan int)(nil), 0},
	}

	for i, tc := range trueCases {
		True(t, Len(mockT, tc.v, tc.l), "Len should return true for trueCase(%d)", i)
	}

	falseCases := []struct {
		v interface{}
		l int
	}{
		{[]int{1, 2, 3}, 4},
		{[...]int{1, 2, 3}, 2},
		{"ABC", 2},
		{map[int]int{1: 2, 2: 4, 3: 6}, 4},
		{ch, 2},

		{[]int{}, 1},
		{map[int]int{}, 1},
		{make(chan int), 1},

		{[]int(nil), 1},
		{map[int]int(nil), 1},
		{(chan int)(nil), 1},

		{nil, 0},
		{0, 0},
		{0.0, 0},
		{true, 0},
		{false, 0},
		{' ', 0},
		{'0', 0},
		{struct{}{}, 0},
	}

	for _, fc := range falseCases {
		False(t, Len(mockT, fc.v, fc.l), "Len should return false for falseCases(%d)", i)
	}
}

func TestInDelta(t *testing.T) {
	mockT := new(testing.T)

	trueCases := []struct {
		a, b  interface{}
		delta float64
	}{
		{uint8(2), uint8(1), 1},
		{uint16(2), uint16(1), 1},
		{uint32(2), uint32(1), 1},
		{uint64(2), uint64(1), 1},

		{uint8(1), uint8(2), 1},
		{uint16(1), uint16(2), 1},
		{uint32(1), uint32(2), 1},
		{uint64(1), uint64(2), 1},

		{int(2), int(1), 1},
		{int8(2), int8(1), 1},
		{int16(2), int16(1), 1},
		{int32(2), int32(1), 1},
		{int64(2), int64(1), 1},

		{int(1), int(2), 1},
		{int8(1), int8(2), 1},
		{int16(1), int16(2), 1},
		{int32(1), int32(2), 1},
		{int64(1), int64(2), 1},

		{float32(2), float32(1), 1},
		{float64(2), float64(1), 1},

		{float32(1), float32(2), 1},
		{float64(1), float64(2), 1},
	}

	for i, tc := range trueCases {
		True(t, InDelta(mockT, tc.a, tc.b, tc.delta), "InDelta should return true for trueCases(%d)", i)
	}

	falseCases := []struct {
		a, b  interface{}
		delta float64
	}{
		{1, 2, 0.5},
		{2, 1, 0.5},
		{"", 0, 1},
		{0, "", 1},
		{nil, 0, 1},
		{0, nil, 1},
		{nil, nil, 1},
		{"", "", 1},
		{0, math.NaN(), 1},
		{math.NaN(), 0, 1},
		{"", math.NaN(), 1},
		{math.NaN(), "", 1},
		{nil, math.NaN(), 1},
		{math.NaN(), nil, 1},
	}

	for i, fc := range falseCases {
		False(t, InDelta(mockT, fc.a, fc.b, fc.delta), "InDelta should return false for falseCases(%d)", i)
	}
}

func TestWithinDuration(t *testing.T) {
	mockT := new(testing.T)

	a := time.Now()
	b := a.Add(10 * time.Second)

	trueCases := []struct {
		a, b     time.Time
		duration time.Duration
	}{
		{a, b, 10 * time.Second},
		{b, a, 10 * time.Second},
		{a, b, 11 * time.Second},
		{b, a, 11 * time.Second},
	}

	for i, tc := range trueCases {
		True(t, WithinDuration(mockT, tc.a, tc.b, tc.duration), "A 10s difference is within a %v time difference for trueCases(%d)", tc.duration, i)
	}

	falseCases := []struct {
		a, b     time.Time
		duration time.Duration
	}{
		{a, b, 9 * time.Second},
		{b, a, 9 * time.Second},
		{a, b, -9 * time.Second},
		{b, a, -9 * time.Second},
		{a, b, -11 * time.Second},
		{b, a, -11 * time.Second},
	}

	for i, fc := range falseCases {
		False(t, WithinDuration(mockT, fc.a, fc.b, fc.duration), "A 10s difference is not within a %v time difference for falseCases(%d)", fc.duration, i)
	}
}

type customError struct{}

func (*customError) Error() string { return "fail" }

func TestError(t *testing.T) {
	mockT := new(testing.T)

	// start with a nil error
	var err error
	False(t, Error(mockT, err), "Error should return false for nil")

	// now set an error
	err = errors.New("some error")
	True(t, Error(mockT, err), "Error should return true")

	// an empty error interface
	err = func() error {
		var err *customError
		if err != nil {
			t.Fatal("err should be nil here")
		}
		return err
	}()
	if err == nil { // err is not nil here!
		t.Errorf("Error %#v should not be nil due to empty interface", err)
	}

	True(t, Error(mockT, err), "Error should pass with empty error interface")

	falseCases := []struct {
		v interface{}
	}{
		{""},
		{0},
		{0.0},
		{new(bytes.Buffer)},
		{func() int { return 23 }},
		{func() error { return errors.New("") }},
	}
	for i, fc := range falseCases {
		False(t, Error(mockT, fc.v), "Error should return false for falseCases(%d)", i)
	}
}

func TestNotError(t *testing.T) {
	mockT := new(testing.T)

	// start with a nil error
	var err error
	True(t, NotError(mockT, err), "NotError should return True for nil")

	// now set an error
	err = errors.New("some error")
	False(t, NotError(mockT, err), "NotError with error should return False")

	// returning an empty error interface
	err = func() error {
		var err *customError
		if err != nil {
			t.Fatal("err should be nil here")
		}
		return err
	}()
	if err == nil { // err is not nil here!
		t.Errorf("Error should be nil due to empty interface", err)
	}

	False(t, NotError(mockT, err), "NotError should fail with empty error interface")

	trueCases := []struct {
		v interface{}
	}{
		{""},
		{0},
		{0.0},
		{new(bytes.Buffer)},
		{func() int { return 23 }},
		{func() error { return errors.New("") }},
	}
	for i, fc := range trueCases {
		True(t, NotError(mockT, fc.v), "NotError should return true for trueCases(%d)", i)
	}
}

func TestEqualErrors(t *testing.T) {
	mockT := new(testing.T)

	// start with a nil error
	var err error
	False(t, EqualErrors(mockT, err, ""), "EqualErrors should return false for nil")

	// now set an error
	err = errors.New("some error")
	True(t, EqualErrors(mockT, err, "some error"), "EqualErrors should return true")
	False(t, EqualErrors(mockT, err, "Not some error"), "EqualErrors should return false for different error string")
}

func TestPanics(t *testing.T) {
	mockT := new(testing.T)

	True(t, Panics(mockT, func() {
		panic("Panic!")
	}), "Panics should return true")

	False(t, Panics(mockT, func() {}), "Panics should return false")
}

func TestNotPanics(t *testing.T) {
	mockT := new(testing.T)

	True(t, NotPanics(mockT, func() {}), "NotPanics should return true")
	False(t, NotPanics(mockT, func() {
		panic("Panic!")
	}), "NotPanics should return false")
}

func TestJSONContains(t *testing.T) {
	mockT := new(testing.T)

	trueCases := []struct {
		data, key string
	}{
		{`{"hello": "world", "foo": "bar"}`, `hello`},
		{`{"hello": "world", "foo": "bar"}`, `foo`},
		{`{"numeric": 1.5, "array": [{"foo": "bar"}, 1, "string", ["nested", "array", 5.5]],"hash": {"nested": "hash", "nested_slice": ["this", "is", "nested"]},"string": "foo"}`,
			`array.0.foo`},
		{`["foo", {"hello": "world", "nested": "hash"}]`, `1.hello`},
	}
	for i, tc := range trueCases {
		True(t, JSONContains(mockT, tc.data, tc.key), "JSONContains should return true for trueCases(%d)", i)
	}

	falseCases := []struct {
		data, key string
	}{
		{`{"hello": "bar", "foo": "world"}`, `world`},
		{`{"foo": "bar"}`, `hello`},
		{`["foo", {"hello": "world", "nested": "hash"}]`, `0.hello`},
		{`{"foo": "bar"}`, "Not JSON"},
		{"Not JSON", `Not`},
		{"Not JSON", "Not JSON"},
	}
	for i, fc := range falseCases {
		False(t, JSONContains(mockT, fc.data, fc.key), "JSONContains should return false for falseCases(%d)", i)
	}
}

func TestJSONEqualValues(t *testing.T) {
	mockT := new(testing.T)

	trueCases := []struct {
		data, key string
		value     interface{}
	}{
		{`{"hello": "world", "foo": "bar"}`, `hello`, "world"},
		{`{"hello": "world", "foo": 123.0}`, `foo`, 123.0},
		{`{"numeric": 1.5, "array": [{"foo": "bar"}, 1, "string", ["nested", "array", 5.5]],"hash": {"nested": "hash", "nested_slice": ["this", "is", "nested"]},"string": "foo"}`,
			`array.0.foo`, "bar"},
		{`["foo", {"hello": "world", "nested": "hash"}]`, `1.hello`, "world"},
		{`["foo", {"hello": "world", "nested": 123}]`, `1.nested`, 123},
		{`["foo", {"hello": "world", "nested": true}]`, `1.nested`, true},
	}
	for i, tc := range trueCases {
		True(t, JSONEqualValues(t, tc.data, tc.key, tc.value), "JSONEqualValues should return true for trueCases(%d)", i)
	}

	falseCases := []struct {
		data, key string
		value     interface{}
	}{
		{`{"hello": "bar", "foo": "world"}`, `hello`, "world"},
		{`{"foo": "bar"}`, `hello`, "bar"},
		{`["foo", {"hello": "world", "nested": "hash"}]`, `2.hello`, ""},
		{`{"foo": "bar"}`, "Not JSON", ""},
		{"Not JSON", `Not`, ""},
		{"Not JSON", "Not JSON", ""},
	}
	for i, fc := range falseCases {
		False(t, JSONEqualValues(mockT, fc.data, fc.key, fc.value), "JSONEqualValues should return false for falseCases(%d)", i)
	}
}
