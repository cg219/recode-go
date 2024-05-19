package cmd

import "log"

func PrintErr(err error) {
    if err != nil {
        panic(err)
    }
}

func LogErr(err error) {
    if err != nil {
        log.Fatal(err)
    }
}
