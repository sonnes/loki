package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/jmoiron/sqlx"
)

func AttachDB(db *sqlx.DB, fn func(*sqlx.DB, http.ResponseWriter, *http.Request)) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		fn(db, w, r)
	}
}

func CreateRouter(Db *sqlx.DB) *chi.Mux {
	mux := chi.NewMux()

	mux.Use(middleware.Recoverer)
	mux.Use(middleware.StripSlashes)
	mux.Use(middleware.NoCache)
	mux.Use(middleware.Heartbeat("/_ah/health"))
	mux.Use(middleware.RequestID)
	mux.Use(middleware.RealIP)
	mux.Use(middleware.Logger)

	mux.Route("/v1", func(api chi.Router) {

		jsonRequired := middleware.AllowContentType("application/json")

		api.With(jsonRequired).Post("/edges/init", AttachDB(Db, InitEdgeEndpoint))
		api.With(jsonRequired).Post("/edges/save", AttachDB(Db, SaveEdgesEndpoint))
		api.With(jsonRequired).Post("/edges/delete", AttachDB(Db, DeleteEdgesEndpoint))

	})

	return mux
}

func WriteJson(w http.ResponseWriter, v interface{}, code int) {
	b, err := json.Marshal(v)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(b)
}

type AppError struct {
	Code    int       `json:"code"`
	Message string    `json:"message"`
	Fields  *[]string `json:"fields"`
}

func WriteError(w http.ResponseWriter, appErr *AppError) {

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(appErr.Code)

	errorJson, err := json.Marshal(appErr)

	if err != nil {
		log.Fatal(err)
		log.Fatal("Error while writing an application error")
	}

	w.Write(errorJson)

}
