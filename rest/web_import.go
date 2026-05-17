package rest

import (
	"net/http"
)

func (rh *RESTHandler) WebImport(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 90*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://", nil)
	resp, err := http.DefaultClient.Do(req)
}

