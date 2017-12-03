package gospec

import (
	"testing"
)

const (
	labelNewLine    = "\n\r\t"
	labelErrorTrace = "Error Trace"
	labelError      = "Error"
	labelMessages   = "Message"
)

type (
	// TestingT is the interface that wraps the basic Errorf method.
	//
	// Errorf formats its arguments according to the format, analogous to Printf, and
	// records the text in the error log. A final newline is added if not provided.
	TestingT interface {
		Errorf(format string, args ...interface{})
	}

	// TestingR is the interface that wraps the basic Run method.
	//
	// Run runs runner as a subtest of t called name. It reports whether runner succeeded. Run
	// runs runner in a separate goroutine and will block until all its parallel subtests
	// have completed.
	TestingR interface {
		Run(title string, runner func(t *testing.T)) bool
	}
)

type (
	// S defines a custom func that returns a new grammar with given value
	S func(actual interface{}) *grammar

	// Expectation defines a custom func that wraps testing examples
	Expectation func(example string, runner func(expect S))

	// Comparison defines a custom func that returns true on success and false on failure
	Comparison func() (success bool)

	// PanicRecover defines a func that should be passed to the assert.Panics and assert.NotPanics
	// methods, and represents a simple func that takes no arguments, and returns nothing.
	PanicRecover func()
)
