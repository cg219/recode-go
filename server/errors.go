package main

import "log"

func printErr(err error) {
    if err != nil {
        panic(err)
    }
}

func logErr(err error) {
    if err != nil {
        log.Fatal(err)
    }
}
