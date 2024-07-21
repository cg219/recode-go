package main

import (
	"context"
	"database/sql"
	_ "embed"
	"mentegee/recode/internal/cmd"
	"mentegee/recode/recode"
	"os"

	//    "mentegee/recode/create"
	"log"

	_ "modernc.org/sqlite"
	// "gopkg.in/yaml.v3"
)

//go:embed configs/schema.sql
var ddl string

// // go:embed configs/config.yml
// var config string

func main () {
    // var cfg create.Config
    //
    // if err := yaml.Unmarshal([]byte(config), &cfg); err != nil {
    //     log.Fatal(err)
    // }

    if err := runRecode(ddl, "db/recode.db"); err != nil {
        cmd.LogErr(err)
    }

    os.Exit(0)
}

func runRecode(schema string, dbpath string) error {
    ctx := context.Background()
    db, err := sql.Open("sqlite", dbpath)
    if err != nil {
        return err
    }
    defer db.Close()

    if _, err := db.ExecContext(ctx, schema); err != nil {
        return err
    }

    err = recode.Encode("test/test1.mp4", "test/test2.mkv")
    
    if err != nil {
        return err
    }

    return nil
}
