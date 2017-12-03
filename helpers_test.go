package gospec

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
	"time"
)

func TestDeepEqual(t *testing.T) {
	trueCases := []struct {
		expect, actual interface{}
	}{
		{"Hello World", "Hello World"},
		{[]byte("Hello World"), []byte("Hello World")},
		{0, 0},
		{0.0, 0.0},
		{nil, nil},
	}

	for i, tc := range trueCases {
		True(t, DeepEqual(tc.expect, tc.actual), "DeepEqual shoud return true for true case(%d)", i)
	}

	falseCases := []struct {
		expect, actual interface{}
	}{
		{"Hello World", "Hello, World!"},
		{[]byte("Hello World"), []byte("Hello, World!")},
		{0, 0.0},
		{0.0, 0},
		{uint32(0), int32(0)},
		{int32(0), uint32(0)},
		{uint32(1), int32(0)},
		{int32(0), uint32(1)},
		{nil, 0},
		{0, nil},
		{map[int]int{10: 10}, map[int]int{10: 20}},
		{0, []byte("0")},
		{[]byte("0"), 0},
		{'x', "x"},
		{"x", 'x'},
	}

	for i, fc := range falseCases {
		False(t, DeepEqual(fc.expect, fc.actual), "DeepEqual shoud return true for false case(%d)", i)
	}
}

func TestDeepEqualValues(t *testing.T) {
	trueCases := []struct {
		expect, actual interface{}
	}{
		{"Hello World", "Hello World"},
		{[]byte("Hello World"), []byte("Hello World")},
		{0, 0},
		{0.0, 0.0},
		{0, 0.0},
		{0.0, 0},
		{uint32(0), int32(0)},
		{int32(0), uint32(0)},
		{nil, nil},
		{'x', "x"},
		{"x", 'x'},
	}

	for i, tc := range trueCases {
		True(t, DeepEqualValues(tc.expect, tc.actual), "DeepEqualValues shoud return true for true case(%d)", i)
	}

	falseCases := []struct {
		expect, actual interface{}
	}{
		{"Hello World", "Hello, World!"},
		{[]byte("Hello World"), []byte("Hello, World!")},
		{0, 1},
		{0.0, 1.0},
		{uint32(1), int32(0)},
		{int32(0), uint32(1)},
		{nil, 0},
		{0, nil},
		{map[int]int{10: 10}, map[int]int{10: 20}},
		{0, []byte("0")},
		{[]byte("0"), 0},
	}

	for i, fc := range falseCases {
		False(t, DeepEqualValues(fc.expect, fc.actual), "DeepEqualValues shoud return true for false case(%d)", i)
	}
}

func TestIsNil(t *testing.T) {
	var (
		s     []int
		m     map[string]string
		ch    chan int
		iface interface{}
		f     func()
		p     *int

		st struct{}
	)

	True(t, IsNil(nil))
	True(t, IsNil(s))
	True(t, IsNil(m))
	True(t, IsNil(ch))
	True(t, IsNil(iface))
	True(t, IsNil(f))
	True(t, IsNil(p))

	False(t, IsNil(""))
	False(t, IsNil(0))
	False(t, IsNil(0.0))
	False(t, IsNil([]int{}))
	False(t, IsNil(map[string]string{}))
	False(t, IsNil(make(chan int)))
	False(t, IsNil(func() { return }))
	False(t, IsNil(new(int)))
	False(t, IsNil(st))
}

func TestIsEmpty(t *testing.T) {
	chWithoutValue := make(chan struct{})

	chWithValue := make(chan struct{}, 1)
	chWithValue <- struct{}{}

	True(t, IsEmpty(""))
	True(t, IsEmpty(nil))
	True(t, IsEmpty([]string{}))
	True(t, IsEmpty(0))
	True(t, IsEmpty(int32(0)))
	True(t, IsEmpty(float32(0)))
	True(t, IsEmpty(false))
	True(t, IsEmpty(map[string]string{}))
	True(t, IsEmpty(time.Time{}))
	True(t, IsEmpty(new(time.Time)))
	True(t, IsEmpty(chWithoutValue))

	False(t, IsEmpty("something"))
	False(t, IsEmpty(errors.New("something")))
	False(t, IsEmpty([]string{"something"}))
	False(t, IsEmpty(1))
	False(t, IsEmpty(int32(1)))
	False(t, IsEmpty(float32(1)))
	False(t, IsEmpty(true))
	False(t, IsEmpty(map[string]string{"Hello": "World"}))
	False(t, IsEmpty(time.Now()))
	False(t, IsEmpty(chWithValue))
}

type testingReader struct {
	s        string
	i        int64 // current reading index
	prevRune int   // index of previous rune; or < 0
}

func (tr *testingReader) Read(p []byte) (n int, err error) {
	if tr.i >= int64(len(tr.s)) {
		return 0, io.EOF
	}
	tr.prevRune = -1
	n = copy(p, tr.s[tr.i:])
	tr.i += int64(n)
	return
}

func TestContainsElement(t *testing.T) {
	list1 := []string{"Foo", "Bar"}
	list2 := []int{1, 2}
	list3 := map[interface{}]interface{}{"Foo": "Bar"}
	list4 := make(chan string, 1)
	list4 <- "Foo"

	True(t, ContainsElement("Hello World", "World"))
	False(t, ContainsElement("Hello World", "word"))
	False(t, ContainsElement(1234, "1"))
	False(t, ContainsElement("1234", 1))
	True(t, ContainsElement(list1, "Foo"))
	True(t, ContainsElement(list1, "Bar"))
	False(t, ContainsElement(list1, "Foo Bar"))
	False(t, ContainsElement(list1, "Boo"))
	True(t, ContainsElement(list2, 1))
	True(t, ContainsElement(list2, 2))
	False(t, ContainsElement(list2, 3))
	False(t, ContainsElement(list2, "1"))
	True(t, ContainsElement(list3, "Foo"))
	False(t, ContainsElement(list3, "Bar"))
	True(t, ContainsElement(list4, "Foo"))
	False(t, ContainsElement(list4, "Bar"))

	// for io.ReadSeeker
	rs := strings.NewReader("Hello, world!")
	True(t, ContainsElement(rs, "Hello"))
	True(t, ContainsElement(rs, "world"))
	False(t, ContainsElement(rs, "Foo Bar"))

	// for io.ReadWriter
	rw := bytes.NewBuffer([]byte("Hello, world!"))
	True(t, ContainsElement(rw, "Hello"))
	True(t, ContainsElement(rw, "world"))
	False(t, ContainsElement(rw, "Foo Bar"))

	r := &testingReader{
		s: "Hello, world!",
	}
	True(t, ContainsElement(r, "Hello"))
	False(t, ContainsElement(r, "world"))

	r = &testingReader{
		s: "Hello, world!",
	}
	True(t, ContainsElement(r, "world"))
	False(t, ContainsElement(r, "Hello"))
}

func Test_recovery(t *testing.T) {
	if didPanic, _ := recovery(func() {
		panic("Panic!")
	}); !didPanic {
		t.Error("didPanic should return true")
	}

	if didPanic, _ := recovery(func() {}); didPanic {
		t.Error("didPanic should return false")
	}
}

func Test_diff(t *testing.T) {
	result := `--- Expected
+++ Actual
@@ -1,2 +1,2 @@
-(string) (len=12) "Hello, world"
+(string) (len=5) "world"
 
`
	expected := "Hello, world"
	actual := "world"
	Equal(t, result, diff(expected, actual))

	result = `--- int(123)
+++ int(12)

`
	Equal(t, result, diff(123, 12))

	expected = `--- Expected
+++ Actual
@@ -1,3 +1,3 @@
 (struct { foo string }) {
- foo: (string) (len=5) "hello"
+ foo: (string) (len=3) "bar"
 }
`
	actual = diff(
		struct{ foo string }{"hello"},
		struct{ foo string }{"bar"},
	)
	Equal(t, expected, actual)

	expected = `--- Expected
+++ Actual
@@ -2,5 +2,5 @@
  (int) 1,
- (int) 2,
  (int) 3,
- (int) 4
+ (int) 5,
+ (int) 7
 }
`
	actual = diff(
		[]int{1, 2, 3, 4},
		[]int{1, 3, 5, 7},
	)
	Equal(t, expected, actual)

	expected = `--- Expected
+++ Actual
@@ -2,4 +2,4 @@
  (int) 1,
- (int) 2,
- (int) 3
+ (int) 3,
+ (int) 5
 }
`
	actual = diff(
		[]int{1, 2, 3, 4}[0:3],
		[]int{1, 3, 5, 7}[0:3],
	)
	Equal(t, expected, actual)

	expected = `--- Expected
+++ Actual
@@ -1,6 +1,6 @@
 (map[string]int) (len=4) {
- (string) (len=4) "four": (int) 4,
+ (string) (len=4) "five": (int) 5,
  (string) (len=3) "one": (int) 1,
- (string) (len=5) "three": (int) 3,
- (string) (len=3) "two": (int) 2
+ (string) (len=5) "seven": (int) 7,
+ (string) (len=5) "three": (int) 3
 }
`
	actual = diff(
		map[string]int{"one": 1, "two": 2, "three": 3, "four": 4},
		map[string]int{"one": 1, "three": 3, "five": 5, "seven": 7},
	)
	Equal(t, expected, actual)
}

func Test_diffWithEmptyCases(t *testing.T) {
	Equal(t, "", diff(nil, nil))
	Equal(t, "", diff(1, 1))
	Equal(t, "", diff("", ""))
	Equal(t, "--- int(0)\n+++ float64(0)\n\n", diff(0, 0.0))
	Equal(t, "--- int(1)\n+++ int(2)\n\n", diff(1, 2))
	Equal(t, "--- struct { foo string }({})\n+++ <nil>(<nil>)\n\n", diff(struct{ foo string }{}, nil))
	Equal(t, "--- <nil>(<nil>)\n+++ struct { foo string }({})\n\n", diff(nil, struct{ foo string }{}))
	Equal(t, "--- []int([1])\n+++ []bool([true])\n\n", diff([]int{1}, []bool{true}))
}

// Ensure there are no data races
func Test_diffWithRace(t *testing.T) {
	t.Parallel()

	expected := map[string]string{
		"a": "A",
		"b": "B",
		"c": "C",
	}

	actual := map[string]string{
		"d": "D",
		"e": "E",
		"f": "F",
	}

	// run diffs in parallel simulating tests with t.Parallel()
	numRoutines := 10
	rChans := make([]chan string, numRoutines)
	for idx := range rChans {
		rChans[idx] = make(chan string)
		go func(ch chan string) {
			defer close(ch)
			ch <- diff(expected, actual)
		}(rChans[idx])
	}

	for _, ch := range rChans {
		for msg := range ch {
			NotZero(t, msg) // dummy assert
		}
	}
}

func Test_tryMatch(t *testing.T) {
	trueCases := []struct {
		rx, str string
	}{
		{"^start", "start of the line"},
		{"end$", "in the end"},
		{"[0-9]{3}[.-]?[0-9]{2}[.-]?[0-9]{2}", "My phone number is 650.12.34"},
	}

	for i, tc := range trueCases {
		_, ok := tryMatch(tc.rx, tc.str)

		True(t, ok, "tryMatch should return true for true case(%d)", i)
	}

	falseCases := []struct {
		rx, str string
	}{
		{"^asdfastart", "Not the start of the line"},
		{"end$", "in the end."},
		{"[0-9]{3}[.-]?[0-9]{2}[.-]?[0-9]{2}", "My phone number is 650.12a.34"},
	}

	for i, fc := range falseCases {
		_, ok := tryMatch(fc.rx, fc.str)

		False(t, ok, "tryMatch should return false for false case(%d)", i)
	}
}

func Test_tryLen(t *testing.T) {
	falseCases := []interface{}{
		nil,
		0,
		true,
		false,
		'A',
		struct{}{},
	}
	for _, v := range falseCases {
		l, ok := tryLen(v)
		False(t, ok, "Expected tryLen fail to get length of %#v", v)
		Equal(t, 0, l, "tryLen should return 0 for %#v", v)
	}

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

	for _, c := range trueCases {
		l, ok := tryLen(c.v)
		True(t, ok, "Expected tryLen success to get length of %#v", c.v)
		Equal(t, c.l, l)
	}
}

func Test_toString(t *testing.T) {
	expected, actual := toString("foo", "bar")
	Equal(t, `"foo"`, expected, "value should not include type")
	Equal(t, `"bar"`, actual, "value should not include type")

	expected, actual = toString(123, 123)
	Equal(t, `123`, expected, "value should not include type")
	Equal(t, `123`, actual, "value should not include type")

	expected, actual = toString(int64(123), int32(123))
	Equal(t, `int64(123)`, expected, "value should include type")
	Equal(t, `int32(123)`, actual, "value should include type")

	expected, actual = toString(int64(123), nil)
	Equal(t, `<int64 Value>`, expected, "value should include type")
	Equal(t, `<nil>`, actual, "value should include type")

	expected, actual = toString(nil, int32(123))
	Equal(t, `<nil>`, expected, "value should include type")
	Equal(t, `<int32 Value>`, actual, "value should include type")

	type testStructType struct {
		Val string
	}

	expected, actual = toString(&testStructType{Val: "test"}, &testStructType{Val: "test"})
	Equal(t, `&gospec.testStructType{Val:"test"}`, expected, "value should not include type annotation")
	Equal(t, `&gospec.testStructType{Val:"test"}`, actual, "value should not include type annotation")

	toString(nil, nil)
}
