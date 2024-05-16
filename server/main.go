package main

import (
	"context"
	"database/sql"
	"log"
	"mentegee/recode/gui/xerr"
	"net/http"
	"strings"
	"sync"
    _ "embed"

	rc "mentegee/recode/gui/recode"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var ddl string

type Anime struct{
    Name string `json:"name"`
    Path string `json:"path"`
}

type Recode struct {
    Origin string `json:"origin"`
    Destination string `json:"destination"`
    Season string `json:"season"`
    Episode string `json:"episode"`
}

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

type server struct {
    queries *rc.Queries
    mux *http.ServeMux
    rootdir string
    logger *log.Logger
    mtx *sync.RWMutex
}

func NewServer(db *sql.DB) *server {
    q := rc.New(db)

    return &server{
        queries: q,
        mux: http.NewServeMux(),
        logger: log.Default(),
        mtx: &sync.RWMutex{},
    }
}


func run() error {
    ctx := context.Background()
    db, err := sql.Open("sqlite", "recode.db")
    if err != nil {
        return err
    }
    defer db.Close()

    if _, err := db.ExecContext(ctx, ddl); err != nil {
        return err
    }

    srv := NewServer(db)

    addRoutes(srv)
    
    http.ListenAndServe(":3000", srv.mux)
    return nil
}

func main () {
    if err := run(); err != nil {
        xerr.LErr(err)
    }
}
