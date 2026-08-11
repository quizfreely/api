package loader

// import vikstrous/dataloadgen with your other imports
import (
	"context"
	"net/http"
	"quizfreely/api/graph/model"
	"time"

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

// Loaders wrap our data loaders to inject via middleware
type Loaders struct {
	UserLoader                          *dataloadgen.Loader[string, *model.User]
	TermByIDLoader                      *dataloadgen.Loader[string, *model.Term]
	TermByStudysetIDLoader              *dataloadgen.Loader[string, []*model.Term]
	TermProgressLoader                  *dataloadgen.Loader[string, *model.TermProgress]
	TermsCountByStudysetIDLoader        *dataloadgen.Loader[string, *int32]
	PracticeTestByStudysetIDLoader      *dataloadgen.Loader[string, []*model.PracticeTest]
	PracticeTestByTermIDLoader          *dataloadgen.Loader[string, []*model.PracticeTest]
	FSRSCardByTermIDLoader              *dataloadgen.Loader[string, *model.FSRSCard]
	FSRSReviewLogsByTermIDLoader        *dataloadgen.Loader[string, []*model.FSRSReviewLog]
	MatchActivityTermIDsLoader          *dataloadgen.Loader[string, []string]
	MatchActivityIncorrectPairIDsLoader *dataloadgen.Loader[string, [][]string]
	MatchActivityStudysetIDsLoader      *dataloadgen.Loader[string, []string]
}

// NewLoaders instantiates data loaders for the middleware
func NewLoaders(db *pgxpool.Pool, usercontentBaseURL *string) *Loaders {
	// define the data loader
	dr := &dataReader{
		db:                 db,
		usercontentBaseURL: usercontentBaseURL,
	}
	return &Loaders{
		/* loader/user.go */
		UserLoader: dataloadgen.NewLoader(dr.getUsers, dataloadgen.WithWait(time.Millisecond)),
		/* loader/term.go */
		TermByIDLoader:         dataloadgen.NewLoader(dr.getTermsByIDs, dataloadgen.WithWait(time.Millisecond)),
		TermByStudysetIDLoader: dataloadgen.NewLoader(dr.getTermsByStudysetIDs, dataloadgen.WithWait(time.Millisecond)),
		TermProgressLoader:     dataloadgen.NewLoader(dr.getTermsProgress, dataloadgen.WithWait(time.Millisecond)),
		/* loader/studyset.go */
		TermsCountByStudysetIDLoader: dataloadgen.NewLoader(dr.getTermsCountByStudysetIDs, dataloadgen.WithWait(time.Millisecond)),
		/* loader/pt.go */
		PracticeTestByStudysetIDLoader: dataloadgen.NewLoader(dr.getPracticeTestsByStudysetIDs, dataloadgen.WithWait(time.Millisecond)),
		PracticeTestByTermIDLoader:     dataloadgen.NewLoader(dr.getPracticeTestsByTermIDs, dataloadgen.WithWait(time.Millisecond)),
		/* loader/fsrs.go */
		FSRSCardByTermIDLoader:       dataloadgen.NewLoader(dr.getFSRSCardsByTermIDs, dataloadgen.WithWait(time.Millisecond)),
		FSRSReviewLogsByTermIDLoader: dataloadgen.NewLoader(dr.getFSRSReviewLogsByTermIDs, dataloadgen.WithWait(time.Millisecond)),
		/* loader/match.go */
		MatchActivityTermIDsLoader:          dataloadgen.NewLoader(dr.getMatchActivityTermIDs, dataloadgen.WithWait(time.Millisecond)),
		MatchActivityIncorrectPairIDsLoader: dataloadgen.NewLoader(dr.getMatchActivityIncorrectPairIDs, dataloadgen.WithWait(time.Millisecond)),
		MatchActivityStudysetIDsLoader:      dataloadgen.NewLoader(dr.getMatchActivityStudysetIDs, dataloadgen.WithWait(time.Millisecond)),
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
