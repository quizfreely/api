package resolver

//go:generate go run github.com/99designs/gqlgen generate

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

import (
	"fmt"
	"time"
	"regexp"
	"quizfreely/api/graph/model"

	"github.com/jackc/pgx/v5/pgxpool"
	fsrs "github.com/open-spaced-repetition/go-fsrs/v4"
)

// ✅ This regex is NOT vulnerable to ReDoS because there's no repetition operator. It does not contain any quantifiers, nested groups, or alternation. It's a single character class.
var validTitleRegex = regexp.MustCompile(`[\p{L}\p{M}\p{N}]`)

const MaxBatchMutationSize = 9000
const MaxFolderNameLen = 1000

type Resolver struct {
	DB                 *pgxpool.Pool
	UsercontentBaseURL *string
}

func ptrToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

type fsrsCardUpdate struct {
	termID string
	card   fsrs.Card
}
type fsrsReviewLogWrite struct {
	termID string
	log    fsrs.ReviewLog
}
func fsrsStateToDB(s fsrs.State) string {
	switch s {
	case fsrs.New:
		return "NEW"
	case fsrs.Learning:
		return "LEARNING"
	case fsrs.Review:
		return "REVIEW"
	case fsrs.Relearning:
		return "RELEARNING"
	}
	return "NEW"
}
func fsrsStateFromModel(s model.FSRSState) fsrs.State {
	switch s {
	case model.FSRSStateNew:
		return fsrs.New
	case model.FSRSStateLearning:
		return fsrs.Learning
	case model.FSRSStateReview:
		return fsrs.Review
	case model.FSRSStateRelearning:
		return fsrs.Relearning
	}
	return fsrs.New
}
func fsrsRatingToDB(r fsrs.Rating) string {
	switch r {
	case fsrs.Again:
		return "AGAIN"
	case fsrs.Hard:
		return "HARD"
	case fsrs.Good:
		return "GOOD"
	case fsrs.Easy:
		return "EASY"
	}
	return "MANUAL"
}
func parseFSRSTime(s string) (time.Time, error) {
	t, err := time.Parse("2006-01-02T15:04:05-07:00", s)
	if err != nil {
		t, err = time.Parse("2006-01-02T15:04:05Z07:00", s)
		if err != nil {
			return time.Time{}, fmt.Errorf("failed to parse timestamp %q: %w", s, err)
		}
	}
	return t, nil
}
func fsrsCardFromModel(c *model.FSRSCard) (fsrs.Card, error) {
	due, err := parseFSRSTime(c.Due)
	if err != nil {
		return fsrs.Card{}, err
	}
	var lastReview time.Time
	if c.LastReview != nil {
		lastReview, err = parseFSRSTime(*c.LastReview)
		if err != nil {
			return fsrs.Card{}, err
		}
	}
	return fsrs.Card{
		Due:            due,
		Stability:      c.Stability,
		Difficulty:     c.Difficulty,
		ScheduledDays:  uint64(c.ScheduledDays),
		Reps:           uint64(c.Reps),
		Lapses:         uint64(c.Lapses),
		State:          fsrsStateFromModel(c.State),
		LastReview:     lastReview,
		RemainingSteps: int(c.LearningSteps),
	}, nil
}
