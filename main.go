package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func main() {

	const port = "8080"
	const filePathRoot = "."

	apiCfg := apiConfig{fileserverHits: 0}

	/*
		mux := http.NewServeMux()

		mux.HandleFunc("/healthz", func(w http.ResponseWriter, req *http.Request) {

			w.Header().Set("Content-Type", "text/plain; charset=utf-8")

			w.WriteHeader(http.StatusOK)

			w.Write([]byte(http.StatusText(http.StatusOK)))

		})

		mux.HandleFunc("/metrics", apiCfg.metricsHandler)

		mux.HandleFunc("/reset", apiCfg.metricsReset)

		mux.Handle("/app/", apiCfg.middlewareMetricsIncrementer(http.StripPrefix("/app/", http.FileServer(http.Dir(filePathRoot)))))

		corsMux := middlewareCors(mux)
	*/

	router := chi.NewRouter()

	fileServerHandler := apiCfg.middlewareMetricsIncrementer(http.StripPrefix("/app", http.FileServer(http.Dir(filePathRoot))))

	router.Handle("/app", fileServerHandler)

	router.Handle("/app/*", fileServerHandler)

	rApi := chi.NewRouter()

	rApi.Get("/healthz", healthHandler)

	rApi.Get("/reset", apiCfg.metricsReset)

	rApi.Post("/validate_chirp", validateChirpHandler)

	router.Mount("/api", rApi)

	rAdmin := chi.NewRouter()

	rAdmin.Get("/metrics", apiCfg.metricsHandler)

	router.Mount("/admin", rAdmin)

	corsMux := middlewareCors(router)

	srv := &http.Server{
		Addr:    "localhost:" + port,
		Handler: corsMux,
	}

	log.Printf("Serving on port: %s\n", port)
	log.Fatal(srv.ListenAndServe())
}
