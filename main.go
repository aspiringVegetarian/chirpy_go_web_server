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

	r := chi.NewRouter()

	fileServerHandler := apiCfg.middlewareMetricsIncrementer(http.StripPrefix("/app", http.FileServer(http.Dir(filePathRoot))))

	r.Handle("/app", fileServerHandler)

	r.Handle("/app/*", fileServerHandler)

	r.Get("/healthz", healthHandler)

	r.Get("/metrics", apiCfg.metricsHandler)

	r.Get("/reset", apiCfg.metricsReset)

	corsMux := middlewareCors(r)

	srv := &http.Server{
		Addr:    "localhost:" + port,
		Handler: corsMux,
	}

	log.Printf("Serving on port: %s\n", port)
	log.Fatal(srv.ListenAndServe())
}
