package recode

import (
    "context"
    "os"

    "github.com/xfrr/goffmpeg/v2/ffmpeg"
    "github.com/xfrr/goffmpeg/v2/ffmpeg/progress"
    "github.com/xfrr/goffmpeg/v2/ffprobe"
    "github.com/xfrr/goffmpeg/v2/pkg/media"
)

type passdata struct {
    firstPass *ffmpeg.Command
    secondPass *ffmpeg.Command
    firstPassChannel chan bool
    secondPassChannel chan bool
    progressChannel chan float64
    doneChannel chan struct{}
    fileInfo media.File
}

func Encode(in string, out string, progress chan float64) error {
    _, err := os.Create("/dev/null")

    if err != nil {
        return err
    }

    ch1 := make(chan bool)
    ch2 := make(chan bool)

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

    pd, err := newPassData(c1, c2, ch1, ch2, in, progress)

    if err != nil {
        return err
    }

    go pass(pd, nil)
    go pass(pd, pd.secondPass)

    ch1 <- true

    <-pd.doneChannel

    os.Remove("fil.log")
    return nil
}

func newPassData(firstPass, secondPass *ffmpeg.Command, ch1, ch2 chan bool, inputPath string, p chan float64) (*passdata, error) {
    d := make(chan struct{})

    ctx := context.Background()
    ctx, cancel := context.WithCancel(ctx)
    defer cancel()

    f, err := ffprobe.NewCommand().WithInputPath(inputPath).Run(ctx)

    if err != nil {
        return nil, err
    }

    return &passdata{
        firstPass: firstPass,
        secondPass: secondPass,
        firstPassChannel: ch1,
        secondPassChannel: ch2,
        progressChannel: p,
        doneChannel: d,
        fileInfo: f,
    }, nil
}

func pass (d *passdata, command *ffmpeg.Command) {
    ctx := context.Background()
    ctx, cancel := context.WithCancel(ctx)
    defer cancel()

    isSecondPass := false

    if command != nil {
        isSecondPass = true
    }

    pc := make(chan struct{}, 1)

    go func() {
        var p <-chan progress.Progress
        var err error

        if !isSecondPass {
            <-d.firstPassChannel
            p, err = d.firstPass.Start(ctx)
        } else {
            <-d.secondPassChannel
            p, err = d.secondPass.Start(ctx)
        }


        if err != nil {
            panic(err)
        }

        go func() {
            for {
                select {
                case <-pc:
                    if !isSecondPass {
                        close(d.firstPassChannel)
                    } else {
                        close(d.secondPassChannel)
                    }

                    close(pc)
                    return
                default:
                    for msg := range p {
                        n := float64(msg.Duration().Milliseconds()) / (float64(d.fileInfo.Duration().Milliseconds()) * 2) 

                        if !isSecondPass {
                            d.progressChannel <-n
                        } else {
                            d.progressChannel <-n + 0.50
                        }

                    }
            }
            }
        }()

    }()

    var err error

    if !isSecondPass {
        err = d.firstPass.Wait()
    } else {
        err = d.secondPass.Wait()
    }

    if err != nil {
        panic(err)
    }

    pc <- struct{}{}

    if !isSecondPass {
        d.secondPassChannel  <- true
    } else {
        d.doneChannel <- struct{}{}
        close(d.doneChannel)
        close(d.progressChannel)
    }
}
