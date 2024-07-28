package cmd

import (
	"fmt"
	"mentegee/recode/internal/helpers"
	"mentegee/recode/recode"
	"os"

	"github.com/spf13/cobra"
)

type RootCmd struct {
    src *string
    dest *string
}

func Init() (*cobra.Command, *RootCmd) {
    src := ""
    dest := ""
    cv := &RootCmd{
        src: &src,
        dest: &dest,
    }

    c :=  &cobra.Command {
        Use: "recode",
        Short: "Re-encode video to a lower bitrate (H.265)",
        Long: "Re-encode video to a lower bitrate (H.265)",
        Run: func(c *cobra.Command, args []string) {
            jp := false

            if src != "" {
                jp = true
            }

            if err := recode.Tui(jp, src, dest); err != nil {
                helpers.LogErr(err)
                os.Exit(1)
            }

            os.Exit(0)
        },
    }

    c.Flags().StringVarP(cv.src, "file", "f", "",  "File to encode")
    c.Flags().StringVarP(cv.dest, "output", "o", "", "Filename to save encoded file")

    return c, cv
}

func Run(c *cobra.Command) error {
    if err := c.Execute(); err != nil {
        fmt.Println(err)
        return err
    }

    return nil
}

