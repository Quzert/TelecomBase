package store

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	loadOnce sync.Once
	loadErr  error
	byName   map[string]string
)

type DBTX interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type Queries struct {
	db DBTX
}

func New(db *pgxpool.Pool) *Queries {
	return NewDB(db)
}

func NewDB(db DBTX) *Queries {
	loadOnce.Do(func() {
		dir, err := resolveQueriesDir()
		if err != nil {
			loadErr = err
			return
		}
		byName, loadErr = parseNamedQueriesFromDir(dir)
	})
	if loadErr != nil {
		panic(loadErr)
	}
	return &Queries{db: db}
}

func resolveQueriesDir() (string, error) {
	if v := strings.TrimSpace(os.Getenv("TELECOMBASE_SQL_DIR")); v != "" {
		return v, nil
	}

	// Common dev case: run from module root (server/).
	if dirExists("db/queries") {
		return "db/queries", nil
	}

	// If someone runs from repo root.
	if dirExists("server/db/queries") {
		return "server/db/queries", nil
	}

	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("resolve executable path: %w", err)
	}
	exeDir := filepath.Dir(exe)
	// Typical container layout: /app/telecombase-api + /app/db/queries
	if dirExists(filepath.Join(exeDir, "db/queries")) {
		return filepath.Join(exeDir, "db/queries"), nil
	}
	// If binary is in a subdir.
	if dirExists(filepath.Join(exeDir, "../db/queries")) {
		return filepath.Clean(filepath.Join(exeDir, "../db/queries")), nil
	}

	return "", fmt.Errorf("SQL queries directory not found (tried db/queries, server/db/queries, %s)", path.Join(exeDir, "db/queries"))
}

func dirExists(p string) bool {
	st, err := os.Stat(p)
	return err == nil && st.IsDir()
}

func sql(name string) string {
	q, ok := byName[name]
	if !ok {
		panic(fmt.Sprintf("missing SQL query: %s", name))
	}
	return q
}

func parseNamedQueriesFromDir(dir string) (map[string]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("readdir %s: %w", dir, err)
	}

	nameLine := regexp.MustCompile(`^--\s*name:\s*([A-Za-z0-9_]+)\s*:`)

	result := make(map[string]string)
	for _, ent := range entries {
		if ent.IsDir() {
			continue
		}
		if !strings.HasSuffix(ent.Name(), ".sql") {
			continue
		}

		path := filepath.Join(dir, ent.Name())
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", path, err)
		}

		scanner := bufio.NewScanner(strings.NewReader(string(content)))

		var currentName string
		var buf strings.Builder
		flush := func() {
			if currentName == "" {
				return
			}
			s := strings.TrimSpace(buf.String())
			if s == "" {
				return
			}
			result[currentName] = s
			currentName = ""
			buf.Reset()
		}

		for scanner.Scan() {
			line := scanner.Text()
			if m := nameLine.FindStringSubmatch(line); m != nil {
				flush()
				currentName = m[1]
				continue
			}
			if currentName == "" {
				continue
			}
			buf.WriteString(line)
			buf.WriteString("\n")
		}
		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("scan %s: %w", path, err)
		}
		flush()
	}

	// Fail fast if dir exists but no named queries were parsed.
	if len(entries) > 0 && len(result) == 0 {
		return nil, fmt.Errorf("no named queries found in %s (expected '-- name: <QueryName> :...')", dir)
	}

	return result, nil
}
