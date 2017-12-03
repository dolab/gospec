package gospec

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
)

type (
	// Comparison defines a custom func that returns true on success and false on failure
	Comparison func() (success bool)

	// PanicRecover defines a func that should be passed to the assert.Panics and assert.NotPanics
	// methods, and represents a simple func that takes no arguments, and returns nothing.
	PanicRecover func()
)
