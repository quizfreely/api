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

func (dr *dataReader) getFSRSCardsByTermIDs(ctx context.Context, termIDs []string) ([]*model.FSRSCard, []error) {
	authedUser := auth.AuthedUserContext(ctx)
	if authedUser == nil || authedUser.ID == nil {
		return nil, nil
	}

	rows, err := dr.db.Query(
		ctx,
		`SELECT fc.term_id,
    fc.difficulty,
	to_char(fc.due, 'YYYY-MM-DD"T"HH24:MI:SS.MSTZH:TZM') as due,
    fc.lapses,
	to_char(fc.last_review, 'YYYY-MM-DD"T"HH24:MI:SS.MSTZH:TZM') as last_review,
    fc.learning_steps,
    fc.reps,
    fc.scheduled_days,
    fc.stability,
    fc.state
FROM unnest($1::uuid[]) WITH ORDINALITY AS input(term_id, og_order)
LEFT JOIN fsrs_cards fc
	ON fc.term_id = input.term_id
	AND fc.user_id = $2
ORDER BY input.og_order`,
		termIDs,
		authedUser.ID,
	)
	if err != nil {
		return nil, []error{err}
	}
	defer rows.Close()

	type dbFSRSCard struct {
		TermID        *string          `db:"term_id"`
		Difficulty    *float64         `db:"difficulty"`
		Due           *string          `db:"due"`
		Lapses        *int32           `db:"lapses"`
		LastReview    *string          `db:"last_review"`
		LearningSteps *int32           `db:"learning_steps"`
		Reps          *int32           `db:"reps"`
		ScheduledDays *int32           `db:"scheduled_days"`
		Stability     *float64         `db:"stability"`
		State         *model.FSRSState `db:"state"`
	}

	var fsrsCards []*model.FSRSCard
	for rows.Next() {
		var fc dbFSRSCard
		err := pgxscan.ScanRow(&fc, rows)
		if err != nil {
			return nil, []error{err}
		}

		if fc.TermID == nil {
			fsrsCards = append(fsrsCards, nil)
		} else {
			modelFc := &model.FSRSCard{
				Difficulty:    *fc.Difficulty,
				Due:           *fc.Due,
				Lapses:        *fc.Lapses,
				LastReview:    fc.LastReview, /* actually nullable, so it's a pointer */
				LearningSteps: *fc.LearningSteps,
				Reps:          *fc.Reps,
				ScheduledDays: *fc.ScheduledDays,
				Stability:     *fc.Stability,
				State:         *fc.State,
			}
			fsrsCards = append(fsrsCards, modelFc)
		}
	}

	return fsrsCards, nil
}

func (dr *dataReader) getFSRSReviewLogsByTermIDs(ctx context.Context, termIDs []string) ([][]*model.FSRSReviewLog, []error) {
	authedUser := auth.AuthedUserContext(ctx)
	if authedUser == nil || authedUser.ID == nil {
		return nil, nil
	}

	type dbFSRSReviewLog struct {
		ID            *string           `db:"id"`
		TermID        *string           `db:"term_id"`
		Difficulty    *float64          `db:"difficulty"`
		Due           *string           `db:"due"`
		LearningSteps *int32            `db:"learning_steps"`
		Rating        *model.FSRSRating `db:"rating"`
		Review        *string           `db:"review"`
		ScheduledDays *int32            `db:"scheduled_days"`
		Stability     *float64          `db:"stability"`
		State         *model.FSRSState  `db:"state"`
	}
	var dbFSRSReviewLogs []*dbFSRSReviewLog

	err := pgxscan.Select(
		ctx,
		dr.db,
		&dbFSRSReviewLogs,
		`SELECT rl.id,
    rl.term_id,
    rl.user_id,
    rl.difficulty,
	to_char(rl.due, 'YYYY-MM-DD"T"HH24:MI:SS.MSTZH:TZM') as due,
    rl.learning_steps,
    rl.rating,
	to_char(rl.review, 'YYYY-MM-DD"T"HH24:MI:SS.MSTZH:TZM') as review,
    rl.scheduled_days,
    rl.stability,
    rl.state
FROM unnest($1::uuid[]) WITH ORDINALITY AS input(term_id, og_order)
LEFT JOIN fsrs_review_logs rl
	ON rl.term_id = input.term_id
	AND rl.user_id = $2
ORDER BY input.og_order ASC, rl.review DESC`,
		termIDs,
		authedUser.ID,
	)
	if err != nil {
		return nil, []error{err}
	}

	grouped := make(map[string][]*model.FSRSReviewLog)
	for _, rl := range dbFSRSReviewLogs {
		if rl.ID != nil && rl.TermID != nil {
			grouped[*rl.TermID] = append(grouped[*rl.TermID], &model.FSRSReviewLog{
				ID:            *rl.ID,
				Difficulty:    *rl.Difficulty,
				Due:           *rl.Due,
				LearningSteps: *rl.LearningSteps,
				Rating:        *rl.Rating,
				Review:        *rl.Review,
				ScheduledDays: *rl.ScheduledDays,
				Stability:     *rl.Stability,
				State:         *rl.State,
			})
		}
	}

	orderedFSRSReviewLogs := make([][]*model.FSRSReviewLog, len(termIDs))
	for i, id := range termIDs {
		orderedFSRSReviewLogs[i] = grouped[id]
	}

	return orderedFSRSReviewLogs, nil
}

func GetFSRSCardByTermID(ctx context.Context, termID string) (*model.FSRSCard, error) {
	loaders := For(ctx)
	return loaders.FSRSCardByTermIDLoader.Load(ctx, termID)
}
func GETFSRSCardsByTermIDs(ctx context.Context, termIDs []string) ([]*model.FSRSCard, error) {
	loaders := For(ctx)
	return loaders.FSRSCardByTermIDLoader.LoadAll(ctx, termIDs)
}

func GetFSRSReviewLogsByTermID(ctx context.Context, termID string) ([]*model.FSRSReviewLog, error) {
	loaders := For(ctx)
	return loaders.FSRSReviewLogsByTermIDLoader.Load(ctx, termID)
}
func GetFSRSReviewLogsByTermIDs(ctx context.Context, termIDs []string) ([][]*model.FSRSReviewLog, error) {
	loaders := For(ctx)
	return loaders.FSRSReviewLogsByTermIDLoader.LoadAll(ctx, termIDs)
}
