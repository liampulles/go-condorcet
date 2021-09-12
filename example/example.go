//nolint
package example

import (
	"strings"

	gocondorcet "github.com/liampulles/go-condorcet"
)

func runElection() gocondorcet.Results {
	// Each vote is a map which indicates preference given to candidates (lower
	// is better). You can give candidates equal preference, use any variation
	// of values (what is important is the relative preference in a vote) and
	// omit candidates (which is effectively giving them equal last place).
	votes := []gocondorcet.Vote{
		{"DAN": 0, "ALICE": 1, "SALLY": 2},
		// -> This is equivalent to {"SALLY": 0, "DAN": 0, "ALICE": 0}
		{"SALLY": 5, "DAN": 5, "ALICE": 5},
		{"BOB": 0},
	}

	// The output is an ordered slice of CandidateIDs, first being the most
	// preferred and last being the least preferred (according to the Schulze
	// method).
	// In this case, it will be ["DAN","ALICE","SALLY","BOB"].
	return gocondorcet.Evaluate(votes)
}

func readVotes() []gocondorcet.Vote {
	// We use a string in this example - but this could be a file, STDIN, etc.
	input := strings.NewReader(`
	Tom,Sally=Dan
	Bob,tom,DAN
	Sally,Tom,Sally
	`)

	// We then define a vote reader, using a pre-defined Candidate ID parse
	// function. You can use your own to match user input (above) to distinct
	// string IDs.
	voteReader := gocondorcet.NewVoteReader(input, gocondorcet.BasicParseFn)

	// ReadAll will evaluate the reader up to EOF, and return valid and invalid
	// parsed votes. In this case, we ignore the invalid votes,
	// but you may wish to log them, etc.
	valid, _ := voteReader.ReadAll()
	return valid
}
