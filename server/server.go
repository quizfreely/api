package server

import (
	"net/http"
	"os"
	"quizfreely/api/auth"
	"quizfreely/api/graph"
	"quizfreely/api/graph/loader"
	"quizfreely/api/graph/resolver"
	"quizfreely/api/rest"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vektah/gqlparser/v2/ast"
)

func NewRouter(dbPool *pgxpool.Pool) http.Handler {
	router := chi.NewRouter()

	authHandler := &auth.AuthHandler{DB: dbPool}
	restHandler := &rest.RESTHandler{DB: dbPool}

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

	if os.Getenv("ENABLE_OAUTH_GOOGLE") == "true" {
		/* init oauth config here,
		after env vars are loaded */
		auth.InitOAuthGoogle()

		router.Get(
			"/oauth/google",
			authHandler.OAuthGoogleRedirect,
		)
		router.Get(
			"/oauth/google/callback",
			authHandler.OAuthGoogleCallback,
		)
	}

	basePath := os.Getenv("BASE_PATH")
	/* os.Getenv returns "" when not set,
	that's great cause we want to default to "" (blank)
	cause BASE_PATH is prepended/before relative paths */

	if len(basePath) > 0 && basePath[len(basePath)-1] == '/' {
		/* if BASE_PATH ends with "/", remove it */
		basePath = basePath[:(len(basePath) - 1)]
	}

	router.Group(func(r chi.Router) {
		r.Use(authHandler.AuthMiddleware)

		h := handler.New(graph.NewExecutableSchema(graph.Config{Resolvers: &resolver.Resolver{DB: dbPool}}))

		h.AddTransport(transport.Options{})
		h.AddTransport(transport.GET{})
		h.AddTransport(transport.POST{})

		h.SetQueryCache(lru.New[*ast.QueryDocument](1000))

		h.Use(extension.Introspection{})
		h.Use(extension.FixedComplexityLimit(50))

		srv := loader.Middleware(dbPool, h)

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
	})

	return router
}
