package main

import (
	_ "embed"
	"mentegee/recode/pkg/cmd"
	"mentegee/recode/pkg/cmd/recode"
)

//go:embed schema.sql
var ddl string

func main () {
    if err := recode.Run(ddl, "../../db/recode.db"); err != nil {
        cmd.LogErr(err)
    }
}
