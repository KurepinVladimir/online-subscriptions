package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/jackc/pgx/v5/stdlib"
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

	db, err := openDB(cfg.Database.DSN)
	if err != nil {
		logg.Fatal("open db", zap.Error(err))
	}
	defer func() {
		_ = db.Close()
	}()

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

	// запуск сервера
	go func() {
		logg.Info("http server started", zap.String("addr", cfg.Server.Addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logg.Fatal("server failed", zap.Error(err))
		}
	}()

	// ожидание сигнала
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	logg.Info("shutdown started")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logg.Error("http shutdown error", zap.Error(err))
	} else {
		logg.Info("http shutdown completed")
	}

	logg.Info("shutdown finished")
}

func openDB(dsn string) (*sqlx.DB, error) {
	db, err := sqlx.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	// Настройки пула
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	// Ping с таймаутом, чтобы не зависнуть навсегда
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}
