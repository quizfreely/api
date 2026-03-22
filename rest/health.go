package rest

func (rh *RESTHandler) Health(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	err := rh.DB.Ping(ctx)
	if err == nil {
		w.Write([]byte("okayy"))
	} else {
		w.WriteHeader(503)
		w.Write([]byte("503 Service Unavailable: DB error, API running"))
	}
}
