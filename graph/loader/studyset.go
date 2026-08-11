package loader

import (
	"context"
	"net/http"
	"quizfreely/api/auth"
	"quizfreely/api/graph/model"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vikstrous/dataloadgen"
)

func (dr *dataReader) getTermsCountByStudysetIDs(ctx context.Context, studysetIDs []string) ([]*int32, []error) {
	authedUser := auth.AuthedUserContext(ctx)
	var authedUserID *string
	if authedUser != nil {
		authedUserID = authedUser.ID
	}

	type countResult struct {
		StudysetID string `db:"studyset_id"`
		Count      int32  `db:"term_count"`
	}

	var results []countResult

	err := pgxscan.Select(
		ctx,
		dr.db,
		&results,
		`SELECT t.studyset_id, COUNT(t.*) AS term_count
         FROM terms t
         JOIN studysets s ON t.studyset_id = s.id
         WHERE t.studyset_id = ANY($1::uuid[]) AND ((s.private = false AND s.draft = false) OR s.user_id = $2)
         GROUP BY t.studyset_id`,
		studysetIDs,
		authedUserID,
	)
	if err != nil {
		return nil, []error{err}
	}

	// Map studysetID -> count for quick lookup
	countsMap := make(map[string]int32, len(results))
	for _, r := range results {
		countsMap[r.StudysetID] = r.Count
	}

	// Assemble slice in the same order as studysetIDs
	orderedCounts := make([]*int32, len(studysetIDs))
	for i, id := range studysetIDs {
		if c, ok := countsMap[id]; ok {
			orderedCounts[i] = &c
		} else {
			zero := int32(0)
			orderedCounts[i] = &zero
		}
	}

	return orderedCounts, nil
}

// GetTermsCountByStudysetID returns a single studyset's terms count efficiently
func GetTermsCountByStudysetID(ctx context.Context, studysetID string) (*int32, error) {
	loaders := For(ctx)
	return loaders.TermsCountByStudysetIDLoader.Load(ctx, studysetID)
}
// GetTermsCountByStudysetIDs returns many studysets' terms counts efficiently
func GetTermsCountByStudysetIDs(ctx context.Context, studysetIDs []string) ([]*int32, error) {
	loaders := For(ctx)
	return loaders.TermsCountByStudysetIDLoader.LoadAll(ctx, studysetIDs)
}
