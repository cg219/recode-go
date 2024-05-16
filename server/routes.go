package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	rc "mentegee/recode/gui/recode"
	"mentegee/recode/gui/xerr"
	"net/http"
	"os"
	"strings"
	"sync"
)

func addRoutes(srv *server) {
    srv.mux.Handle("/f/", http.StripPrefix("/f/", http.FileServer(http.Dir("../src/static"))))
    srv.mux.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("../src/client/js"))))
    srv.mux.HandleFunc("/", handleIndex())
    srv.mux.HandleFunc("POST /newepisode", handleNewEpisode(*srv.queries, &srv.rootdir, srv.mtx))
    srv.mux.HandleFunc("GET /anime", handleAnime())
    srv.mux.HandleFunc("GET /queue", handleQueue(*srv.queries))
    srv.mux.HandleFunc("POST /rootdirectory", handleRootDir(srv.logger, *srv.queries, &srv.rootdir, srv.mtx))
}

func getQueue(query rc.Queries) <- chan Recode {
    ctx := context.Background()
    rows, err := query.GetQueue(ctx)
    xerr.LErr(err)

    out := make(chan Recode)

    go func() {
        for _, d := range(rows) {
            out <- Recode {Origin: d.Origin, Destination: d.Dest, Season: d.Season, Episode: d.Episode }
        }
        close(out)
    }()

    return out
}

func getRoot(query rc.Queries) <- chan string {
    ctx := context.Background()
    rootdir, err := query.GetPrefs(ctx)
    xerr.LErr(err)

    out := make(chan string)

    go func() {
        if rootdir.Valid {
            out <- rootdir.String
        }

        close(out)
    }()

    return out
}

func handleIndex() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        http.ServeFile(w, r, "../src/client/index.html")
    }
}

func handleNewEpisode(query rc.Queries, rootdir *string, mtx *sync.RWMutex) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        destination := r.PostFormValue("newepisode")
        episode := r.PostFormValue("episode")
        season := r.PostFormValue("season")
        _, video, err := r.FormFile("video")

        xerr.LErr(err)

        recode := Recode{ Season: season , Episode: episode }

        mtx.RLock()
        rootdirValue := *rootdir
        mtx.RUnlock()

        splitText := strings.Split(destination, "/")
        destName := splitText[len(splitText) - 1]
        dest := fmt.Sprintf("%v/%v - s%ve%v.mkv", destination, destName, recode.getSeason(), recode.getEpisode())
        origin := fmt.Sprintf("%v/%v", rootdirValue, video.Filename)

        ctx := context.Background()

        err = query.CreateRecode(ctx, rc.CreateRecodeParams{
            Season: season, 
            Episode: episode,
            Dest: dest,
            Origin: origin,
        })

        xerr.LErr(err)

        fmt.Fprintf(w, fmt.Sprintf("%v %v %v %v", destination, recode.getEpisode(), video.Filename, recode.getSeason()))

    }
}

func handleAnime() http.HandlerFunc {
    return func(w http.ResponseWriter, _ *http.Request) {
        dir := "/Volumes/media/Anime TV"
        entries, err := os.ReadDir(dir)
        xerr.LErr(err)

        list := make([]Anime, len(entries))

        for i, entry := range entries {
            if entry.IsDir() {
                list[i] = Anime { Name: entry.Name(), Path: dir + "/" + entry.Name() } 
            }
        }

        data, err := json.Marshal(list)
        xerr.PErr(err)
        w.Header().Set("Content-Type", "application/json")        
        w.Write(data)
    }
}

func handleQueue(query rc.Queries) http.HandlerFunc {
    return func(w http.ResponseWriter, _ *http.Request) {
        recodes := getQueue(query)
        list := []Recode{} 

        for recode := range recodes {
            list = append(list, recode)
        }

        data, err := json.Marshal(list)
        xerr.LErr(err)

        w.Header().Set("Content-Type", "application/json")
        w.Write(data)
    }
}

func handleRootDir(logger *log.Logger, query rc.Queries, rootdir *string, mtx *sync.RWMutex) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        dir := r.PostFormValue("rootdirectory")

        mtx.RLock()
        rootdirValue := *rootdir
        mtx.RUnlock()

        logger.Printf(dir)

        if rootdirValue == "" {
            in := getRoot(query) 

            mtx.Lock()
            *rootdir = <- in
            mtx.Unlock()
        }

        if dir != "" {
            ctx := context.Background()
            err := query.UpdatePref(ctx, rc.UpdatePrefParams{ Rootdir: sql.NullString{ String: dir, Valid: true } })

            xerr.LErr(err)

            mtx.Lock()
            *rootdir = dir
            mtx.Unlock()
        }

        mtx.RLock()
        logger.Println(*rootdir)

        w.Header().Set("Content-Type", "text/plain")
        w.Write([]byte(fmt.Sprintf("%v", &rootdir)))
        mtx.RUnlock()
    }
}
