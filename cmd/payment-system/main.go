package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"infotecs-tz/internal/config"
	"infotecs-tz/internal/http-server/handlers/get_balance"
	"infotecs-tz/internal/http-server/handlers/get_last"
	"infotecs-tz/internal/http-server/handlers/send"
	"infotecs-tz/internal/storage/sqlite"
	"infotecs-tz/internal/utils"
	"log/slog"
	"net/http"
	"os"
)

const (
	envLocal = "local"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()
	log := setupLogger(cfg.Env)

	log.Info("starting main process")
	storage, err := sqlite.New(cfg.StoragePath)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
	log.Info("storage is set")

	empty, err := storage.IsEmpty()
	if err != nil {
		log.Error(err.Error())
	}

	if empty {
		log.Info("table is empty. generating addresses...")
		addresses, err := utils.GenerateHexAddresses(10)
		if err != nil {
			log.Error(err.Error())
		}
		for _, address := range addresses {
			err := storage.AddAddress(address, 100)
			if err != nil {
				log.Error(err.Error())
			}
		}
	}

	log.Info(cfg.HTTPServer)

	router := chi.NewRouter()

	//router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Get("/api/wallet/{address}/balance", get_balance.New(log, storage))
	router.Post("/api/send", send.New(log, storage))
	router.Get("/api/transactions", get_last.New(log, storage))

	log.Info("server listening requests")
	err = http.ListenAndServe(cfg.HTTPServer, router)
	if err != nil {
		log.Error("failed to start server: ", err.Error())
	}
}

func setupLogger(env string) *slog.Logger {

	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}
