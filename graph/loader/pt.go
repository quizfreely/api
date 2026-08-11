// NOTE: filename can't be `practice_test.go` because that ends with `_test.go`

package loader

import (
	"context"
	"quizfreely/api/auth"
	"quizfreely/api/graph/model"

	"github.com/georgysavva/scany/v2/pgxscan"
)

func (dr *dataReader) getPracticeTestsByStudysetIDs(ctx context.Context, studysetIDs []string) ([][]*model.PracticeTest, []error) {
	authedUser := auth.AuthedUserContext(ctx)
	if authedUser == nil || authedUser.ID == nil {
		return nil, nil
	}

	type dbPracticeTest struct {
		ID               *string `db:"id"`
		Timestamp        *string `db:"timestamp"`
		QuestionsCorrect *int32  `db:"questions_correct"`
		QuestionsTotal   *int32  `db:"questions_total"`
		StudysetID       *string `db:"studyset_id"`
	}
	var dbPracticeTests []*dbPracticeTest

	err := pgxscan.Select(
		ctx,
		dr.db,
		&dbPracticeTests,
		`SELECT
	pt.id,
	to_char(pt.timestamp, 'YYYY-MM-DD"T"HH24:MI:SS.MSTZH:TZM') as timestamp,
    pt.questions_correct,
    pt.questions_total,
	input.studyset_id
FROM unnest($1::uuid[]) WITH ORDINALITY AS input(studyset_id, og_order)
JOIN practice_test_studysets pts ON pts.studyset_id = input.studyset_id
JOIN practice_tests pt ON pt.id = pts.practice_test_id
WHERE pt.user_id = $2
ORDER BY input.og_order ASC, pt.timestamp DESC`,
		studysetIDs,
		authedUser.ID,
	)
	if err != nil {
		return nil, []error{err}
	}

	grouped := make(map[string][]*model.PracticeTest)
	for _, pt := range dbPracticeTests {
		if pt.ID != nil && pt.StudysetID != nil {
			grouped[*pt.StudysetID] = append(grouped[*pt.StudysetID], &model.PracticeTest{
				ID:               pt.ID,
				Timestamp:        pt.Timestamp,
				QuestionsCorrect: pt.QuestionsCorrect,
				QuestionsTotal:   pt.QuestionsTotal,
			})
		}
	}

	orderedPracticeTests := make([][]*model.PracticeTest, len(studysetIDs))
	for i, id := range studysetIDs {
		orderedPracticeTests[i] = grouped[id]
	}

	return orderedPracticeTests, nil
}

func (dr *dataReader) getPracticeTestsByTermIDs(ctx context.Context, termIDs []string) ([][]*model.PracticeTest, []error) {
	authedUser := auth.AuthedUserContext(ctx)
	if authedUser == nil || authedUser.ID == nil {
		results := make([][]*model.PracticeTest, len(termIDs))
		for i := range termIDs {
			results[i] = []*model.PracticeTest{}
		}
		return results, nil
	}

	type dbPracticeTest struct {
		ID               *string `db:"id"`
		Timestamp        *string `db:"timestamp"`
		QuestionsCorrect *int32  `db:"questions_correct"`
		QuestionsTotal   *int32  `db:"questions_total"`
		TermID           *string `db:"term_id"`
	}
	var dbPracticeTests []*dbPracticeTest

	err := pgxscan.Select(
		ctx,
		dr.db,
		&dbPracticeTests,
		`SELECT pt.id,
	to_char(pt.timestamp, 'YYYY-MM-DD"T"HH24:MI:SS.MSTZH:TZM') as timestamp,
    pt.questions_correct,
    pt.questions_total,
    input.term_id
FROM unnest($1::uuid[]) WITH ORDINALITY AS input(term_id, og_order)
JOIN practice_test_questions ptq ON ptq.term_id = input.term_id
JOIN practice_tests pt ON pt.id = ptq.practice_test_id
WHERE pt.user_id = $2
GROUP BY input.term_id, input.og_order, pt.id
ORDER BY input.og_order ASC, pt.timestamp DESC`,
		termIDs,
		authedUser.ID,
	)
	if err != nil {
		return nil, []error{err}
	}

	grouped := make(map[string][]*model.PracticeTest)
	for _, pt := range dbPracticeTests {
		if pt.ID != nil && pt.TermID != nil {
			grouped[*pt.TermID] = append(grouped[*pt.TermID], &model.PracticeTest{
				ID:               pt.ID,
				Timestamp:        pt.Timestamp,
				QuestionsCorrect: pt.QuestionsCorrect,
				QuestionsTotal:   pt.QuestionsTotal,
			})
		}
	}

	orderedPracticeTests := make([][]*model.PracticeTest, len(termIDs))
	for i, id := range termIDs {
		orderedPracticeTests[i] = grouped[id]
		if orderedPracticeTests[i] == nil {
			orderedPracticeTests[i] = []*model.PracticeTest{}
		}
	}

	return orderedPracticeTests, nil
}

func GetPracticeTestsByStudysetID(ctx context.Context, studysetID string) ([]*model.PracticeTest, error) {
	loaders := For(ctx)
	return loaders.PracticeTestByStudysetIDLoader.Load(ctx, studysetID)
}
func GetPracticeTestsByStudysetIDs(ctx context.Context, studysetIDs []string) ([][]*model.PracticeTest, error) {
	loaders := For(ctx)
	return loaders.PracticeTestByStudysetIDLoader.LoadAll(ctx, studysetIDs)
}

func GetPracticeTestsByTermID(ctx context.Context, termID string) ([]*model.PracticeTest, error) {
	loaders := For(ctx)
	return loaders.PracticeTestByTermIDLoader.Load(ctx, termID)
}
func GetPracticeTestsByTermIDs(ctx context.Context, termIDs []string) ([][]*model.PracticeTest, error) {
	loaders := For(ctx)
	return loaders.PracticeTestByTermIDLoader.LoadAll(ctx, termIDs)
}
