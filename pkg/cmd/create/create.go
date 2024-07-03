package create

import (
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Config struct {
    DefaultPath string `yaml:"rootPath"`
}

type KeyMap struct {
    Up key.Binding
    Down key.Binding
    Toggle key.Binding
    Start key.Binding
    Quit key.Binding
}

type model struct {
    options []string
    selected map[string]struct{}
    list list.Model
    keys KeyMap
    itemStyle struct {
        normal lipgloss.Style
        active lipgloss.Style
        selected lipgloss.Style
        selectedactive lipgloss.Style
    }
}

type item struct {
    title, desc string
}

func (i item) Title() string { return i.title }
func (i item) CleanTitle() string {
    return strings.Replace(i.title, "--selected--", "", 1)
}
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

type itemDelegate struct {
    styles struct {
        normal lipgloss.Style
        active lipgloss.Style
        selected lipgloss.Style
        selectedactive lipgloss.Style
    }
}

func (d itemDelegate) Height() int { return 1 }
func (d itemDelegate) Spacing() int { return 0 }
func (d itemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case " ":
            if strings.HasPrefix(m.SelectedItem().FilterValue(), "--selected--") {
                m.SetItem(m.Index(), item{ title: strings.Replace(m.SelectedItem().FilterValue(), "--selected--", "", 1) } )
                return nil
            }

            m.SetItem(m.Index(), item{ title: fmt.Sprintf("--selected--%s", m.SelectedItem().FilterValue())} )
            // log.Println(m.SelectedItem().FilterValue())
        }
    }
    return nil
}
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
    i, ok := listItem.(item)

    if !ok {
        return
    }

    if index == m.Index() && strings.HasPrefix(listItem.FilterValue(), "--selected--") {
        s := strings.Replace(i.Title(), "--selected--", "", 1) 
        fmt.Fprint(w, d.styles.selectedactive.Render("x " + s))
        return
    }

    if index == m.Index() {
        fmt.Fprint(w, d.styles.active.Render("> " + i.Title()))
        return
    }

    if strings.HasPrefix(listItem.FilterValue(), "--selected--") {
        s := strings.Replace(i.Title(), "--selected--", "", 1) 
        fmt.Fprint(w, d.styles.selected.Render("x " + s))
        return
    }

    fmt.Fprint(w, d.styles.normal.Render(i.Title())) 
}

func newModel(options []string) model {
    items := make([]list.Item, 0)

    for _, option := range options {
        items = append(items, item{ title: option, desc: "None" })
    }

    selectedactive := lipgloss.NewStyle().
        Foreground(lipgloss.Color("#EF0038")).
        Background(lipgloss.Color("#D2D2D2")).
        Bold(true).
        PaddingLeft(2)

    selected := lipgloss.NewStyle().
        Background(lipgloss.Color("#EF0038")).
        Foreground(lipgloss.Color("#121212")).
        Bold(true).
        PaddingLeft(2)

    active := lipgloss.NewStyle().
        Foreground(lipgloss.Color("#EF0038")).
        PaddingLeft(2)

    normal := lipgloss.NewStyle().
        Foreground(lipgloss.Color("#FCFCFC")).
        PaddingLeft(4)

    list := list.New(items, itemDelegate{struct{normal lipgloss.Style; active lipgloss.Style; selected lipgloss.Style; selectedactive lipgloss.Style} {
        selected: selected,
        selectedactive: selectedactive,
        active: active,
        normal: normal,
    }}, 0, 0)

    list.Title = "Choose files to recode"
    list.SetShowStatusBar(false)
    list.SetFilteringEnabled(false)
    list.SetShowHelp(true)

    return model{
        options: options,
        list: list,
        selected: make(map[string]struct{}),
        itemStyle: struct{normal lipgloss.Style; active lipgloss.Style; selected lipgloss.Style; selectedactive lipgloss.Style} {
            normal: normal,
            active: active,
            selected: selected,
            selectedactive: selectedactive,
        },
        keys: KeyMap{
            Up: key.NewBinding(
                key.WithKeys("k", "up"),
                key.WithHelp("k/up", "move up"),
            ),
            Down: key.NewBinding(
                key.WithKeys("j", "down"),
                key.WithHelp("j/down", "move down"),
            ),
            Start: key.NewBinding(
                key.WithKeys("enter", "return"),
                key.WithHelp("enter/return", "start recoding"),
            ),
            Toggle: key.NewBinding(
                key.WithKeys("space", " "),
                key.WithHelp("space", "select/deselect"),
            ),
            Quit: key.NewBinding(
                key.WithKeys("q", "ctrl+c"),
                key.WithHelp("q/ctrc+c", "quit"),
            ),
        },
    }
}

func (m model) Init() tea.Cmd {
    return nil
}

func(m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch {
        case key.Matches(msg, m.keys.Quit):
            return m, tea.Quit
        case key.Matches(msg, m.keys.Toggle):
            i, ok := m.list.SelectedItem().(item)
            
            if ok {
                if _, ok := m.selected[i.CleanTitle()]; ok {
                    delete(m.selected, i.CleanTitle())
                } else {
                    m.selected[i.CleanTitle()] = struct{}{}
                } 
            }
        case key.Matches(msg, m.keys.Start):
            log.Print(m.selected)
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
