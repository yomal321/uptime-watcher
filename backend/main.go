package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const schema = `
CREATE TABLE IF NOT EXISTS sites (
	id         SERIAL PRIMARY KEY,
	url        TEXT NOT NULL UNIQUE,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS checks (
	id          BIGSERIAL PRIMARY KEY,
	site_id     INT NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
	up          BOOLEAN NOT NULL,
	status_code INT,
	latency_ms  INT NOT NULL,
	checked_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS checks_site_time ON checks (site_id, checked_at DESC);`

type site struct {
	ID         int        `json:"id"`
	URL        string     `json:"url"`
	Up         *bool      `json:"up"`
	StatusCode *int       `json:"status_code"`
	LatencyMS  *int       `json:"latency_ms"`
	CheckedAt  *time.Time `json:"checked_at"`
}

var client = &http.Client{Timeout: 10 * time.Second}

func main() {
	// JSON logs so Azure Monitor / Log Analytics can query fields later.
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		fatal("DATABASE_URL is required")
	}
	ctx := context.Background()
	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		fatal("bad DATABASE_URL", "err", err)
	}
	if _, err := db.Exec(ctx, schema); err != nil {
		fatal("cannot reach database", "err", err)
	}

	go checkLoop(ctx, db)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		if err := db.Ping(r.Context()); err != nil {
			http.Error(w, "database unreachable", http.StatusServiceUnavailable)
			return
		}
		w.Write([]byte("ok"))
	})
	mux.HandleFunc("GET /sites", func(w http.ResponseWriter, r *http.Request) {
		rows, _ := db.Query(r.Context(), `
			SELECT s.id, s.url, c.up, c.status_code, c.latency_ms, c.checked_at
			FROM sites s
			LEFT JOIN LATERAL (
				SELECT up, status_code, latency_ms, checked_at FROM checks
				WHERE site_id = s.id ORDER BY checked_at DESC LIMIT 1
			) c ON true
			ORDER BY s.id`)
		sites, err := pgx.CollectRows(rows, pgx.RowToStructByPos[site])
		if err != nil {
			slog.Error("list sites", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, sites)
	})
	mux.HandleFunc("POST /sites", func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			URL string `json:"url"`
		}
		r.Body = http.MaxBytesReader(w, r.Body, 4096)
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil || !validURL(in.URL) {
			http.Error(w, `body must be {"url": "https://..."}`, http.StatusBadRequest)
			return
		}
		var id int
		err := db.QueryRow(r.Context(),
			`INSERT INTO sites (url) VALUES ($1)
			 ON CONFLICT (url) DO UPDATE SET url = EXCLUDED.url
			 RETURNING id`, in.URL).Scan(&id)
		if err != nil {
			slog.Error("add site", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		go check(context.Background(), db, id, in.URL) // first status right away, not in a minute
		writeJSON(w, http.StatusCreated, map[string]any{"id": id, "url": in.URL})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	slog.Info("listening", "port", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		fatal("server stopped", "err", err)
	}
}

func checkLoop(ctx context.Context, db *pgxpool.Pool) {
	for {
		rows, _ := db.Query(ctx, `SELECT id, url FROM sites`)
		targets, err := pgx.CollectRows(rows, pgx.RowToStructByPos[struct {
			ID  int
			URL string
		}])
		if err != nil {
			slog.Error("load sites", "err", err)
		}
		// ponytail: sequential checks, 10s timeout each; run them concurrently if the site list grows past ~5.
		for _, t := range targets {
			check(ctx, db, t.ID, t.URL)
		}
		time.Sleep(time.Minute)
	}
}

func check(ctx context.Context, db *pgxpool.Pool, id int, u string) {
	start := time.Now()
	up, code := false, (*int)(nil)
	resp, err := client.Get(u)
	if err == nil {
		resp.Body.Close()
		code = &resp.StatusCode
		up = resp.StatusCode < 400
	}
	latency := time.Since(start).Milliseconds()
	slog.Info("check", "url", u, "up", up, "status", code, "latency_ms", latency)
	if _, err := db.Exec(ctx,
		`INSERT INTO checks (site_id, up, status_code, latency_ms) VALUES ($1, $2, $3, $4)`,
		id, up, code, latency); err != nil {
		slog.Error("save check", "url", u, "err", err)
	}
}

func validURL(s string) bool {
	u, err := url.ParseRequestURI(s)
	return err == nil && (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func fatal(msg string, args ...any) {
	slog.Error(msg, args...)
	os.Exit(1)
}
