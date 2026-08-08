package middleware

import (
	"context"
	"net/http"
	"time"
	_ "time/tzdata"
)

type contextKey struct {
	name string
}

var tzCtxKey = &contextKey{"timezone"}

func TimezoneMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		cookie, err := r.Cookie("tz")
		if err == nil && cookie != nil {
			tz := cookie.Value
			if tz != "" {
				loc, err := time.LoadLocation(tz)
				if err == nil {
					tz = loc.String()
					ctx := context.WithValue(r.Context(), tzCtxKey, tz)
					r = r.WithContext(ctx)
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}

func TimezoneContext(ctx context.Context) *string {
	v, ok := ctx.Value(tzCtxKey).(string)
	if ok && v != "" {
		return &v
	} else {
		return nil
	}
}
