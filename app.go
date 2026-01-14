package main

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
)

type ErrorMecha struct {
	Stats     *StatsInternal
	Server    *http.Server
	Mux       *mux.Router
	Ticker    *time.Ticker
	Value     int
	ErrorMark int
}

// NewErrorMecha sets up stats and the ticker
func NewErrorMecha() *ErrorMecha {
	em := &ErrorMecha{
		Stats:     NewStatsInternal(),
		Ticker:    time.NewTicker(1 * time.Second),
		Value:     100,
		ErrorMark: 0,
	}

	return em
}

func (em *ErrorMecha) SetupMux() *mux.Router {
	r := mux.NewRouter()

	r.Handle("/metrics", em.Stats.Handler())
	r.Handle("/healthz", http.HandlerFunc(em.healthzHandler))

	return r
}

func (em *ErrorMecha) healthzHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "text/plain")
	w.Write([]byte(`ok`))
}

func (em *ErrorMecha) createTable(conn *pgx.Conn, ctx context.Context) error {
	query := `CREATE TABLE IF NOT EXISTS mecha (
    id SERIAL PRIMARY KEY,
    number INTEGER DEFAULT 0
)`
	_, err := conn.Exec(ctx, query)
	if err != nil {
		slog.Error("Error creating table: %v\n", err)
		return fmt.Errorf("could not create table: %w", err)
	}

	slog.Info("Created table: %s\n", query)
	return nil
}

func (em *ErrorMecha) WriteNumber(conn *pgx.Conn, ctx context.Context) error {
	newnum := em.bumpNumber()

	return em.insertNumber(conn, ctx, newnum)
}

func (em *ErrorMecha) insertNumber(conn *pgx.Conn, ctx context.Context, num int) error {
	query := `INSERT INTO mecha (number) VALUES ($1)`
	_, err := conn.Exec(ctx, query, num)
	if err != nil {
		slog.Error("Error inserting number: %v\n", err)
		return fmt.Errorf("could not insert number: %w", err)
	}

	slog.Info("Inserted number: %s\n", query)
	return nil
}

func (em *ErrorMecha) bumpNumber() int {
	// Increase the value in the struct
	em.Value = em.Value + rand.IntN(10)

	// Increase error mark monitor
	em.ErrorMark++

	// Return it for writing
	return em.Value
}
