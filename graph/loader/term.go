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

func (dr *dataReader) getTermsByIDs(ctx context.Context, ids []string) ([]*model.Term, []error) {
	authedUser := auth.AuthedUserContext(ctx)
	var authedUserID *string
	if authedUser != nil {
		authedUserID = authedUser.ID
	}

	type dbTerm struct {
		ID           *string `db:"id"`
		StudysetID   *string `db:"studyset_id"`
		Term         *string `db:"term"`
		Def          *string `db:"def"`
		TermImageURL *string `db:"term_image_url"`
		DefImageURL  *string `db:"def_image_url"`
		SortOrder    *int32  `db:"sort_order"`
		CreatedAt    *string `db:"created_at"`
		UpdatedAt    *string `db:"updated_at"`
	}
	var dbTerms []*dbTerm

	err := pgxscan.Select(
		ctx,
		dr.db,
		&dbTerms,
		`SELECT t.id, t.studyset_id, t.term, t.def, ($3||t.term_image_key) as term_image_url, ($3||t.def_image_key) as def_image_url, t.sort_order,
	to_char(t.created_at, 'YYYY-MM-DD"T"HH24:MI:SS.MSTZH:TZM') as created_at,
	to_char(t.updated_at, 'YYYY-MM-DD"T"HH24:MI:SS.MSTZH:TZM') as updated_at
FROM unnest($1::uuid[]) WITH ORDINALITY AS input(id, og_order)
LEFT JOIN (
	SELECT t.*
	FROM terms t
	JOIN studysets s ON t.studyset_id = s.id
	WHERE (s.private = false AND s.draft = false) OR s.user_id = $2
) t ON t.id = input.id
ORDER BY input.og_order`,
		ids,
		authedUserID,
		dr.usercontentBaseURL,
	)
	if err != nil {
		return nil, []error{err}
	}

	terms := make([]*model.Term, len(dbTerms))
	for i, dt := range dbTerms {
		if dt.ID == nil {
			terms[i] = nil
		} else {
			terms[i] = &model.Term{
				ID:           dt.ID,
				StudysetID:   dt.StudysetID,
				Term:         dt.Term,
				Def:          dt.Def,
				TermImageURL: dt.TermImageURL,
				DefImageURL:  dt.DefImageURL,
				SortOrder:    dt.SortOrder,
				CreatedAt:    dt.CreatedAt,
				UpdatedAt:    dt.UpdatedAt,
			}
		}
	}

	return terms, nil
}

func (dr *dataReader) getTermsByStudysetIDs(ctx context.Context, studysetIDs []string) ([][]*model.Term, []error) {
	authedUser := auth.AuthedUserContext(ctx)
	var authedUserID *string
	if authedUser != nil {
		authedUserID = authedUser.ID
	}

	var terms []*model.Term

	err := pgxscan.Select(
		ctx,
		dr.db,
		&terms,
		`SELECT t.id, t.studyset_id, t.term, t.def, ($3||t.term_image_key) as term_image_url, ($3||t.def_image_key) as def_image_url, t.sort_order,
	to_char(t.created_at, 'YYYY-MM-DD"T"HH24:MI:SS.MSTZH:TZM') as created_at,
	to_char(t.updated_at, 'YYYY-MM-DD"T"HH24:MI:SS.MSTZH:TZM') as updated_at
FROM terms t
JOIN studysets s ON t.studyset_id = s.id
WHERE t.studyset_id = ANY($1::uuid[]) AND ((s.private = false AND s.draft = false) OR s.user_id = $2)
ORDER BY t.studyset_id, t.sort_order`,
		studysetIDs,
		authedUserID,
		dr.usercontentBaseURL,
	)
	if err != nil {
		return nil, []error{err}
	}

	// Group terms by studyset_id
	grouped := make(map[string][]*model.Term)
	for _, t := range terms {
		if t.StudysetID != nil {
			grouped[*t.StudysetID] = append(grouped[*t.StudysetID], t)
		}
	}

	// Reassemble in the same order as studysetIDs
	orderedTerms := make([][]*model.Term, len(studysetIDs))
	for i, id := range studysetIDs {
		orderedTerms[i] = grouped[id]
	}

	return orderedTerms, nil
}

func (dr *dataReader) getTermsProgress(ctx context.Context, termIDs []string) ([]*model.TermProgress, []error) {
	authedUser := auth.AuthedUserContext(ctx)
	if authedUser == nil || authedUser.ID == nil {
		return nil, nil
	}

	rows, err := dr.db.Query(
		ctx,
		`SELECT tp.id,
	to_char(tp.term_first_reviewed_at, 'YYYY-MM-DD"T"HH24:MI:SS.MSTZH:TZM') as term_first_reviewed_at,
	to_char(tp.term_last_reviewed_at, 'YYYY-MM-DD"T"HH24:MI:SS.MSTZH:TZM') as term_last_reviewed_at,
	tp.term_review_count,
	to_char(tp.def_first_reviewed_at, 'YYYY-MM-DD"T"HH24:MI:SS.MSTZH:TZM') as def_first_reviewed_at,
	to_char(tp.def_last_reviewed_at, 'YYYY-MM-DD"T"HH24:MI:SS.MSTZH:TZM') as def_last_reviewed_at,
	tp.def_review_count,
	tp.term_correct_count, tp.term_incorrect_count,
	tp.def_correct_count, tp.def_incorrect_count
FROM unnest($1::uuid[]) WITH ORDINALITY AS input(term_id, og_order)
LEFT JOIN term_progress tp
	ON tp.term_id = input.term_id
	AND tp.user_id = $2
ORDER BY input.og_order`,
		termIDs,
		authedUser.ID,
	)
	if err != nil {
		return nil, []error{err}
	}
	defer rows.Close()

	// Define a struct with pointers to handle potential NULLs from LEFT JOIN
	type dbTermProgress struct {
		ID                   *string `db:"id"`
		TermFirstReviewedAt  *string `db:"term_first_reviewed_at"`
		TermLastReviewedAt   *string `db:"term_last_reviewed_at"`
		TermReviewCount      *int32  `db:"term_review_count"`
		DefFirstReviewedAt   *string `db:"def_first_reviewed_at"`
		DefLastReviewedAt    *string `db:"def_last_reviewed_at"`
		DefReviewCount       *int32  `db:"def_review_count"`
		TermCorrectCount     *int32  `db:"term_correct_count"`
		TermIncorrectCount   *int32  `db:"term_incorrect_count"`
		DefCorrectCount      *int32  `db:"def_correct_count"`
		DefIncorrectCount    *int32  `db:"def_incorrect_count"`
	}

	var termsProgress []*model.TermProgress
	for rows.Next() {
		var tp dbTermProgress
		err := pgxscan.ScanRow(&tp, rows)
		if err != nil {
			return nil, []error{err}
		}

		if tp.ID == nil {
			termsProgress = append(termsProgress, nil)
		} else {
			modelTp := &model.TermProgress{
				ID:                   tp.ID,
				TermFirstReviewedAt:  tp.TermFirstReviewedAt,
				TermLastReviewedAt:   tp.TermLastReviewedAt,
				TermReviewCount:      tp.TermReviewCount,
				DefFirstReviewedAt:   tp.DefFirstReviewedAt,
				DefLastReviewedAt:    tp.DefLastReviewedAt,
				DefReviewCount:       tp.DefReviewCount,
			}
			if tp.TermCorrectCount != nil {
				modelTp.TermCorrectCount = *tp.TermCorrectCount
			}
			if tp.TermIncorrectCount != nil {
				modelTp.TermIncorrectCount = *tp.TermIncorrectCount
			}
			if tp.DefCorrectCount != nil {
				modelTp.DefCorrectCount = *tp.DefCorrectCount
			}
			if tp.DefIncorrectCount != nil {
				modelTp.DefIncorrectCount = *tp.DefIncorrectCount
			}
			termsProgress = append(termsProgress, modelTp)
		}
	}

	return termsProgress, nil
}

func GetTermByID(ctx context.Context, id string) (*model.Term, error) {
	loaders := For(ctx)
	return loaders.TermByIDLoader.Load(ctx, id)
}
func GetTermsByIDs(ctx context.Context, ids []string) ([]*model.Term, error) {
	loaders := For(ctx)
	return loaders.TermByIDLoader.LoadAll(ctx, ids)
}

// GetTermsByStudysetID returns a single studyset's terms efficiently
func GetTermsByStudysetID(ctx context.Context, studysetID string) ([]*model.Term, error) {
	loaders := For(ctx)
	return loaders.TermByStudysetIDLoader.Load(ctx, studysetID)
}
// GetTermsByStudysetIDs returns many studysets' terms efficiently
func GetTermsByStudysetIDs(ctx context.Context, studysetIDs []string) ([][]*model.Term, error) {
	loaders := For(ctx)
	return loaders.TermByStudysetIDLoader.LoadAll(ctx, studysetIDs)
}

// GetTermProgress returns a single term's progress record by term id efficiently
func GetTermProgress(ctx context.Context, termID string) (*model.TermProgress, error) {
	loaders := For(ctx)
	return loaders.TermProgressLoader.Load(ctx, termID)
}
// GetTermsProgress returns many terms' progress records by term ids efficiently
func GetTermsProgress(ctx context.Context, termIDs []string) ([]*model.TermProgress, error) {
	loaders := For(ctx)
	return loaders.TermProgressLoader.LoadAll(ctx, termIDs)
}
