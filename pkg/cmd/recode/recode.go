package recode

import (
	"context"
	"database/sql"
	_ "embed"
	"log"
	"mentegee/recode/pkg/mq"
	"net/http"
	"strings"
	"sync"

	_ "modernc.org/sqlite"
)

func (r *Recode) getEpisode() string {
    count  := 1 + ((3-len(r.Episode)) / len("0"))
    str := strings.Repeat("0", count) + r.Episode

    return str[len(str) - 3:]
}

func (r *Recode) getSeason() string {
    count  := 1 + ((2-len(r.Season)) / len("0"))
    str := strings.Repeat("0", count) + r.Season

    return str[len(str) - 2:]
}

func NewServer(db *sql.DB) *server {
    q := mq.New(db)

    return &server{
        queries: q,
        mux: http.NewServeMux(),
        logger: log.Default(),
        mtx: &sync.RWMutex{},
    }
}

func Run(schema string, dbpath string) error {
    ctx := context.Background()
    db, err := sql.Open("sqlite", dbpath)
    if err != nil {
        return err
    }
    defer db.Close()

    if _, err := db.ExecContext(ctx, schema); err != nil {
        return err
    }

    srv := NewServer(db)

    addRoutes(srv)

    http.ListenAndServe(":3000", srv.mux)
    return nil
}
