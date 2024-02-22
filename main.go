package main

import (
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aspiringVegetarian/chirpy_go_web_server/internal/database"
	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

type apiConfig struct {
	fileserverHits int
	chirpyDatabase database.DB
	jwtSecret      string
	revokedTokens  map[string]time.Time
}

func main() {

	const port = "8080"
	const filePathRoot = "."
	const dbFilePath = "./chirpy_database.json"

	godotenv.Load()
	jwtSecret := os.Getenv("JWT_SECRET")

	dbg := flag.Bool("debug", false, "Enable debug mode")
	flag.Parse()

	if *dbg {
		_, err := os.Stat(dbFilePath)
		if !errors.Is(err, os.ErrNotExist) {
			err := os.Remove(dbFilePath)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	chirpyDB, err := database.NewDB(dbFilePath)
	if err != nil {
		log.Fatal("Failed to init database")
	}

	apiCfg := apiConfig{
		fileserverHits: 0,
		chirpyDatabase: *chirpyDB,
		jwtSecret:      jwtSecret,
		revokedTokens:  make(map[string]time.Time),
	}

	// File server routing /app and /app/*

	router := chi.NewRouter()

	fileServerHandler := apiCfg.middlewareMetricsIncrementer(http.StripPrefix("/app", http.FileServer(http.Dir(filePathRoot))))

	router.Handle("/app", fileServerHandler)

	router.Handle("/app/*", fileServerHandler)

	// API routing

	rApi := chi.NewRouter()

	rApi.Get("/healthz", healthHandler)

	rApi.Get("/reset", apiCfg.metricsReset)

	rApi.Get("/chirps", apiCfg.getChirpsHandler)

	rApi.Get("/chirps/{chirpID}", apiCfg.getChirpHandler)

	rApi.Post("/chirps", apiCfg.postChirpHandler)

	rApi.Post("/users", apiCfg.postUserHandler)

	rApi.Put("/users", apiCfg.putUserHandler)

	rApi.Post("/login", apiCfg.postLoginHandler)

	rApi.Post("/refresh", apiCfg.postRefreshHandler)

	rApi.Post("/revoke", apiCfg.postRevokeTokenHandler)

	router.Mount("/api", rApi)

	// Admin routing

	rAdmin := chi.NewRouter()

	rAdmin.Get("/metrics", apiCfg.metricsHandler)

	rAdmin.Get("/dbreset", apiCfg.chirpyDatabase.DatabaseResetHandler)

	router.Mount("/admin", rAdmin)

	// Fix headers with Cors middleware

	corsMux := middlewareCors(router)

	srv := &http.Server{
		Addr:    "localhost:" + port,
		Handler: corsMux,
	}

	log.Printf("Serving on port: %s\n", port)
	log.Fatal(srv.ListenAndServe())
}
