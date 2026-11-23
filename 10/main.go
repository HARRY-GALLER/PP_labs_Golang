package main

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"userapi/internal/api"
	"userapi/internal/auth"
	"userapi/internal/storage"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	dsn := getenv("DATABASE_URL", "postgres://postgres:postgres@localhost:54329/users?sslmode=disable")
	addr := ":" + getenv("PORT", "8080")

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		logger.Error("open db", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	store := storage.New(db)
	if err := db.PingContext(ctx); err != nil {
		cancel()
		logger.Error("ping db", "err", err)
		os.Exit(1)
	}
	if err := store.Migrate(ctx); err != nil {
		cancel()
		logger.Error("migrate", "err", err)
		os.Exit(1)
	}
	cancel()

	srv := &http.Server{
		Addr:              addr,
		Handler:           api.NewRouter(store, auth.NewManager(store), logger),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("server started", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("listen", "err", err)
			os.Exit(1)
		}
	}()

	stop, stopFn := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stopFn()
	<-stop.Done()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown", "err", err)
	}
	logger.Info("server stopped")
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
