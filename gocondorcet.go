/*
Package gocondorcet provides methods for running Condorcet elections.
*/
package gocondorcet

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strings"
)

// Evaluate will perform a Schulze Condorcet election.
func Evaluate(votes []Vote) Results {
	return nil
}

// --- Utility methods ---

// VoteReader wraps an io.Reader to read Votes.
type VoteReader struct {
	scanner *bufio.Scanner
	parseFn CandidateIDParseFn
}

// NewVoteReader is a constructor.
func NewVoteReader(r io.Reader, parseFn CandidateIDParseFn) *VoteReader {
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanLines)

	return &VoteReader{
		scanner: scanner,
		parseFn: parseFn,
	}
}

// BasicDiscovering provides a basic implementation of CandidateIDParseFn,
// which also records discovered CandidateIDs to a list (when used).
// The parser will trim whitespace and uppercase input.
func BasicDiscovering() (CandidateIDParseFn, *[]CandidateID) {
	var result []CandidateID

	seen := make(map[CandidateID]int)
	resultPtr := &result

	return func(in string) (CandidateID, error) {
		trimmed := strings.TrimSpace(in)
		upper := strings.ToUpper(trimmed)
		id := CandidateID(upper)
		_, ok := seen[id]

		if !ok {
			appended := append(*resultPtr, id)
			*resultPtr = appended
			seen[id] = len(*resultPtr)
		}

		return id, nil
	}, resultPtr
}

// ReadAll reads all votes from the VoteReader (until the reader reaches EOF).
// Lines which can't be parsed are returned as InvalidVotes.
func (vr *VoteReader) ReadAll() ([]Vote, []InvalidVote) {
	var valid []Vote

	var invalid []InvalidVote

	finished := false

	for count := 0; !finished; count++ {
		vote, err := vr.Read()

		if err != nil {
			if errors.Is(err, io.EOF) {
				finished = true
				continue
			}

			invalid = append(invalid, InvalidVote{
				Line:  count,
				Issue: err.Error(),
			})
		}

		if vote != nil {
			valid = append(valid, *vote)
		}
	}

	return valid, invalid
}

func (vr *VoteReader) Read() (*Vote, error) {
	if !vr.scanner.Scan() {
		if err := vr.scanner.Err(); err != nil {
			return nil, fmt.Errorf("scanner error: %w", err)
		}

		return nil, io.EOF
	}

	record := splitChar(vr.scanner.Text(), ',')
	result := make(map[CandidateID]Preference)
	pref := -1

	for _, elem := range record {
		equals := splitChar(elem, '=')

		if len(equals) == 0 {
			continue
		}

		pref++

		for _, equalElem := range equals {
			cID, err := vr.parseFn(equalElem)

			if err != nil {
				return nil, fmt.Errorf("could not parse candidate ID: %w", err)
			}

			if _, ok := result[cID]; ok {
				return nil, ErrCyclicVote
			}

			result[cID] = Preference(pref)
		}
	}

	if len(result) == 0 {
		return nil, nil
	}

	vote := Vote(result)

	return &vote, nil
}

func splitChar(elem string, split rune) []string {
	r := csv.NewReader(strings.NewReader(elem))
	r.Comma = split
	equals, err := r.Read()

	if errors.Is(err, io.EOF) {
		return nil
	}

	return equals
}

// Defined errors.
var (
	ErrCyclicVote = errors.New("cyclic vote detected: cannot reference a candidate twice in a vote")
)

// --- Types ---

// Vote allocates preferences to candidates (0 being
// highest preference). A vote may give candidates equal preference, and
// may omit candidates (effectively giving them equal lowest preference).
type Vote map[CandidateID]Preference

// InvalidVote describes an issue with input which could not be parsed into
// a vote.
type InvalidVote struct {
	Line  int
	Issue string
}

// Preference indicates to what degree something is preferred over other
// options. 0 is the highest preference, indicating most preferred.
type Preference uint

// CandidateID is a unique identifier for a candidate.
type CandidateID string

// Results consists of the ranked set of candidates, resulting from an election.
type Results []CandidateID

// CandidateIDParseFn should convert a user inputted candidate name into a
// CandidateID. you may wish to trim whitespace, convert to uppercase, etc.
// If the input corresponds to no valid candidate, you should return an error
// with the main reason.
type CandidateIDParseFn func(string) (CandidateID, error)
