package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5"
)

func main() {
	// Provide defaults for required environment variables
	PGDB := envVar("PG_DB", "devdb")
	USER := envVar("PG_USER", "devuser")
	PASS := envVar("PG_PASS", "devpass")
	HOST := envVar("PG_HOST", "postgres.local-db.svc.cluster.local")
	PORT := envVar("PG_PORT", "5432")

	ctx := context.Background()

	// Create the connection string and connect
	cs := "postgres://" + USER + ":" + PASS + "@" + HOST + ":" + PORT + "/" + PGDB
	conn, err := pgx.Connect(ctx, cs)
	if err != nil {
		slog.Error("Could not connect to database", slog.Any("error", err))
		os.Exit(1)
	}
	if err = conn.Ping(ctx); err != nil {
		slog.Error("Could not ping database", slog.Any("error", err))
		os.Exit(1)
	}

	fmt.Println("Successfully connected to database, starting Error Mecha")

	// Create the app control interface
	em := NewErrorMecha()
	defer em.Ticker.Stop()

	// Create the table if needed
	if err = em.createTable(conn, ctx); err != nil {
		slog.Warn("Warning: could not create table", slog.Any("error", err))
	}

	// Run webserver in parallel
	go func() {
		fmt.Println("Starting Error Mecha Metrics Server")
		em.Server = &http.Server{Addr: ":8080", Handler: em.SetupMux()}
		if err = em.Server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	// Blocking main loop that writes numbers to the database
	for {
		select {
		case <-em.Ticker.C:
			// Catch a real error when writing
			if err = em.WriteNumber(conn, ctx); err != nil {
				slog.Error("Error writing number", slog.Any("error", err))
				em.Stats.RecErrorCounter()
				continue
			}

			// Every 9 ticks, create a purposeful error
			if em.ErrorMark%9 == 0 {
				slog.Error("MECHA stole the write!")
				em.Stats.RecErrorCounter()
				continue
			}

			// If we've gotten this far, it's a good write
			em.Stats.RecOkCounter()
		}
	}
}

// envVar grabs a single ENV VAR with a provided default
func envVar(env, alt string) string {
	value, ext := os.LookupEnv(env)
	if !ext {
		return alt
	}
	return value
}
