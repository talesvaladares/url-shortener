package api

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewHandler(db map[string]string) http.Handler {
	routes := chi.NewMux()

	routes.Use(middleware.Recoverer)
	routes.Use(middleware.RequestID)
	routes.Use(middleware.Logger)

	routes.Post("/api/shorten", handlePost(db))
	routes.Get("/{code}", handleGet(db))

	return routes
}

type PostBody struct {
	URL string `json:"url"`
}

type Response struct {
	Error string `json:"error,omitempty"`
	Data  any    `json:"data,omitempty"`
}

func sendJSON(w http.ResponseWriter, res Response, status int) {
	data, err := json.Marshal(res)
	if err != nil {
		sendJSON(w, Response{Error: "something went wrong"}, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(status)
	if _, err := w.Write(data); err != nil {
		slog.Error("Error ao fazer marshal de json", "error", err)
		return
	}
}

func handlePost(db map[string]string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body PostBody
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			sendJSON(w, Response{Error: "invalid body"}, http.StatusUnprocessableEntity)
		}

		if _, err := url.Parse(body.URL); err != nil {
			sendJSON(w, Response{Error: "invalid url passed"}, http.StatusBadRequest)
		}

		code := genCode()
		db[code] = body.URL

		fmt.Println("code", code)
		sendJSON(w, Response{Data: code}, http.StatusCreated)
	}
}

func handleGet(db map[string]string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := chi.URLParam(r, "code")

		data, ok := db[code]

		if !ok {
			http.Error(w, "url não encontrada", http.StatusNotFound)
			return
		}

		http.Redirect(w, r, data, http.StatusPermanentRedirect)
	}
}

func genCode() string {
	const characteres = "qwertyuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM0123456789"
	const n = 8
	bytes := make([]byte, 8)

	for index := range n {
		bytes[index] = characteres[rand.IntN(len(characteres))]
	}

	return string(bytes)
}
