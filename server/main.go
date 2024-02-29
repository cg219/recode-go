package main

import (
	"database/sql"
	"fmt"
	"log"
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

func check(err error) {
    if err != nil {
        log.Fatal(err)
    }
}

func getQueue(db *sql.DB) <- chan Recode {
    rows, err := db.Query("select origin, dest, season, episode from recodes")
    check(err)

    out := make(chan Recode)
    
    go func() {
        for rows.Next() {
            var origin string
            var dest string
            var season string
            var episode string

            err := rows.Scan(&origin, &dest, &season, &episode)
            check(err)      

            out <- Recode {Origin: origin, Destination: dest, Season: season, Episode: episode }
        }

        rows.Close()
        close(out)
    }()

    return out
}

func getRoot(db *sql.DB) <- chan string {
    rows, err := db.Query("select rootdir from prefs")
    check(err)

    out := make(chan string)
    
    go func() {
        for rows.Next() {
            var rootdir string

            err := rows.Scan(&rootdir)
            check(err)      

            out <- rootdir
        }

        rows.Close()
        close(out)
    }()

    return out
}

func main () {
    db, err := sql.Open("sqlite", "recode.db")
    check(err)
    defer db.Close()

    sql := `create table if not exists recodes (
        id INT PRIMARY KEY,
        origin TEXT NOT NULL,
        dest TEXT NOT NULL,
        season TEXT NOT NULL,
        episode TEXT NOT NULL
    );

    create table if not exists prefs (
        id INT PRIMARY KEY CHECK (id = 1),
        rootdir TEXT
    )`

    _, err = db.Exec(sql)
    check(err)

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
        
        check(err)

        recode := Recode{ Season: season , Episode: episode }

        fmt.Fprintf(res, fmt.Sprintf("%v %v %v %v", destination, recode.getEpisode(), video.Filename, recode.getSeason()))
    })

    mux.HandleFunc("GET /anime", func(res http.ResponseWriter, req *http.Request) {
        dir := "/Volumes/media/Anime TV"
        entries, err := os.ReadDir(dir)
        check(err)

        list := make([]Anime, len(entries))

        for i, entry := range entries {
            if entry.IsDir() {
                list[i] = Anime { Name: entry.Name(), Path: dir + "/" + entry.Name() } 
            }
        }

        tmpl, err := template.New("animelist").Parse(animeTmpl)
        err = tmpl.Execute(res, list)
        check(err)
    })

    mux.HandleFunc("GET /queue", func(res http.ResponseWriter, req *http.Request) {
        recodes := getQueue(db)
        list := []Recode{} 

        for recode := range recodes {
            list = append(list, recode)
        }

            tmpl, err := template.New("queue").Parse(queueTmpl)
            check(err)
            check(tmpl.Execute(res, list))

    })

    mux.HandleFunc("POST /rootdirectory", func(res http.ResponseWriter, req *http.Request) {
        dir := req.PostFormValue("rootdirectory")

        if rootdir == "" || dir == "" {
            in := getRoot(db) 

            rootdir = <- in
        } else {
            query, err := db.Prepare("insert into prefs (id, rootdir) values (?, ?) on conflict (id) do update set rootdir = excluded.rootdir")
            check(err)

            _, err = query.Exec(1, dir)
            check(err)

            rootdir = dir
        }

        tmpl, err := template.New("rootdir").Parse(rootTmpl)
        check(err)

        err = tmpl.Execute(res, rootdir)
        check(err)
    })

    http.ListenAndServe(":3000", mux)
}
