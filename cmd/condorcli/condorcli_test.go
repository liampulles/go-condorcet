package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRun_GivenMixedCase_ShouldPrintInvalidVotesAndElectionResults(t *testing.T) {
	// Setup fixture
	input := strings.NewReader(`A,C,B,E,D
A,C,B,E,D
C,A,E,B,D
B,E,D,A,C
D,C,E,B,A
E,B,A,D,C
B,E,D,A,C
D,C,E,B,A
A,D,E,C,B
D,C,E,B,A
B,E,D,A,C
A,A
D,C,E,B,A
D,C,E,B,A
C,A,E,B,D
C,A,E,B,D
B,E,D,A,C
A,D,E,C,B
C,B,A,D,E
A,D,E,C,B
C,A,E,B,D
C,A,E,B,D
A,C,B,E,D
C,A,B,E,D
E,B,A,D,C
C,A,E,B,D
E,B,A,D,C
D,C,E,B,A
B,E,D,A,C
C,B,A,D,E
A,C,B,E,D
A,D,E,C,B
E,B,A,D,C
E,B,A,D,C
C,A,E,B,D
B,E,D,A,C
B,E,D,A,C
E,B,A,D,C
A,D,E,C,B
D,C,E,B,A
B,E,D,A,C
E,B,A,D,C
E,B,A,D,C
C,A,B,E,D
C,A,B,E,D
A,C,B,E,D`)
	output := strings.Builder{}
	errOutput := strings.Builder{}

	// Setup expectations
	expectedOutput := `E
A
C
B
D
`
	expectedErr := "ERROR [Line 11]: could not read record: could not add equal elements: cyclic vote detected: cannot reference a candidate twice in a vote\n"

	// Exercise SUT
	run(input, &output, &errOutput)

	// Verify results
	assert.Equal(t, expectedOutput, output.String())
	assert.Equal(t, expectedErr, errOutput.String())
}
