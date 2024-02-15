package main

import (
	"log"
	"net/http"
)

func main() {

	const port = "8080"
	const filePathRoot = "."

	apiCfg := apiConfig{fileserverHits: 0}

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

	srv := &http.Server{
		Addr:    "localhost:" + port,
		Handler: corsMux,
	}

	log.Printf("Serving on port: %s\n", port)
	log.Fatal(srv.ListenAndServe())
}
