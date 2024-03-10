package xerr

import "log"

func PErr(err error) {
    if err != nil {
        panic(err)
    }
}

func LErr(err error) {
    if err != nil {
        log.Fatal(err)
    }
}
