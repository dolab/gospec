package gospec

import (
	"bufio"
	"bytes"
	"strings"
)

type labeledOutput struct {
	label   string
	content string
}

// testingOutput returns a string consisting of the provided labeledOutput.
// Each labeled output is appended in the following manner:
//
//   \n\r\t{{label}}:{{align_spaces}}\t{{content}}\n
//
// The initial carriage return is required to undo/erase any padding added by testing.T.Errorf. And
// the "\t{{label}}:" is for the label.
// If a label is shorter than the longest label provided, padding spaces are added to make all the labels
// align to right. Once this alignment is achieved, "\t{{content}}\n" is added for the output.
//
// If the content of the testingOutput contains line breaks, the subsequent lines are aligned
// so that they start at the same location as the first line.
type testingOutput struct {
	labels  []labeledOutput
	padding int
}

func (output *testingOutput) Add(label labeledOutput) *testingOutput {
	output.labels = append(output.labels, label)

	return output
}

func (output *testingOutput) Padding(padding int) *testingOutput {
	output.padding = padding

	return output
}

func (output *testingOutput) LongestLabelLen() int {
	longestLabel := 0
	for _, label := range output.labels {
		if len(label.label) > longestLabel {
			longestLabel = len(label.label)
		}
	}

	return longestLabel
}

// formatMessages aligns the provided message so that all lines after the first line start at the same location as the first line.
// Assumes that the first line starts at the correct location (after carriage return, tab, label, spacer and tab).
// The longestLabelLen parameter specifies the length of the longest label in the output (required becaues this is the
// basis on which the alignment occurs).
func (output *testingOutput) formatMessages(message string, longestLabelLen int) string {
	buf := new(bytes.Buffer)

	for i, scanner := 0, bufio.NewScanner(strings.NewReader(message)); scanner.Scan(); i++ {
		// no need to align first line because it starts at the correct location (after the label)
		if i != 0 {
			// append alignLen+1 spaces to align with "{{longestLabel}}:" before adding tab
			buf.WriteString(labelNewLine + strings.Repeat(" ", longestLabelLen+1) + "\t")
		}

		buf.WriteString(scanner.Text())
	}

	return buf.String()
}

func (output *testingOutput) String() string {
	s := ""
	longestLabelLen := output.LongestLabelLen()

	// output message first if exists
	for _, label := range output.labels {
		if label.label != labelMessages || label.content == "" {
			continue
		}

		s += "\t" + output.formatMessages(label.content, longestLabelLen)
	}

	for _, label := range output.labels {
		if label.label == labelMessages {
			continue
		}

		nl := labelNewLine
		if output.padding-len(label.label) > 0 {
			nl += strings.Repeat(" ", output.padding-len(label.label))
		}

		s += nl
		s += label.label + ":"
		s += "\t" + output.formatMessages(strings.Replace(label.content, labelNewLine, nl, -1), output.padding)
	}

	return s + "\n\r"
}
