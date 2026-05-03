package loader

// import vikstrous/dataloadgen with your other imports
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

type ctxKey string

const (
	loadersKey = ctxKey("dataloaders")
)

type dataReader struct {
	db                 *pgxpool.Pool
	usercontentBaseURL *string
}

// getUsers implements a batch function that can retrieve many users by ID,
// for use in a dataloader
func (dr *dataReader) getUsers(ctx context.Context, userIDs []string) ([]*model.User, []error) {
	type dbUser struct {
		ID          *string `db:"id"`
		Username    *string `db:"username"`
		DisplayName *string `db:"display_name"`
	}
	var dbUsers []*dbUser

	err := pgxscan.Select(
		ctx,
		dr.db,
		&dbUsers,
		`SELECT u.id, u.username, u.display_name
FROM unnest($1::uuid[]) WITH ORDINALITY AS input(id, og_order)
LEFT JOIN auth.users u ON u.id = input.id
ORDER BY input.og_order`,
		userIDs,
	)
	if err != nil {
		return nil, []error{err}
	}

	users := make([]*model.User, len(dbUsers))
	for i, du := range dbUsers {
		if du.ID == nil {
			users[i] = nil
		} else {
			users[i] = &model.User{
				ID:          du.ID,
				Username:    du.Username,
				DisplayName: du.DisplayName,
			}
		}
	}

	return users, nil
}

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
	tp.term_leitner_system_box, tp.def_leitner_system_box,
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
		TermLeitnerSystemBox *int32  `db:"term_leitner_system_box"`
		DefLeitnerSystemBox  *int32  `db:"def_leitner_system_box"`
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
			termsProgress = append(termsProgress, &model.TermProgress{
				ID:                   tp.ID,
				TermFirstReviewedAt:  tp.TermFirstReviewedAt,
				TermLastReviewedAt:   tp.TermLastReviewedAt,
				TermReviewCount:      tp.TermReviewCount,
				DefFirstReviewedAt:   tp.DefFirstReviewedAt,
				DefLastReviewedAt:    tp.DefLastReviewedAt,
				DefReviewCount:       tp.DefReviewCount,
				TermLeitnerSystemBox: tp.TermLeitnerSystemBox,
				DefLeitnerSystemBox:  tp.DefLeitnerSystemBox,
				TermCorrectCount:     *tp.TermCorrectCount,
				TermIncorrectCount:   *tp.TermIncorrectCount,
				DefCorrectCount:      *tp.DefCorrectCount,
				DefIncorrectCount:    *tp.DefIncorrectCount,
			})
		}
	}

	return termsProgress, nil
}

func (dr *dataReader) getTermsProgressHistory(ctx context.Context, termIDs []string) ([][]*model.TermProgressHistory, []error) {
	authedUser := auth.AuthedUserContext(ctx)
	if authedUser == nil || authedUser.ID == nil {
		return nil, nil
	}

	type dbTermProgressHistory struct {
		ID                 *string `db:"id"`
		TermID             *string `db:"term_id"`
		Timestamp          *string `db:"timestamp"`
		TermCorrectCount   *int32  `db:"term_correct_count"`
		TermIncorrectCount *int32  `db:"term_incorrect_count"`
		DefCorrectCount    *int32  `db:"def_correct_count"`
		DefIncorrectCount  *int32  `db:"def_incorrect_count"`
	}

	var dbHistory []*dbTermProgressHistory
	err := pgxscan.Select(
		ctx,
		dr.db,
		&dbHistory,
		`SELECT tph.id,
    tph.term_id,
    to_char(tph.timestamp, 'YYYY-MM-DD"T"HH24:MI:SS.MSTZH:TZM') as timestamp,
    tph.term_correct_count,
    tph.term_incorrect_count,
    tph.def_correct_count,
    tph.def_incorrect_count
FROM unnest($1::uuid[]) WITH ORDINALITY AS input(term_id, og_order)
LEFT JOIN term_progress_history tph
	ON tph.term_id = input.term_id
	AND tph.user_id = $2
ORDER BY input.og_order ASC, tph.timestamp DESC`,
		termIDs,
		authedUser.ID,
	)
	if err != nil {
		return nil, []error{err}
	}

	grouped := make(map[string][]*model.TermProgressHistory)
	for _, dbTph := range dbHistory {
		if dbTph.ID != nil && dbTph.TermID != nil {
			grouped[*dbTph.TermID] = append(grouped[*dbTph.TermID], &model.TermProgressHistory{
				ID:                 dbTph.ID,
				Timestamp:          dbTph.Timestamp,
				TermCorrectCount:   dbTph.TermCorrectCount,
				TermIncorrectCount: dbTph.TermIncorrectCount,
				DefCorrectCount:    dbTph.DefCorrectCount,
				DefIncorrectCount:  dbTph.DefIncorrectCount,
			})
		}
	}

	orderedProgressHistory := make([][]*model.TermProgressHistory, len(termIDs))
	for i, id := range termIDs {
		orderedProgressHistory[i] = grouped[id]
	}

	return orderedProgressHistory, nil
}

func (dr *dataReader) getTermsTopConfusionPairs(ctx context.Context, termIDs []string) ([][]*model.TermConfusionPair, []error) {
	authedUser := auth.AuthedUserContext(ctx)
	if authedUser == nil || authedUser.ID == nil {
		return nil, nil
	}

	type dbTermConfusionPair struct {
		ID             *string `db:"id"`
		TermID         *string `db:"term_id"`
		ConfusedTermID *string `db:"confused_term_id"`
		AnsweredWith   *string `db:"answered_with"`
		ConfusedCount  *int32  `db:"confused_count"`
		LastConfusedAt *string `db:"last_confused_at"`
	}

	var dbPairs []*dbTermConfusionPair
	err := pgxscan.Select(
		ctx,
		dr.db,
		&dbPairs,
		`SELECT id,
    term_id,
    confused_term_id,
    answered_with,
    confused_count,
    to_char(last_confused_at, 'YYYY-MM-DD"T"HH24:MI:SS.MSTZH:TZM') as last_confused_at
FROM (
    SELECT tcp.*,
        ROW_NUMBER() OVER (
            PARTITION BY tcp.term_id
            ORDER BY tcp.confused_count DESC
        ) AS rn
    FROM unnest($1::uuid[]) WITH ORDINALITY AS input(term_id, og_order)
    LEFT JOIN term_confusion_pairs tcp
        ON tcp.term_id = input.term_id
       AND tcp.user_id = $2
) ranked
WHERE rn <= 3
ORDER BY term_id, confused_count DESC`,
		termIDs,
		authedUser.ID,
	)
	if err != nil {
		return nil, []error{err}
	}

	grouped := make(map[string][]*model.TermConfusionPair)
	for _, dbTcp := range dbPairs {
		if dbTcp.ID != nil && dbTcp.TermID != nil {
			answeredWith := model.AnswerWith(*dbTcp.AnsweredWith)
			grouped[*dbTcp.TermID] = append(grouped[*dbTcp.TermID], &model.TermConfusionPair{
				ID:             dbTcp.ID,
				TermID:         dbTcp.TermID,
				ConfusedTermID: dbTcp.ConfusedTermID,
				AnsweredWith:   &answeredWith,
				ConfusedCount:  dbTcp.ConfusedCount,
				LastConfusedAt: dbTcp.LastConfusedAt,
			})
		}
	}

	orderedConfusionPairs := make([][]*model.TermConfusionPair, len(termIDs))
	for i, id := range termIDs {
		orderedConfusionPairs[i] = grouped[id]
	}

	return orderedConfusionPairs, nil
}

func (dr *dataReader) getTermsTopReverseConfusionPairs(ctx context.Context, termIDs []string) ([][]*model.TermConfusionPair, []error) {
	authedUser := auth.AuthedUserContext(ctx)
	if authedUser == nil || authedUser.ID == nil {
		return nil, nil
	}

	type dbTermConfusionPair struct {
		ID             *string `db:"id"`
		TermID         *string `db:"term_id"`
		ConfusedTermID *string `db:"confused_term_id"`
		AnsweredWith   *string `db:"answered_with"`
		ConfusedCount  *int32  `db:"confused_count"`
		LastConfusedAt *string `db:"last_confused_at"`
	}

	var dbPairs []*dbTermConfusionPair
	err := pgxscan.Select(
		ctx,
		dr.db,
		&dbPairs,
		`SELECT id,
    term_id,
    confused_term_id,
    answered_with,
    confused_count,
    to_char(last_confused_at, 'YYYY-MM-DD"T"HH24:MI:SS.MSTZH:TZM') as last_confused_at
FROM (
    SELECT tcp.*,
        ROW_NUMBER() OVER (
            PARTITION BY tcp.confused_term_id
            ORDER BY tcp.confused_count DESC
        ) AS rn
    FROM unnest($1::uuid[]) WITH ORDINALITY AS input(term_id, og_order)
    LEFT JOIN term_confusion_pairs tcp
        ON tcp.confused_term_id = input.term_id
       AND tcp.user_id = $2
) ranked
WHERE rn <= 3
ORDER BY confused_term_id, confused_count DESC`,
		termIDs,
		authedUser.ID,
	)
	if err != nil {
		return nil, []error{err}
	}

	grouped := make(map[string][]*model.TermConfusionPair)
	for _, dbTcp := range dbPairs {
		if dbTcp.ID != nil && dbTcp.ConfusedTermID != nil {
			answeredWith := model.AnswerWith(*dbTcp.AnsweredWith)
			grouped[*dbTcp.ConfusedTermID] = append(grouped[*dbTcp.ConfusedTermID], &model.TermConfusionPair{
				ID:             dbTcp.ID,
				TermID:         dbTcp.TermID,
				ConfusedTermID: dbTcp.ConfusedTermID,
				AnsweredWith:   &answeredWith,
				ConfusedCount:  dbTcp.ConfusedCount,
				LastConfusedAt: dbTcp.LastConfusedAt,
			})
		}
	}

	orderedConfusionPairs := make([][]*model.TermConfusionPair, len(termIDs))
	for i, id := range termIDs {
		orderedConfusionPairs[i] = grouped[id]
	}

	return orderedConfusionPairs, nil
}

func (dr *dataReader) getPracticeTestsByStudysetIDs(ctx context.Context, studysetIDs []string) ([][]*model.PracticeTest, []error) {
	authedUser := auth.AuthedUserContext(ctx)
	if authedUser == nil || authedUser.ID == nil {
		return nil, nil
	}

	type dbPracticeTest struct {
		ID               *string           `db:"id"`
		Timestamp        *string           `db:"timestamp"`
		StudysetID       *string           `db:"studyset_id"`
		QuestionsCorrect *int32            `db:"questions_correct"`
		QuestionsTotal   *int32            `db:"questions_total"`
		Questions        []*model.Question `db:"questions"`
	}
	var dbPracticeTests []*dbPracticeTest

	err := pgxscan.Select(
		ctx,
		dr.db,
		&dbPracticeTests,
		`SELECT pt.id,
	to_char(pt.timestamp, 'YYYY-MM-DD"T"HH24:MI:SS.MSTZH:TZM') as timestamp,
	pt.studyset_id,
    pt.questions_correct,
    pt.questions_total,
    pt.questions
FROM unnest($1::uuid[]) WITH ORDINALITY AS input(studyset_id, og_order)
LEFT JOIN practice_tests pt
	ON pt.studyset_id = input.studyset_id
	AND pt.user_id = $2
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
				StudysetID:       pt.StudysetID,
				Timestamp:        pt.Timestamp,
				QuestionsCorrect: pt.QuestionsCorrect,
				QuestionsTotal:   pt.QuestionsTotal,
				Questions:        pt.Questions,
			})
		}
	}

	orderedPracticeTests := make([][]*model.PracticeTest, len(studysetIDs))
	for i, id := range studysetIDs {
		orderedPracticeTests[i] = grouped[id]
	}

	return orderedPracticeTests, nil
}

func (dr *dataReader) getTermsFSRSCards(ctx context.Context, termIDs []string) ([][]*model.TermProgressHistory, []error) {
	authedUser := auth.AuthedUserContext(ctx)
	if authedUser == nil || authedUser.ID == nil {
		return nil, nil
	}

	type dbTermProgressHistory struct {
		ID                 *string `db:"id"`
		TermID             *string `db:"term_id"`
		Timestamp          *string `db:"timestamp"`
		TermCorrectCount   *int32  `db:"term_correct_count"`
		TermIncorrectCount *int32  `db:"term_incorrect_count"`
		DefCorrectCount    *int32  `db:"def_correct_count"`
		DefIncorrectCount  *int32  `db:"def_incorrect_count"`
	}

	var dbHistory []*dbTermProgressHistory
	err := pgxscan.Select(
		ctx,
		dr.db,
		&dbHistory,
		`SELECT tph.id,
    tph.term_id,
    to_char(tph.timestamp, 'YYYY-MM-DD"T"HH24:MI:SS.MSTZH:TZM') as timestamp,
    tph.term_correct_count,
    tph.term_incorrect_count,
    tph.def_correct_count,
    tph.def_incorrect_count
FROM unnest($1::uuid[]) WITH ORDINALITY AS input(term_id, og_order)
LEFT JOIN term_progress_history tph
	ON tph.term_id = input.term_id
	AND tph.user_id = $2
ORDER BY input.og_order ASC, tph.timestamp DESC`,
		termIDs,
		authedUser.ID,
	)
	if err != nil {
		return nil, []error{err}
	}

	grouped := make(map[string][]*model.TermProgressHistory)
	for _, dbTph := range dbHistory {
		if dbTph.ID != nil && dbTph.TermID != nil {
			grouped[*dbTph.TermID] = append(grouped[*dbTph.TermID], &model.TermProgressHistory{
				ID:                 dbTph.ID,
				Timestamp:          dbTph.Timestamp,
				TermCorrectCount:   dbTph.TermCorrectCount,
				TermIncorrectCount: dbTph.TermIncorrectCount,
				DefCorrectCount:    dbTph.DefCorrectCount,
				DefIncorrectCount:  dbTph.DefIncorrectCount,
			})
		}
	}

	orderedProgressHistory := make([][]*model.TermProgressHistory, len(termIDs))
	for i, id := range termIDs {
		orderedProgressHistory[i] = grouped[id]
	}

	return orderedProgressHistory, nil
}

// Loaders wrap your data loaders to inject via middleware
type Loaders struct {
	UserLoader                         *dataloadgen.Loader[string, *model.User]
	TermByIDLoader                     *dataloadgen.Loader[string, *model.Term]
	TermByStudysetIDLoader             *dataloadgen.Loader[string, []*model.Term]
	TermsCountByStudysetIDLoader       *dataloadgen.Loader[string, *int32]
	TermProgressLoader                 *dataloadgen.Loader[string, *model.TermProgress]
	TermProgressHistoryLoader          *dataloadgen.Loader[string, []*model.TermProgressHistory]
	TermTopConfusionPairsLoader        *dataloadgen.Loader[string, []*model.TermConfusionPair]
	TermTopReverseConfusionPairsLoader *dataloadgen.Loader[string, []*model.TermConfusionPair]
	PracticeTestByStudysetIDLoader     *dataloadgen.Loader[string, []*model.PracticeTest]
}

// NewLoaders instantiates data loaders for the middleware
func NewLoaders(db *pgxpool.Pool, usercontentBaseURL *string) *Loaders {
	// define the data loader
	dr := &dataReader{
		db:                 db,
		usercontentBaseURL: usercontentBaseURL,
	}
	return &Loaders{
		UserLoader:                         dataloadgen.NewLoader(dr.getUsers, dataloadgen.WithWait(time.Millisecond)),
		TermByIDLoader:                     dataloadgen.NewLoader(dr.getTermsByIDs, dataloadgen.WithWait(time.Millisecond)),
		TermByStudysetIDLoader:             dataloadgen.NewLoader(dr.getTermsByStudysetIDs, dataloadgen.WithWait(time.Millisecond)),
		TermsCountByStudysetIDLoader:       dataloadgen.NewLoader(dr.getTermsCountByStudysetIDs, dataloadgen.WithWait(time.Millisecond)),
		TermProgressLoader:                 dataloadgen.NewLoader(dr.getTermsProgress, dataloadgen.WithWait(time.Millisecond)),
		TermProgressHistoryLoader:          dataloadgen.NewLoader(dr.getTermsProgressHistory, dataloadgen.WithWait(time.Millisecond)),
		TermTopConfusionPairsLoader:        dataloadgen.NewLoader(dr.getTermsTopConfusionPairs, dataloadgen.WithWait(time.Millisecond)),
		TermTopReverseConfusionPairsLoader: dataloadgen.NewLoader(dr.getTermsTopReverseConfusionPairs, dataloadgen.WithWait(time.Millisecond)),
		PracticeTestByStudysetIDLoader:     dataloadgen.NewLoader(dr.getPracticeTestsByStudysetIDs, dataloadgen.WithWait(time.Millisecond)),
	}
}

// Middleware injects data loaders into the context
func Middleware(db *pgxpool.Pool, usercontentBaseURL *string, next http.Handler) http.Handler {
	// return a middleware that injects the loader to the request context
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		loader := NewLoaders(db, usercontentBaseURL)
		r = r.WithContext(context.WithValue(r.Context(), loadersKey, loader))
		next.ServeHTTP(w, r)
	})
}

// For returns the dataloader for a given context
func For(ctx context.Context) *Loaders {
	return ctx.Value(loadersKey).(*Loaders)
}

// GetUser returns single user by id efficiently
func GetUser(ctx context.Context, userID string) (*model.User, error) {
	loaders := For(ctx)
	return loaders.UserLoader.Load(ctx, userID)
}

// GetUsers returns many users by ids efficiently
func GetUsers(ctx context.Context, userIDs []string) ([]*model.User, error) {
	loaders := For(ctx)
	return loaders.UserLoader.LoadAll(ctx, userIDs)
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

// GetTermProgressHistory returns a single term's progress history
func GetTermProgressHistory(ctx context.Context, termID string) ([]*model.TermProgressHistory, error) {
	loaders := For(ctx)
	return loaders.TermProgressHistoryLoader.Load(ctx, termID)
}

// GetTermsProgressHistory returns many terms' progress histories
func GetTermsProgressHistory(ctx context.Context, termIDs []string) ([][]*model.TermProgressHistory, error) {
	loaders := For(ctx)
	return loaders.TermProgressHistoryLoader.LoadAll(ctx, termIDs)
}

// GetTermTopReverseConfusionPairs returns a single term's confusion pairs
func GetTermTopReverseConfusionPairs(ctx context.Context, termID string) ([]*model.TermConfusionPair, error) {
	loaders := For(ctx)
	return loaders.TermTopReverseConfusionPairsLoader.Load(ctx, termID)
}

// GetTermsTopReverseConfusionPairs returns many terms' confusion pairs
func GetTermsTopReverseConfusionPairs(ctx context.Context, termIDs []string) ([][]*model.TermConfusionPair, error) {
	loaders := For(ctx)
	return loaders.TermTopReverseConfusionPairsLoader.LoadAll(ctx, termIDs)
}

// GetTermTopConfusionPairs returns a single term's confusion pairs
func GetTermTopConfusionPairs(ctx context.Context, termID string) ([]*model.TermConfusionPair, error) {
	loaders := For(ctx)
	return loaders.TermTopConfusionPairsLoader.Load(ctx, termID)
}

// GetTermsTopConfusionPairs returns many terms' confusion pairs
func GetTermsTopConfusionPairs(ctx context.Context, termIDs []string) ([][]*model.TermConfusionPair, error) {
	loaders := For(ctx)
	return loaders.TermTopConfusionPairsLoader.LoadAll(ctx, termIDs)
}

func GetPracticeTestsByStudysetID(ctx context.Context, studysetID string) ([]*model.PracticeTest, error) {
	loaders := For(ctx)
	return loaders.PracticeTestByStudysetIDLoader.Load(ctx, studysetID)
}

func GetPracticeTestsByStudysetIDs(ctx context.Context, studysetIDs []string) ([][]*model.PracticeTest, error) {
	loaders := For(ctx)
	return loaders.PracticeTestByStudysetIDLoader.LoadAll(ctx, studysetIDs)
}

