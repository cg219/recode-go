package recode

import (
	"context"
	"log"
	"os"
	"sync"

	"github.com/xfrr/goffmpeg/v2/ffmpeg"
)

func Encode(in string, out string) error {
    _, err := os.Create("/dev/null")

    if err != nil {
        return err
    }

    var wg sync.WaitGroup

    ch1 := make(chan bool)
    ch2 := make(chan bool)
    defer close(ch1)
    defer close(ch2)

    c1 := ffmpeg.NewCommand().
        WithInputPath(in).
        WithPass(1).
        WithPassLogFile("fil.log").
        WithOutputFormat("null").
        WithVideoCodec("libx265").
        WithVideoBitrate("1200k").
        WithThreadAmount(16).
        WithOutputPath("/dev/null")

    c2 := ffmpeg.NewCommand().
        WithInputPath(in).
        WithPass(2).
        WithPassLogFile("fil.log").
        WithMap("0").
        WithVideoCodec("libx265").
        WithVideoBitrate("1200k").
        WithAudioCodec("aac").
        WithAudioBitrate("128K").
        WithSubtitleCodec("copy").
        WithThreadAmount(16).
        WithOutputPath(out)

    go pass(c1, ch1, ch2, &wg)
    go pass(c2, ch2, nil, &wg)

    ch1 <- true

    wg.Wait()

    os.Remove("fil.log")

    return nil
}

func pass (c *ffmpeg.Command, start <-chan bool, stop chan<- bool, wg *sync.WaitGroup) {
    wg.Add(1)

    log.Println("Added")
    ctx := context.Background()
    ctx, cancel1 := context.WithCancel(ctx)
    defer cancel1()

    go func() {
        log.Println("Starting")
        log.Println(c)

        <-start

        p, err := c.Start(ctx)

        if err != nil {
            panic(err)
        }

        go func() {
            for msg := range p {
                log.Printf("%2.f", msg.Duration().Seconds())
            }
        }()

    }()

    err := c.Wait()

    if err != nil {
        panic(err)
    }

    wg.Done()

    if stop != nil {
        stop <-true
    }

    log.Println("Done")
}
