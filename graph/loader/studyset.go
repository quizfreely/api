package loader

import (
	"context"
	"fmt"
	"quizfreely/api/auth"
	"quizfreely/api/graph/model"

	"github.com/georgysavva/scany/v2/pgxscan"
)

func (dr *dataReader) getStudysetsByIDs(ctx context.Context, ids []string) ([]*model.Studyset, []error) {
	if len(ids) == 0 {
		return []*model.Studyset{}, nil
	}

	authedUser := auth.AuthedUserContext(ctx)

	type row struct {
		model.Studyset
		Ordinality int `db:"ordinality"`
	}

	selectCols := `
		s.id, s.user_id, s.title, s.private, s.draft, s.subject_id, s.seo_indexing_approved,
		to_char(s.created_at, 'YYYY-MM-DD"T"HH24:MI:SS.MSTZH:TZM') as created_at,
		to_char(s.updated_at, 'YYYY-MM-DD"T"HH24:MI:SS.MSTZH:TZM') as updated_at,
		input.ordinality
	`

	var rows []row
	var err error

	if authedUser != nil {
		err = pgxscan.Select(ctx, dr.db, &rows, `
			SELECT `+selectCols+`
			FROM unnest($1::uuid[]) WITH ORDINALITY AS input(id, ordinality)
			LEFT JOIN studysets s ON s.id = input.id
				AND ((s.private = false AND s.draft = false) OR s.user_id = $2)
			ORDER BY input.ordinality
		`, ids, authedUser.ID)
	} else {
		err = pgxscan.Select(ctx, dr.db, &rows, `
			SELECT `+selectCols+`
			FROM unnest($1::uuid[]) WITH ORDINALITY AS input(id, ordinality)
			LEFT JOIN studysets s ON s.id = input.id
				AND s.draft = false
				AND s.private = false
			ORDER BY input.ordinality
		`, ids)
	}
	if err != nil {
		errs := make([]error, len(ids))
		batchErr := fmt.Errorf("failed to fetch studysets: %w", err)
		for i := range errs {
			errs[i] = batchErr
		}
		return nil, errs
	}

	result := make([]*model.Studyset, len(ids))
	for _, r := range rows {
		idx := r.Ordinality - 1
		if r.Studyset.ID != nil {
			s := r.Studyset
			result[idx] = &s
		}
	}

	return result, nil
}

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
		/* NOTE: $2 is NULL if authedUserID is nil
		   `s.user_id = NULL` does NOT select rows where user_id is NULL,
		   only `IS NULL` would select rows with NULL,
		   SQL has TRUE, FALSE, and UNKNOWN,
		   and this will correctly handle NULL for $2
		   because `TRUE OR UNKNOWN` is true, `FALSE OR UNKNOWN` is unknown,
		   and `WHERE UNKNOWN` does not select anything,
		   so a nil ID will only return public studysets (correctly) */
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

func GetStudysetByID(ctx context.Context, id string) (*model.Studyset, error) {
	loaders := For(ctx)
	return loaders.StudysetLoader.Load(ctx, id)
}

func GetStudysetsByIDs(ctx context.Context, ids []string) ([]*model.Studyset, error) {
	loaders := For(ctx)
	return loaders.StudysetLoader.LoadAll(ctx, ids)
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
