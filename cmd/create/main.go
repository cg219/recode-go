package main

import (
	"log"
	"mentegee/recode/pkg/cmd/create"
	"gopkg.in/yaml.v3"
    _ "embed"
)

//go:embed config.yml
var config string

func main() {
    var cfg create.Config

    if err := yaml.Unmarshal([]byte(config), &cfg); err != nil {
        log.Fatal(err)
    }

    if err := create.Run(&cfg); err != nil {
        log.Fatal(err)
    }
}
