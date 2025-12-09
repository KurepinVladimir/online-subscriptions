package main

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	"github.com/KurepinVladimir/online-subscriptions/internal/config"
	"github.com/KurepinVladimir/online-subscriptions/internal/handler"
	"github.com/KurepinVladimir/online-subscriptions/internal/logger"
	"github.com/KurepinVladimir/online-subscriptions/internal/repository"
	"github.com/KurepinVladimir/online-subscriptions/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	logg, err := logger.New(cfg.Logging.Level)
	if err != nil {
		log.Fatalf("init logger: %v", err)
	}
	defer logg.Sync()

	logg.Info("starting subscriptions service",
		zap.String("addr", cfg.Server.Addr),
	)

	db := openDB(cfg.Database.DSN, logg)
	defer db.Close()

	if err := repository.RunMigrations(db, "migrations"); err != nil {
		logg.Fatal("run migrations", zap.Error(err))
	}

	repo := repository.NewPostgresRepository(db)
	svc := service.NewSubscriptionService(repo)
	h := handler.NewSubscriptionHandler(svc, logg)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)

	h.RegisterRoutes(r)

	srv := &http.Server{
		Addr:         cfg.Server.Addr,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logg.Fatal("server failed", zap.Error(err))
	}
}

func openDB(dsn string, logger *zap.Logger) *sqlx.DB {
	sql.Register("pgx", stdlib.GetDefaultDriver())

	db, err := sqlx.Open("pgx", dsn)
	if err != nil {
		logger.Fatal("open db", zap.Error(err))
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	if err := db.Ping(); err != nil {
		logger.Fatal("ping db", zap.Error(err))
	}

	return db
}
