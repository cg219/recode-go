package main

import (
	_ "embed"
	"mentegee/recode/internal/cmd"
	"mentegee/recode/recode"
    "mentegee/recode/create"
	"log"
	"gopkg.in/yaml.v3"
)

//go:embed configs/schema.sql
var ddl string

//go:embed configs/config.yml
var config string

func main () {
    var cfg create.Config

    if err := yaml.Unmarshal([]byte(config), &cfg); err != nil {
        log.Fatal(err)
    }

    if err := recode.Run(ddl, "db/recode.db"); err != nil {
        cmd.LogErr(err)
    }
}
