package create

import (
	"io/fs"
	"os"
	"path/filepath"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Config struct {
    DefaultPath string `yaml:"rootPath"`
}

type model struct {
    options []string
    cursor int
    selected map[int]string
    list list.Model
}

type item struct {
    title, desc string
}

func (i item) Title() string { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

func newModel(options []string) model {
    items := make([]list.Item, 0)

    for _, option := range options {
        items = append(items, item{ title: option, desc: "None" })
    }

    return model{
        options: options,
        list: list.New(items, list.NewDefaultDelegate(), 0, 0),
    }
}

func (m model) Init() tea.Cmd {
    return nil
}

func(m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case tea.KeyCtrlC.String(), "q":
            return m, tea.Quit
        }
    case tea.WindowSizeMsg:
        h,v := lipgloss.NewStyle().Margin(1, 2).GetFrameSize()
        m.list.SetSize(msg.Width-h, msg.Height-v)
    }

    var cmd tea.Cmd

    m.list, cmd = m.list.Update(msg)
    return m, cmd
}

func (m model) View() string {
    return m.list.View()
}

func Run(cfg  *Config) error {
    root := cfg.DefaultPath
    files := os.DirFS(root)
    options := make([]string, 0)

    fs.WalkDir(files, ".", func(path string, d fs.DirEntry, err error) error {
        if err != nil {
            return err
        }

        if filepath.Dir(path) == "." && !d.IsDir() {
            options = append(options, path)
        }

        return nil
    })

    p := tea.NewProgram(newModel(options))

    if _, err := p.Run(); err != nil {
        return err
    }

    return nil
}
