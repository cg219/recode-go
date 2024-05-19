package recode

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"mentegee/recode/pkg/cmd"
	"mentegee/recode/pkg/mq"
	"net/http"
	"os"
	"strings"
	"sync"
)

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

type server struct {
    queries *mq.Queries
    mux *http.ServeMux
    rootdir string
    logger *log.Logger
    mtx *sync.RWMutex
}

func addRoutes(srv *server) {
    srv.mux.Handle("/f/", http.StripPrefix("/f/", http.FileServer(http.Dir("../../web/static"))))
    srv.mux.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("../../web/ui/js"))))
    srv.mux.HandleFunc("/", handleIndex())
    srv.mux.HandleFunc("POST /newepisode", handleNewEpisode(*srv.queries, &srv.rootdir, srv.mtx))
    srv.mux.HandleFunc("GET /anime", handleAnime())
    srv.mux.HandleFunc("GET /queue", handleQueue(*srv.queries))
    srv.mux.HandleFunc("POST /rootdirectory", handleRootDir(srv.logger, *srv.queries, &srv.rootdir, srv.mtx))
}

func getQueue(query mq.Queries) <- chan Recode {
    ctx := context.Background()
    rows, err := query.GetQueue(ctx)
    cmd.LogErr(err)

    out := make(chan Recode)

    go func() {
        for _, d := range(rows) {
            out <- Recode {Origin: d.Origin, Destination: d.Dest, Season: d.Season, Episode: d.Episode }
        }
        close(out)
    }()

    return out
}

func getRoot(query mq.Queries) <- chan string {
    ctx := context.Background()
    rootdir, err := query.GetPrefs(ctx)
    cmd.LogErr(err)

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
        http.ServeFile(w, r, "../../web/ui/index.html")
    }
}

func handleNewEpisode(query mq.Queries, rootdir *string, mtx *sync.RWMutex) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        destination := r.PostFormValue("newepisode")
        episode := r.PostFormValue("episode")
        season := r.PostFormValue("season")
        _, video, err := r.FormFile("video")

        cmd.LogErr(err)

        recode := Recode{ Season: season , Episode: episode }

        mtx.RLock()
        rootdirValue := *rootdir
        mtx.RUnlock()

        splitText := strings.Split(destination, "/")
        destName := splitText[len(splitText) - 1]
        dest := fmt.Sprintf("%v/%v - s%ve%v.mkv", destination, destName, recode.getSeason(), recode.getEpisode())
        origin := fmt.Sprintf("%v/%v", rootdirValue, video.Filename)

        ctx := context.Background()

        err = query.CreateRecode(ctx, mq.CreateRecodeParams{
            Season: season, 
            Episode: episode,
            Dest: dest,
            Origin: origin,
        })

        cmd.LogErr(err)

        fmt.Fprintf(w, fmt.Sprintf("%v %v %v %v", destination, recode.getEpisode(), video.Filename, recode.getSeason()), nil)

    }
}

func handleAnime() http.HandlerFunc {
    return func(w http.ResponseWriter, _ *http.Request) {
        dir := "/Volumes/media/Anime TV"
        entries, err := os.ReadDir(dir)
        cmd.LogErr(err)

        list := make([]Anime, len(entries))

        for i, entry := range entries {
            if entry.IsDir() {
                list[i] = Anime { Name: entry.Name(), Path: dir + "/" + entry.Name() } 
            }
        }

        data, err := json.Marshal(list)
        cmd.PrintErr(err)
        w.Header().Set("Content-Type", "application/json")        
        w.Write(data)
    }
}

func handleQueue(query mq.Queries) http.HandlerFunc {
    return func(w http.ResponseWriter, _ *http.Request) {
        recodes := getQueue(query)
        list := []Recode{} 

        for recode := range recodes {
            list = append(list, recode)
        }

        data, err := json.Marshal(list)
        cmd.LogErr(err)

        w.Header().Set("Content-Type", "application/json")
        w.Write(data)
    }
}

func handleRootDir(logger *log.Logger, query mq.Queries, rootdir *string, mtx *sync.RWMutex) http.HandlerFunc {
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
            err := query.UpdatePref(ctx, mq.UpdatePrefParams{ Rootdir: sql.NullString{ String: dir, Valid: true } })

            cmd.LogErr(err)

            mtx.Lock()
            *rootdir = dir
            mtx.Unlock()
        }

        mtx.RLock()
        logger.Println(*rootdir)

        w.Header().Set("Content-Type", "text/plain")
        w.Write([]byte(fmt.Sprintf("%v", *rootdir)))
        mtx.RUnlock()
    }
}
