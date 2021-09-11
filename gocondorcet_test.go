package gocondorcet_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"testing/iotest"

	gocondorcet "github.com/liampulles/go-condorcet"
	"github.com/stretchr/testify/assert"
)

func TestEvaluate_WhenThereAreNoVotes_ShouldReturnNoResults(t *testing.T) {
	// Setup fixture
	var fixtures = [][]gocondorcet.Vote{
		nil,
		{},
		{{}},
	}

	for _, fixture := range fixtures {
		t.Run(fmt.Sprintf("\"%+v\"", fixture), func(t *testing.T) {
			// Exercise SUT
			actual := gocondorcet.Evaluate(fixture)

			// Verify result
			assert.Nil(t, actual)
		})
	}
}

func TestVoteReader_GivenInvalidReader_ShouldReturnInvalidVote(t *testing.T) {
	// Setup fixture
	r := iotest.ErrReader(errors.New("fail")) // nolint:goerr113
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
		{
			"sam=dave\nsam,sam\n\n\n\n\ntammy,,sue,bob=alice=dave\nbob,dan=jim",
			[]gocondorcet.InvalidVote{
				{1, "cyclic vote detected: cannot reference a candidate twice in a vote"},
				{7, "could not parse candidate ID: no jims allowed"},
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
		return gocondorcet.CandidateID(""), errors.New("no jims allowed") // nolint:goerr113
	}

	return gocondorcet.CandidateID(in), nil
}
