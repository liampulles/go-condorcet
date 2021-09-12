<div align="center"><img src="pnyx.jpg" alt="Panorama view of Pnyx, Athens."></div>
<div align="center"><small><sup>Nikthestoned, C 2012, <i>An image of the Pnyx, in central Athens</i>, Pnyx, Athens</sup></small></div>
<h1 align="center">
  <b>go-condorcet</b>
</h1>

<h4 align="center">A Go library and CLI tool for conducting Schulze Condorcet elections.</h4>

<p align="center">
  <a href="#status">Status</a> •
  <a href="#usage">Usage</a> •
  <a href="#contributing">Contributing</a> •
  <a href="#license">License</a>
</p>

<p align="center">
  <a href="https://github.com/liampulles/go-condorcet/releases">
    <img src="https://img.shields.io/github/release/go-condorcet/go-condorcet.svg" alt="[GitHub release]"/>
  </a>
  <a href="https://app.travis-ci.com/liampulles/go-condorcet">
    <img src="https://app.travis-ci.com/liampulles/go-condorcet.svg?branch=main" alt="[Build Status]"/>
  </a>
    <img src="https://img.shields.io/github/go-mod/go-version/liampulles/go-condorcet" alt="GitHub go.mod Go version"/>
  <a href="https://goreportcard.com/report/github.com/liampulles/go-condorcet">
    <img src="https://goreportcard.com/badge/github.com/liampulles/go-condorcet" alt="[Go Report Card]"/>
  </a>
  <a href="https://codecov.io/gh/liampulles/go-condorcet">
    <img src="https://codecov.io/gh/liampulles/go-condorcet/branch/main/graph/badge.svg?token=HNBUXFX8AZ" alt="[Code Coverage Report]"/>
  </a>
  <a href="https://github.com/liampulles/go-condorcet/blob/main/LICENSE.md">
    <img src="https://img.shields.io/github/license/liampulles/go-condorcet.svg" alt="[License]"/>
  </a>
</p>

## Status

go-condorcet is complete, and the API is stable.

## Usage

### Library

```go
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
```

## CLI

Given a `votes.txt` file:

```
C,A,B,E,D
A,C=B,E,D
B,E,D,A,C
C,B,A=D,E
C,A,B
B,E=D,A,C
A,C,B,E
```

You can then calculate the Schulze condorcet result as follows:

```shell
$ cat votes.txt | condorcli
A
C
B
E
D
```

## Contributing

Please submit an issue with your proposal.

## License

See [LICENSE](LICENSE)