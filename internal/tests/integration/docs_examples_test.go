package integration

import (
	"testing"
)

// This file contains integration tests to make sure that (some of) the examples
// in the gomplate docs work correctly

func TestDocsExamples_DataExamples(t *testing.T) {
	inOutTest(t,
		"{{ $rows := (jsonArray `[[\"first\",\"second\"],[\"1\",\"2\"],[\"3\",\"4\"]]`) }}{{ data.ToCSV \";\" $rows }}",
		"first;second\r\n1;2\r\n3;4\r\n")
}
