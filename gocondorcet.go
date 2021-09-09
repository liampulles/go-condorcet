package gocondorcet

func Evaluate(votes []Vote) Results {
	return nil
}

// --- Types ---

type Vote map[CandidateID]Preference

type Preference uint

type CandidateID string

type Results []CandidateID
