package main

import (
	"fmt"
	"io"
	"os"

	gocondorcet "github.com/liampulles/go-condorcet"
)

func main() {
	run(os.Stdin, os.Stdout, os.Stderr)
}

func run(input io.Reader, output, errOutput io.Writer) {
	votes, invalid := readVotes(input)
	writeInvalid(invalid, errOutput)

	results := gocondorcet.Evaluate(votes)
	writeResults(results, output)
}

func readVotes(input io.Reader) ([]gocondorcet.Vote, []gocondorcet.InvalidVote) {
	vr := gocondorcet.NewVoteReader(input, gocondorcet.BasicParseFn)
	return vr.ReadAll()
}

func writeInvalid(invalid []gocondorcet.InvalidVote, errOutput io.Writer) {
	for _, in := range invalid {
		fmt.Fprintf(errOutput, "ERROR [Line %d]: %s\n", in.Line, in.Issue)
	}
}

func writeResults(results []gocondorcet.CandidateID, output io.Writer) {
	for _, result := range results {
		fmt.Fprintf(output, "%s\n", result)
	}
}
