package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"mentegee/recode/gui/xerr"
	"net/http"
	"os"
	"strings"

	_ "modernc.org/sqlite"
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

func getQueue(db *sql.DB) <- chan Recode {
    rows, err := db.Query("SELECT origin, dest, season, episode FROM recodes WHERE processed = 0")
    xerr.LErr(err)

    out := make(chan Recode)

    go func() {
        for rows.Next() {
            var origin string
            var dest string
            var season string
            var episode string

            err := rows.Scan(&origin, &dest, &season, &episode)
            xerr.LErr(err)

            out <- Recode {Origin: origin, Destination: dest, Season: season, Episode: episode }
        }

        rows.Close()
        close(out)
    }()

    return out
}

func getRoot(db *sql.DB) <- chan string {
    rows, err := db.Query("SELECT rootdir FROM prefs")
    xerr.LErr(err)

    out := make(chan string)

    go func() {
        for rows.Next() {
            var rootdir string

            err := rows.Scan(&rootdir)
            xerr.LErr(err)

            out <- rootdir
        }

        rows.Close()
        close(out)
    }()

    return out
}

func main () {
    db, err := sql.Open("sqlite", "recode.db")
    xerr.LErr(err)
    defer db.Close()

    sql := `CREATE TABLE IF NOT EXISTS recodes (
    id INT PRIMARY KEY,
    origin TEXT NOT NULL,
    dest TEXT NOT NULL,
    season TEXT NOT NULL,
    episode TEXT NOT NULL,
    processed BOOLEAN NOT NULL DEFAULT(0),
    createdAt INTEGER NOT NULL DEFAULT(unixepoch(CURRENT_TIMESTAMP)),
    updatedAt INTEGER NOT NULL DEFAULT(unixepoch(CURRENT_TIMESTAMP))
    );

    CREATE TABLE IF NOT EXISTS prefs (
    id INT PRIMARY KEY CHECK (id = 1),
    rootdir TEXT
    )`

    _, err = db.Exec(sql)
    xerr.LErr(err)

    mux := http.NewServeMux()
    files := http.FileServer(http.Dir("../src/static"))
    jsfiles := http.FileServer(http.Dir("../src/client/js"))
    var rootdir string

    mux.Handle("/f/", http.StripPrefix("/f/", files))
    mux.Handle("/js/", http.StripPrefix("/js/", jsfiles))

    mux.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
        fmt.Println("GET /")
        http.ServeFile(res, req, "../src/client/index.html")
    })

    mux.HandleFunc("POST /newepisode", func(res http.ResponseWriter, req *http.Request) {
        destination := req.PostFormValue("newepisode")
        episode := req.PostFormValue("episode")
        season := req.PostFormValue("season")
        _, video, err := req.FormFile("video")

        xerr.LErr(err)

        recode := Recode{ Season: season , Episode: episode }

        query, err := db.Prepare("INSERT INTO recodes (season, episode, dest, origin) VALUES (?, ?, ?, ?)")
        xerr.LErr(err)

        splitText := strings.Split(destination, "/")
        destName := splitText[len(splitText) - 1]
        dest := fmt.Sprintf("%v/%v - s%ve%v.mkv", destination, destName, recode.getSeason(), recode.getEpisode())
        origin := fmt.Sprintf("%v/%v", rootdir, video.Filename)
        _, err = query.Exec(season, episode, dest, origin)
        xerr.LErr(err)

        fmt.Fprintf(res, fmt.Sprintf("%v %v %v %v", destination, recode.getEpisode(), video.Filename, recode.getSeason()))
    })

    mux.HandleFunc("GET /anime", func(res http.ResponseWriter, req *http.Request) {
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
        res.Header().Set("Content-Type", "application/json")        
        res.Write(data)
    })

    mux.HandleFunc("GET /queue", func(res http.ResponseWriter, req *http.Request) {
        recodes := getQueue(db)
        list := []Recode{} 

        for recode := range recodes {
            list = append(list, recode)
        }

        data, err := json.Marshal(list)
        xerr.LErr(err)

        res.Header().Set("Content-Type", "application/json")
        res.Write(data)
    })

    mux.HandleFunc("POST /rootdirectory", func(res http.ResponseWriter, req *http.Request) {
        dir := req.PostFormValue("rootdirectory")

        fmt.Printf(dir)

        if rootdir == "" {
            in := getRoot(db) 

            rootdir = <- in
        }

        if dir != "" {
            query, err := db.Prepare("INSERT INTO prefs (id, rootdir) VALUES (?, ?) ON CONFLICT (id) DO UPDATE SET rootdir = excluded.rootdir")
            xerr.LErr(err)

            _, err = query.Exec(1, dir)
            xerr.LErr(err)

            rootdir = dir
        }

        fmt.Println(rootdir)

        res.Header().Set("Content-Type", "text/plain")
        res.Write([]byte(fmt.Sprintf("%v", rootdir)))
    })

    http.ListenAndServe(":3000", mux)
}
