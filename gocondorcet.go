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
	"math"
	"strings"
)

// Evaluate will perform a Schulze Condorcet election.
func Evaluate(votes []Vote) Results {
	ids := findCandidateIDSet(votes)
	pairPrefs := findPairwisePreferences(votes, ids)
	strongestPaths := findStrongestPathStrengths(pairPrefs, ids)
	fmt.Println(strongestPaths)
	rank := findRank(strongestPaths, ids)

	return Results(rank)
}

func findRank(strongestPaths idPairValueMap, ids []CandidateID) []CandidateID {
	winnerMap := make(map[CandidateID]int)

	for _, id := range ids {
		winnerMap[id] = 0
	}

	iterateOverIDPairs(ids, func(idI, idJ CandidateID) {
		if strongestPaths.Get(idI, idJ) > strongestPaths.Get(idJ, idI) {
			winnerMap[idI]++
		}
	})

	var result []CandidateID

	for len(winnerMap) > 0 {
		bestRank := -1

		var bestID CandidateID

		for id, rank := range winnerMap {
			if rank > bestRank {
				bestRank = rank
				bestID = id
			}
		}

		result = append(result, bestID)

		delete(winnerMap, bestID)
	}

	return result
}

func findStrongestPathStrengths(pairPrefs idPairValueMap, ids []CandidateID) idPairValueMap {
	result := newIDPairValueMap()

	iterateOverIDPairs(ids, func(idI, idJ CandidateID) {
		dIJ := pairPrefs.Get(idI, idJ)
		if dIJ > pairPrefs.Get(idJ, idI) {
			result.Set(idI, idJ, dIJ)
		} else {
			result.Set(idI, idJ, 0)
		}
	})

	iterateOverIDTriples(ids, func(idI, idJ, idK CandidateID) {
		result.Set(idJ, idK, max(
			result.Get(idJ, idK),
			min(
				result.Get(idJ, idI),
				result.Get(idI, idK),
			),
		))
	})

	return result
}

func findPairwisePreferences(votes []Vote, ids []CandidateID) idPairValueMap {
	result := newIDPairValueMap()

	for _, vote := range votes {
		addPairwisePreferencesForVote(vote, ids, result)
	}

	return result
}

func addPairwisePreferencesForVote(
	vote Vote,
	ids []CandidateID,
	result idPairValueMap,
) {
	iterateOverIDPairs(ids, func(idI, idJ CandidateID) {
		prefI := findVotePref(vote, idI)
		prefJ := findVotePref(vote, idJ)
		// If I is preferred to J, that is its value is strictly lower.
		if prefI < prefJ {
			result.Add(idI, idJ, 1)
		}
	})
}

func findVotePref(vote Vote, id CandidateID) uint {
	pref, ok := vote[id]
	if ok {
		return uint(pref)
	}

	return math.MaxUint
}

func findCandidateIDSet(votes []Vote) []CandidateID {
	set := make(map[CandidateID]bool)

	for _, vote := range votes {
		for id := range vote {
			set[id] = true
		}
	}

	result := make([]CandidateID, len(set))
	count := -1

	for id := range set {
		count++

		result[count] = id
	}

	return result
}

func iterateOverIDPairs(ids []CandidateID, fn func(idI, idJ CandidateID)) {
	for i := 0; i < len(ids); i++ {
		idI := ids[i]

		for j := 0; j < len(ids); j++ {
			if i == j {
				continue
			}

			idJ := ids[j]
			fn(idI, idJ)
		}
	}
}

func iterateOverIDTriples(ids []CandidateID, fn func(idI, idJ, idK CandidateID)) {
	for i := 0; i < len(ids); i++ {
		idI := ids[i]

		for j := 0; j < len(ids); j++ {
			if i == j {
				continue
			}

			idJ := ids[j]

			for k := 0; k < len(ids); k++ {
				if i == k || j == k {
					continue
				}

				idK := ids[k]
				fn(idI, idJ, idK)
			}
		}
	}
}

func min(a, b uint) uint {
	if a < b {
		return a
	}

	return b
}

func max(a, b uint) uint {
	if a > b {
		return a
	}

	return b
}

type idPairValueMap map[candidatePair]uint

func newIDPairValueMap() idPairValueMap {
	return idPairValueMap(make(map[candidatePair]uint))
}

func (m idPairValueMap) Set(i, j CandidateID, value uint) {
	m[candidatePair{i, j}] = value
}

func (m idPairValueMap) Get(i, j CandidateID) uint {
	v, ok := m[candidatePair{i, j}]
	if ok {
		return v
	}

	return 0
}

func (m idPairValueMap) Add(i, j CandidateID, value uint) {
	current := m.Get(i, j)
	m.Set(i, j, current+value)
}

type candidatePair struct {
	A CandidateID
	B CandidateID
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
