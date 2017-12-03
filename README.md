## gospec

[![Build Status](https://travis-ci.org/dolab/gospec.svg?branch=master)](https://travis-ci.org/dolab/gospec)

> golang testing framework inspired from [testify](https://github.com/stretchr/testify)

Gospec allow you to easily write test code, and are global funcs in the `gospec` package.
All assertion functions take, as the first argument, the `*testing.T` object provided by the
testing framework. This allows the assertion funcs to write the failings and other details to
the correct place.

Every assertion function also takes an optional string message as the final argument,
allowing custom error messages to be appended to the message the assertion method outputs.


### Usage

```go
import (
    "testing"

    "github.com/golib/assert"
)

// use global assertions
func TestWithGlobalAssertions(t *testing.T) {
    var (
        a = "Hello"
        b = "Hello"
    )

    gospec.Equal(t, a, b, "The two words should be the same.")
}

// use assert instance
func TestWithAssertInstance(t *testing.T) {
    assert := gospec.NewAssertion(t)

    var (
        a = "Hello"
        b = "Hello"
    )

    assert.Equal(a, b, "The two words should be the same.")
}

// use expect instance
func TestWithExpectInstance(t *testing.T) {
    it := gospec.NewExpectation(t)

    var (
        a = "Hello"
        b = "Hello"
    )

    it("should be the same", func(expect *gospec.S){
        expect(a).Equal(b)
    })
}
```
