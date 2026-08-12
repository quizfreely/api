package loader

import (
	"context"
	"quizfreely/api/auth"
	// "quizfreely/api/graph/model"

	"github.com/georgysavva/scany/v2/pgxscan"
)

func (dr *dataReader) getMatchActivityTermIDs(ctx context.Context, matchActivityIDs []string) ([][]string, []error) {
	authedUser := auth.AuthedUserContext(ctx)
	if authedUser == nil || authedUser.ID == nil {
		results := make([][]string, len(matchActivityIDs))
		return results, nil
	}

	type dbRow struct {
		MatchActivityID string `db:"match_activity_id"`
		TermID          string `db:"term_id"`
	}
	var rows []dbRow

	err := pgxscan.Select(
		ctx,
		dr.db,
		&rows,
		`SELECT re.match_activity_id, re.term_id
FROM review_events re
WHERE re.match_activity_id = ANY($1::uuid[]) AND re.correct = true AND re.user_id = $2
ORDER BY re.match_activity_id`,
		matchActivityIDs,
		authedUser.ID,
	)
	if err != nil {
		return nil, []error{err}
	}

	grouped := make(map[string][]string)
	for _, row := range rows {
		grouped[row.MatchActivityID] = append(grouped[row.MatchActivityID], row.TermID)
	}

	ordered := make([][]string, len(matchActivityIDs))
	for i, id := range matchActivityIDs {
		ordered[i] = grouped[id]
		if ordered[i] == nil {
			ordered[i] = []string{}
		}
	}

	return ordered, nil
}

func (dr *dataReader) getMatchActivityIncorrectPairIDs(ctx context.Context, matchActivityIDs []string) ([][][]string, []error) {
	authedUser := auth.AuthedUserContext(ctx)
	if authedUser == nil || authedUser.ID == nil {
		results := make([][][]string, len(matchActivityIDs))
		return results, nil
	}

	type dbRow struct {
		MatchActivityID string `db:"match_activity_id"`
		TermID          string `db:"term_id"`
		AnsweredTermID  string `db:"answered_term_id"`
	}
	var rows []dbRow

	err := pgxscan.Select(
		ctx,
		dr.db,
		&rows,
		`SELECT re.match_activity_id, re.term_id, re.answered_term_id
FROM review_events re
WHERE re.match_activity_id = ANY($1::uuid[]) AND re.correct = false AND re.user_id = $2
ORDER BY re.match_activity_id`,
		matchActivityIDs,
		authedUser.ID,
	)
	if err != nil {
		return nil, []error{err}
	}

	grouped := make(map[string][][]string)
	for _, row := range rows {
		grouped[row.MatchActivityID] = append(grouped[row.MatchActivityID], []string{row.TermID, row.AnsweredTermID})
	}

	ordered := make([][][]string, len(matchActivityIDs))
	for i, id := range matchActivityIDs {
		ordered[i] = grouped[id]
		if ordered[i] == nil {
			ordered[i] = [][]string{}
		}
	}

	return ordered, nil
}

func (dr *dataReader) getMatchActivityStudysetIDs(ctx context.Context, matchActivityIDs []string) ([][]string, []error) {
	authedUser := auth.AuthedUserContext(ctx)
	if authedUser == nil || authedUser.ID == nil {
		emptyRes := make([][]string, len(matchActivityIDs))
		return emptyRes, nil
	}

	type dbRow struct {
		MatchActivityID string `db:"match_id"`
		StudysetID      string `db:"studyset_id"`
	}
	var rows []dbRow

	err := pgxscan.Select(
		ctx,
		dr.db,
		&rows,
		`SELECT mas.match_id, mas.studyset_id
FROM match_activity_studysets mas
JOIN match_activities ma ON ma.id = mas.match_id
WHERE mas.match_id = ANY($1::uuid[])
	AND ma.user_id = $2::uuid
ORDER BY mas.match_id`,
		matchActivityIDs,
		authedUser.ID,
	)
	if err != nil {
		return nil, []error{err}
	}

	grouped := make(map[string][]string)
	for _, row := range rows {
		grouped[row.MatchActivityID] = append(grouped[row.MatchActivityID], row.StudysetID)
	}

	ordered := make([][]string, len(matchActivityIDs))
	for i, id := range matchActivityIDs {
		ordered[i] = grouped[id]
		if ordered[i] == nil {
			ordered[i] = []string{}
		}
	}

	return ordered, nil
}

func GetMatchActivityTermIDs(ctx context.Context, matchActivityID string) ([]string, error) {
	loaders := For(ctx)
	return loaders.MatchActivityTermIDsLoader.Load(ctx, matchActivityID)
}
func GetMatchActivityTermIDsByIDs(ctx context.Context, matchActivityIDs []string) ([][]string, error) {
	loaders := For(ctx)
	return loaders.MatchActivityTermIDsLoader.LoadAll(ctx, matchActivityIDs)
}

func GetMatchActivityIncorrectPairIDs(ctx context.Context, matchActivityID string) ([][]string, error) {
	loaders := For(ctx)
	return loaders.MatchActivityIncorrectPairIDsLoader.Load(ctx, matchActivityID)
}
func GetMatchActivityIncorrectPairIDsByIDs(ctx context.Context, matchActivityIDs []string) ([][][]string, error) {
	loaders := For(ctx)
	return loaders.MatchActivityIncorrectPairIDsLoader.LoadAll(ctx, matchActivityIDs)
}

func GetMatchActivityStudysetIDs(ctx context.Context, matchActivityID string) ([]string, error) {
	loaders := For(ctx)
	return loaders.MatchActivityStudysetIDsLoader.Load(ctx, matchActivityID)
}
func GetMatchActivityStudysetIDsByIDs(ctx context.Context, matchActivityIDs []string) ([][]string, error) {
	loaders := For(ctx)
	return loaders.MatchActivityStudysetIDsLoader.LoadAll(ctx, matchActivityIDs)
}
