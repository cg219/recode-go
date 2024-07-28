// TODO:
// - Update Dest Tinput Styles
// - Update Filepicker to go up directories on step 2
package main

import (
	_ "embed"
	"mentegee/recode/cmd"
	"mentegee/recode/internal/helpers"
	"os"

	//    "mentegee/recode/create"

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

    c, _ := cmd.Init()

    if err := cmd.Run(c); err != nil {
        helpers.LogErr(err)
    }

    os.Exit(0)
}

