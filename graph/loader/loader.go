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
	/* loader/user.go */
	UserLoader                          *dataloadgen.Loader[string, *model.User]
	/* loader/term.go */
	TermByIDLoader                      *dataloadgen.Loader[string, *model.Term]
	TermByStudysetIDLoader              *dataloadgen.Loader[string, []*model.Term]
	TermProgressLoader                  *dataloadgen.Loader[string, *model.TermProgress]
	/* loader/studyset.go */
	StudysetLoader                      *dataloadgen.Loader[string, *model.Studyset]
	TermsCountByStudysetIDLoader        *dataloadgen.Loader[string, *int32]
	/* loader/pt.go */
	PracticeTestByStudysetIDLoader      *dataloadgen.Loader[string, []*model.PracticeTest]
	PracticeTestByTermIDLoader          *dataloadgen.Loader[string, []*model.PracticeTest]
	PracticeTestStudysetIDsLoader       *dataloadgen.Loader[string, []string]
	/* loader/fsrs.go */
	FSRSCardByTermIDLoader              *dataloadgen.Loader[string, *model.FSRSCard]
	FSRSReviewLogsByTermIDLoader        *dataloadgen.Loader[string, []*model.FSRSReviewLog]
	/* loader/match.go */
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
	w := dataloadgen.WithWait(time.Millisecond)
	return &Loaders{
		/* loader/user.go */
		UserLoader:                          dataloadgen.NewLoader(dr.getUsers, w),
		/* loader/term.go */
		TermByIDLoader:                      dataloadgen.NewLoader(dr.getTermsByIDs, w),
		TermByStudysetIDLoader:              dataloadgen.NewLoader(dr.getTermsByStudysetIDs, w),
		TermProgressLoader:                  dataloadgen.NewLoader(dr.getTermsProgress, w),
		/* loader/studyset.go */
		StudysetLoader:                      dataloadgen.NewLoader(dr.getStudysetsByIDs, w),
		TermsCountByStudysetIDLoader:        dataloadgen.NewLoader(dr.getTermsCountByStudysetIDs, w),
		/* loader/pt.go */
		PracticeTestByStudysetIDLoader:      dataloadgen.NewLoader(dr.getPracticeTestsByStudysetIDs, w),
		PracticeTestByTermIDLoader:          dataloadgen.NewLoader(dr.getPracticeTestsByTermIDs, w),
		PracticeTestStudysetIDsLoader:       dataloadgen.NewLoader(dr.getPracticeTestStudysetIDs, w),
		/* loader/fsrs.go */
		FSRSCardByTermIDLoader:              dataloadgen.NewLoader(dr.getFSRSCardsByTermIDs, w),
		FSRSReviewLogsByTermIDLoader: 		 dataloadgen.NewLoader(dr.getFSRSReviewLogsByTermIDs, w),
		/* loader/match.go */
		MatchActivityTermIDsLoader:          dataloadgen.NewLoader(dr.getMatchActivityTermIDs, w),
		MatchActivityIncorrectPairIDsLoader: dataloadgen.NewLoader(dr.getMatchActivityIncorrectPairIDs, w),
		MatchActivityStudysetIDsLoader:      dataloadgen.NewLoader(dr.getMatchActivityStudysetIDs, w),
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
