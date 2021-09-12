package gocondorcet_test

import (
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"testing"
	"testing/iotest"

	gocondorcet "github.com/liampulles/go-condorcet"
	"github.com/stretchr/testify/assert"
)

func TestEvaluate_GivenValidVotes_ShouldReturnResults(t *testing.T) {
	// Setup fixture
	var tests = []struct {
		fixture  []gocondorcet.Vote
		expected gocondorcet.Results
	}{
		// Empty cases
		{nil, nil},
		{[]gocondorcet.Vote{}, nil},
		{[]gocondorcet.Vote{{}}, nil},
		// Examples from https://electowiki.org/wiki/Schulze_method
		{
			testVotes(map[string]int{
				"ACBED": 5,
				"ADECB": 5,
				"BEDAC": 8,
				"CABED": 3,
				"CAEBD": 7,
				"CBADE": 2,
				"DCEBA": 7,
				"EBADC": 8,
			}),
			testResults("EACBD"),
		},
		{
			testVotes(map[string]int{
				"ACBD": 5,
				"ACDB": 2,
				"ADCB": 3,
				"BACD": 4,
				"CBDA": 3,
				"CDBA": 3,
				"DACB": 1,
				"DBAC": 5,
				"DCBA": 4,
			}),
			testResults("DACB"),
		},
		{
			testVotes(map[string]int{
				"ABDEC": 3,
				"ADEBC": 5,
				"ADECB": 1,
				"BADEC": 2,
				"BDECA": 2,
				"CABDE": 4,
				"CBADE": 6,
				"DBECA": 2,
				"DECAB": 5,
			}),
			testResults("BADEC"),
		},
		// -> Same as above, but we've omitted the last candidate (should be
		//    picked up and effectively filled in)
		{
			testVotes(map[string]int{
				"ABDE": 3,
				"ADEB": 5,
				"ADEC": 1,
				"BADE": 2,
				"BDEC": 2,
				"CABD": 4,
				"CBAD": 6,
				"DBEC": 2,
				"DECA": 5,
			}),
			testResults("BADEC"),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("\"%+v\"", test.fixture), func(t *testing.T) {
			// Exercise SUT
			actual := gocondorcet.Evaluate(test.fixture)

			// Verify result
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestEvaluate_GivenVariantResultCase_ShouldReturnOneValidResult(t *testing.T) {
	// Setup fixture
	fixture := testVotes(map[string]int{
		"ABCD": 3,
		"DABC": 2,
		"DBCA": 2,
		"CBDA": 2,
	})

	// Setup expectations
	potentialExpected := []gocondorcet.Results{
		testResults("BCDA"),
		testResults("BDAC"),
		testResults("BDCA"),
		testResults("DABC"),
		testResults("DBAC"),
		testResults("DBCA"),
	}

	// Exercise SUT
	actual := gocondorcet.Evaluate(fixture)

	// Verify result
	oneMatch := false
	for _, expected := range potentialExpected {
		if reflect.DeepEqual(actual, expected) {
			oneMatch = true
		}
	}
	if !oneMatch {
		t.Error("actual does not match any of expected (actual...)", actual)
	}
}

func TestVoteReader_GivenInvalidReader_ShouldReturnInvalidVote(t *testing.T) {
	// Setup fixture
	r := iotest.ErrReader(errors.New("fail"))
	vr := gocondorcet.NewVoteReader(r, identParse)

	// Exercise SUT
	actual, err := vr.Read()

	// Verify results
	assert.Nil(t, actual)
	assert.Error(t, err)
}

func TestVoteReader_GivenValidReader_ShouldParse(t *testing.T) {
	// Setup fixture
	var tests = []struct {
		fixture         string
		expectedInvalid []gocondorcet.InvalidVote
		expected        []gocondorcet.Vote
	}{
		// Empty cases
		{
			"",
			nil,
			nil,
		},
		{
			"\n\n\n",
			nil,
			nil,
		},
		// Simple cases
		{
			"bob",
			nil,
			[]gocondorcet.Vote{{"bob": 0}},
		},
		{
			"bob,sam\nalice",
			nil,
			[]gocondorcet.Vote{
				{"bob": 0, "sam": 1},
				{"alice": 0},
			},
		},
		{
			"bob,sam=alice,dave",
			nil,
			[]gocondorcet.Vote{
				{"bob": 0, "sam": 1, "alice": 1, "dave": 2},
			},
		},
		// Complex case
		{
			"sam=dave\nsam,sam\n\n\n\n\ntammy,,sue,bob=alice=dave\nbob,dan=jim",
			[]gocondorcet.InvalidVote{
				{1, "could not read record: could not add equal elements: cyclic vote detected: cannot reference a candidate twice in a vote"},
				{7, "could not read record: could not add equal elements: could not parse candidate ID: no jims allowed"},
			},
			[]gocondorcet.Vote{
				{"sam": 0, "dave": 0},
				{"tammy": 0, "sue": 1, "bob": 2, "alice": 2, "dave": 2},
			},
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("\"%v\"", test.fixture), func(t *testing.T) {
			// Setup fixture
			r := gocondorcet.NewVoteReader(strings.NewReader(test.fixture), failJimParse)

			// Exercise SUT
			actual, invalid := r.ReadAll()

			// Verify result
			assert.Equal(t, test.expectedInvalid, invalid)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestBasicDiscovering_ShouldAccumulateTrimmedUpperIDs(t *testing.T) {
	// Setup fixture
	fn, accum := gocondorcet.BasicDiscovering()
	fixtures := []string{
		"bob",
		"alice",
		" Bob",
		"dave ",
		"ALICE ",
		"sam",
		" dAvE ",
	}
	expectedFn := []gocondorcet.CandidateID{
		"BOB",
		"ALICE",
		"BOB",
		"DAVE",
		"ALICE",
		"SAM",
		"DAVE",
	}
	expectedAccum := []gocondorcet.CandidateID{
		"BOB",
		"ALICE",
		"DAVE",
		"SAM",
	}

	// Exercise SUT
	for i, fixture := range fixtures {
		expected := expectedFn[i]

		t.Run(fmt.Sprintf("\"%s\"", fixture), func(t *testing.T) {
			// Exercise SUT
			actual, err := fn(fixture)

			// Verify result
			assert.NoError(t, err)
			assert.Equal(t, expected, actual)
		})
	}

	// Verify results
	assert.Equal(t, expectedAccum, *accum)
}

func identParse(in string) (gocondorcet.CandidateID, error) {
	return gocondorcet.CandidateID(in), nil
}

func failJimParse(in string) (gocondorcet.CandidateID, error) {
	if in == "jim" {
		return gocondorcet.CandidateID(""), errors.New("no jims allowed")
	}

	return gocondorcet.CandidateID(in), nil
}

func testVotes(voteCounts map[string]int) []gocondorcet.Vote {
	var result []gocondorcet.Vote

	for voteStr, count := range voteCounts {
		vote := gocondorcet.Vote(make(map[gocondorcet.CandidateID]gocondorcet.Preference))
		for i, r := range voteStr {
			vote[gocondorcet.CandidateID(r)] = gocondorcet.Preference(i)
		}
		for repeat := 0; repeat < count; repeat++ {
			result = append(result, vote)
		}
	}

	rand.Seed(7)
	rand.Shuffle(len(result),
		func(i, j int) { result[i], result[j] = result[j], result[i] })

	return result
}

func testResults(resultStr string) []gocondorcet.CandidateID {
	result := make([]gocondorcet.CandidateID, len(resultStr))

	for i, r := range resultStr {
		result[i] = gocondorcet.CandidateID(r)
	}

	return result
}
