package server

import (
	"net/http"
	"quizfreely/api/auth"
	qzfrAPIConfig "quizfreely/api/config"
	"quizfreely/api/graph"
	"quizfreely/api/graph/loader"
	"quizfreely/api/graph/resolver"
	"quizfreely/api/rest"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vektah/gqlparser/v2/ast"
)

func NewRouter(config qzfrAPIConfig.Config, dbPool *pgxpool.Pool, s3Client *s3.Client) http.Handler {
	router := chi.NewRouter()

	authHandler := &auth.AuthHandler{DB: dbPool}
	restHandler := &rest.RESTHandler{
		DB:                 dbPool,
		Storage:            s3Client,
		UsercontentBucket:  &config.UsercontentBucket,
		UsercontentBaseURL: &config.UsercontentBaseURL,
	}

	router.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})
	router.Get("/health", restHandler.Health)
	router.Post(
		"/v0/auth/sign-up",
		authHandler.SignUp,
	)
	router.Post(
		"/v0/auth/sign-in",
		authHandler.SignIn,
	)
	router.Post(
		"/v0/auth/sign-out",
		authHandler.SignOut,
	)
	router.With(
		authHandler.AuthMiddleware,
	).Post(
		"/v0/auth/delete-account",
		authHandler.DeleteAccount,
	)

	if config.EnableOAuthGoogle {
		/* init oauth config here,
		after config is loaded */
		auth.InitOAuthGoogle(config)

		router.Get(
			"/oauth/google",
			authHandler.OAuthGoogleRedirect,
		)
		router.Get(
			"/oauth/google/callback",
			authHandler.OAuthGoogleCallback,
		)
	}

	basePath := config.BasePath
	if len(basePath) > 0 && basePath[len(basePath)-1] == '/' {
		/* if BASE_PATH ends with "/", remove it */
		basePath = basePath[:(len(basePath) - 1)]
	}

	router.Group(func(r chi.Router) {
		r.Use(authHandler.AuthMiddleware)

		h := handler.New(graph.NewExecutableSchema(graph.Config{Resolvers: &resolver.Resolver{
			DB:                 dbPool,
			UsercontentBaseURL: &config.UsercontentBaseURL,
		}}))

		h.AddTransport(transport.Options{})
		h.AddTransport(transport.GET{})
		h.AddTransport(transport.POST{})

		h.SetQueryCache(lru.New[*ast.QueryDocument](1000))

		h.Use(extension.Introspection{})
		h.Use(extension.FixedComplexityLimit(100))

		srv := loader.Middleware(dbPool, &config.UsercontentBaseURL, h)

		r.Handle(
			"/graphiql",
			playground.Handler(
				"Quizfreely API GraphiQL",
				basePath+"/graphql",
			),
		)
		r.Handle("/graphql", srv)
	})

	router.Group(func(r chi.Router) {
		r.Use(authHandler.AuthMiddleware)

		r.Get(
			"/v0/search-queries",
			restHandler.GetSearchQueries,
		)

		if s3Client != nil {
			router.With(
				authHandler.AuthMiddleware,
			).Put(
				"/term-images/{termID}/{side}",
				restHandler.UploadTermImage,
			)
		}
		router.With(
			authHandler.AuthMiddleware,
		).Delete(
			"/term-images/{termID}/{side}",
			restHandler.RemoveTermImage,
		)
	})

	return router
}
