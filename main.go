package main

import (
	"context"
	"database/sql"
	_ "embed"
	"mentegee/recode/internal/cmd"
	"mentegee/recode/recode"
	"os"
	"time"

	//    "mentegee/recode/create"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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

    if err := runRecode(ddl, "db/recode.db"); err != nil {
        cmd.LogErr(err)
    }

    os.Exit(0)
}

type progressMsg float64
type recodeMsg bool
type errMsg struct{ err error }
type tickMsg time.Time

type recodeModel struct {
    progress progress.Model
    lastPercent float64
    percent float64
    channel chan float64
    titleStyle lipgloss.Style
    helpStyle lipgloss.Style
}

func (m recodeModel) Init() tea.Cmd {
    go encode(m)
    return tick()
}

func (m recodeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        k := msg.String()

        if k == "q" || k == "ctrl+c" {
            return m, tea.Quit
        }

        return m, nil

    case tea.WindowSizeMsg:
        m.progress.Width = msg.Width - 8 - 4

        if m.progress.Width > 120 {
            m.progress.Width = 120
        }

        return m, nil

    case tickMsg:
        p, ok := <-m.channel

        if ok {
            if p < m.lastPercent {
                return m, tea.Quit
            }

            m.lastPercent = m.percent
            m.percent = p

            if m.percent > 1.0 {
                m.percent = 1.0
                return m, tea.Quit
            }

            return m, tick()
        }

        return m, tea.Quit
    case errMsg:
        return m, tea.Quit
    case recodeMsg:
        return m, tea.Quit
    default:
        return m, nil
} 
}

func (m recodeModel) View() string {
    comp := lipgloss.JoinVertical(lipgloss.Left,
        m.titleStyle.Render("Recoding: test1.mp4"),
        m.progress.ViewAs(m.percent),
        m.helpStyle.Render("q/ctrl+c : quit"))

    return lipgloss.NewStyle().
        PaddingLeft(5).
        PaddingTop(2).
        PaddingBottom(2).
        Render(comp)
}

func newModel() recodeModel {
    helpStyle := lipgloss.NewStyle().
        Foreground(lipgloss.Color("#999999")).
        PaddingTop(2)
        
    titleStyle := lipgloss.NewStyle().
        Foreground(lipgloss.Color("#EFEFEF")).
        BorderBottomBackground(lipgloss.Color("#8C8C8C")).
        Bold(true).
        BorderBottom(true).
        PaddingBottom(2)
        
    return recodeModel{
        progress: progress.New(progress.WithScaledGradient("#0f8a7f", "#0979ad")),
        channel: make(chan float64, 1),
        helpStyle: helpStyle,
        titleStyle: titleStyle,
    }
} 

func tick() tea.Cmd {
    return tea.Tick(time.Second * 1, func(t time.Time) tea.Msg {
        return tickMsg(t)
    })
}

func encode(m recodeModel) {
    go func() {
        for p := range m.channel {
            m.Update(progressMsg(p))
        }
    }()

    err := recode.Encode("test/test1.mp4", "test/test2.mkv", m.channel)

    if err != nil {
        panic(err)
    }

    m.Update(recodeMsg(true))
}

func runRecode(schema string, dbpath string) error {
    ctx := context.Background()
    db, err := sql.Open("sqlite", dbpath)
    if err != nil {
        return err
    }
    defer db.Close()

    if _, err := db.ExecContext(ctx, schema); err != nil {
        return err
    }

    // progressChannel := make(chan float64)
    //
    // go func() {
    //     for p := range progressChannel {
    //         fmt.Println(p)
    //     }
    // }()
    //
    // if err := recode.Encode("test/test1.mp4", "test/test2.mkv", progressChannel); err != nil {
    //     return err
    // }

    if _, err:= tea.NewProgram(newModel()).Run(); err != nil {
        return err
    }

    return nil
}
