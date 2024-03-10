package main

import (
    "database/sql"
    "fmt"
    "mentegee/recode/gui/xerr"
    "net/http"
    "os"
    "strings"
    "text/template"

    _ "modernc.org/sqlite"
)

type Anime struct{
    Name string
    Path string
}

type Recode struct {
    Origin string
    Destination string
    Season string 
    Episode string
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
    rows, err := db.Query("SELECT rootdir FRON prefs")
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
    animeTmpl := `{{range .}}<option value="{{.Path}}">{{.Name}}</option>{{end}}` 
    queueTmpl := `{{range .}}<li>{{.Episode}} - {{.Season}}</li>{{end}}` 
    rootTmpl := `<input id="rootdirectory" name="rootdirectory" type="text" value="{{.}}" placeholder="Root Directory" hx-trigger="keyup changed delay:500ms" hx-post="/rootdirectory" hx-swap="outerHTML" hx-target="this" />`
    var rootdir string

    mux.Handle("/f/", http.StripPrefix("/f/", files))

    mux.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
        fmt.Println("GET /")
        http.ServeFile(res, req, "../src/client/index.html")
    })

    mux.HandleFunc("POST /newepisode", func(res http.ResponseWriter, req *http.Request) {
        destination := req.PostFormValue("showPath")
        episode := req.PostFormValue("episode")
        season := req.PostFormValue("season")
        _, video, err := req.FormFile("video")

        xerr.LErr(err)

        recode := Recode{ Season: season , Episode: episode }

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

        tmpl, err := template.New("animelist").Parse(animeTmpl)
        err = tmpl.Execute(res, list)
        xerr.LErr(err)
    })

    mux.HandleFunc("GET /queue", func(res http.ResponseWriter, req *http.Request) {
        recodes := getQueue(db)
        list := []Recode{} 

        for recode := range recodes {
            list = append(list, recode)
        }

        tmpl, err := template.New("queue").Parse(queueTmpl)
        xerr.LErr(err)
        xerr.LErr(tmpl.Execute(res, list))

    })

    mux.HandleFunc("POST /rootdirectory", func(res http.ResponseWriter, req *http.Request) {
        dir := req.PostFormValue("rootdirectory")

        fmt.Printf(dir)

        if rootdir == "" {
            in := getRoot(db) 

            rootdir = <- in
        }

        if dir != "" {
            query, err := db.Prepare("insert into prefs (id, rootdir) values (?, ?) on conflict (id) do update set rootdir = excluded.rootdir")
            xerr.LErr(err)

            _, err = query.Exec(1, dir)
            xerr.LErr(err)

            rootdir = dir
        }

        fmt.Println(rootdir)
        tmpl, err := template.New("rootdir").Parse(rootTmpl)
        xerr.LErr(err)

        err = tmpl.Execute(res, rootdir)
        xerr.LErr(err)
    })

    http.ListenAndServe(":3000", mux)
}
