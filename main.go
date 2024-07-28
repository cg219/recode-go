// TODO:
// - Update Dest Tinput Styles
// - Update Filepicker to go up directories on step 2
package main

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"mentegee/recode/internal/cmd"
	"mentegee/recode/recode"
	"os"
	"time"

	//    "mentegee/recode/create"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	_ "modernc.org/sqlite"
	// "gopkg.in/yaml.v3"
)

//go:embed configs/schema.sql
var ddl string

// // go:embed configs/config.yml
// var config string

const (
    chooseFile state = 1
    chooseDest state = 2
    recodeStep state = 3
)

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
type state int

type recodeModel struct {
    progress progress.Model
    filepicker filepicker.Model
    lastPercent float64
    percent float64
    channel chan float64
    titleStyle lipgloss.Style
    helpStyle lipgloss.Style
    state state
    srcFile string
    destFile string
    destFileInput textinput.Model
}

func (m recodeModel) Init() tea.Cmd {
    if m.state == chooseFile {
        return m.filepicker.Init()
    }

    return nil
}

func (m recodeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        k := msg.String()

        if k == "q" && !m.destFileInput.Focused() {
            return m, tea.Quit
        }

        if k == "ctrl+c" {
            return m, tea.Quit
        }

        if k == "enter" && m.destFileInput.Focused() && m.state == chooseDest {
            m.state = recodeStep
            return m, startEncode(m)
        }

    case tea.WindowSizeMsg:
        if m.state == chooseFile {
            m.filepicker.Height = 10

            return m, nil
        }

        if m.state == chooseDest {
            m.filepicker.Height = 10
            m.destFileInput.Width = msg.Width - 12

            return m, nil
        }

        m.progress.Width = msg.Width - 12

        if m.progress.Width > 120 {
            m.progress.Width = 120
        }

        return m, nil

    case tickMsg:
        if m.state == recodeStep {
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
        }

    case errMsg:
        return m, tea.Quit
    case recodeMsg:
        return m, tea.Quit
} 
    var cmd tea.Cmd

    if m.state == chooseFile {
        m.filepicker, cmd = m.filepicker.Update(msg)

        if didSelect, path := m.filepicker.DidSelectFile(msg); didSelect {
            m.srcFile = path
            m.destFile = path
            m.destFileInput.SetValue(path)
            m.state = chooseDest
            return m, m.destFileInput.Focus()
        }
    }

    if m.state == chooseDest {
        if m.destFileInput.Focused() {
            m.destFileInput, cmd = m.destFileInput.Update(msg)
            m.destFile = m.destFileInput.Value()
        }
    }

    return m, cmd
}

func (m recodeModel) View() string {
    comp := lipgloss.JoinVertical(lipgloss.Left,
        m.titleStyle.Render(fmt.Sprintf("Recoding: %s to %s", m.srcFile, m.destFile)),
        m.progress.ViewAs(m.percent),
        m.helpStyle.Render("q/ctrl+c : quit"))

    switch m.state {
    case chooseFile:
        return m.filepicker.View()

    case chooseDest:
        return lipgloss.JoinVertical(lipgloss.Left,
            m.filepicker.View(),
            m.destFileInput.View())

    default:
        return lipgloss.NewStyle().
            PaddingLeft(5).
            PaddingTop(2).
            PaddingBottom(2).
            Render(comp)
}

}

func setSrcFp(fp *filepicker.Model) {
    fp.AllowedTypes = []string{".mp4", ".mkv", ".mov", ".mpg", ".mpeg"}
    fp.DirAllowed = false
    fp.FileAllowed = true
}

func setDestFp(fp *filepicker.Model) {
    fp.AllowedTypes = []string{"*"}
    fp.DirAllowed = true
    fp.FileAllowed = false
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

    destStyle := lipgloss.NewStyle().
        Background(lipgloss.Color("#555555")).
        Border(lipgloss.BlockBorder(), true).
        Foreground(lipgloss.Color("#EFEFEF")).
        Padding(2)
    
    dir, err := os.Getwd()

    if err != nil {
        panic(err)
    }

    fp := filepicker.New()
    fp.CurrentDirectory = dir
    setSrcFp(&fp)

    ti := textinput.New()
    ti.Placeholder = "Output Filename"
    ti.TextStyle = destStyle
        
    return recodeModel{
        filepicker: fp,
        progress: progress.New(progress.WithScaledGradient("#0f8a7f", "#0979ad")),
        channel: make(chan float64, 1),
        helpStyle: helpStyle,
        titleStyle: titleStyle,
        state: chooseFile,
        destFileInput: ti,
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

    err := recode.Encode(m.srcFile, m.destFile, m.channel)

    if err != nil {
        panic(err)
    }

    m.Update(recodeMsg(true))
}

func startEncode(m recodeModel) tea.Cmd {
    return func() tea.Msg {
        go encode(m)
        return tickMsg{}
    }
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

    if _, err:= tea.NewProgram(newModel()).Run(); err != nil {
        return err
    }

    return nil
}
