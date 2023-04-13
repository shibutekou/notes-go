package main

import (
	"context"
	"fmt"
	notemodel "github.com/bruma1994/dyngo/internal/model"
	"github.com/bruma1994/dyngo/internal/repository"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/urfave/cli/v2"
	"io"
	"log"
	"os"
	"os/user"
	"strconv"
	"strings"
	"time"
)

const listHeight = 14

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)

var (
	collection, _ = repository.InitMongoClient()
	notesRepo     = repository.NewNotesRepository(collection, counter)
)

var counter = 0
var ctx = context.Background()

func main() {
	app := initCLI()
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func initCLI() *cli.App {
	app := &cli.App{
		Action: func(cCtx *cli.Context) error {
			var items []list.Item

			notes, _ := notesRepo.All(ctx)
			for _, note := range notes {
				items = append(items, item(note.Name))
			}

			const defaultWidth = 20

			l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
			l.Title = `What note do you want to read? Execute "add" to add a new note`
			l.SetShowStatusBar(false)
			l.SetFilteringEnabled(false)
			l.Styles.Title = titleStyle
			l.Styles.PaginationStyle = paginationStyle
			l.Styles.HelpStyle = helpStyle

			m := model{list: l}

			if _, err := tea.NewProgram(m).Run(); err != nil {
				return err
			}
			return nil

		},
		Commands: []*cli.Command{
			{
				Name:    "add",
				Aliases: []string{"a"},
				Usage:   "add a note to the database",
				Action: func(c *cli.Context) error {
					note := notemodel.Note{
						Name:      c.Args().Get(0),
						Text:      c.Args().Get(1),
						Tag:       c.Args().Get(2),
						CreatedAt: time.Now(),
					}
					err := notesRepo.Add(ctx, note)
					if err != nil {
						return err
					}
					return nil
				},
			},
			{
				Name:    "del",
				Aliases: []string{"d"},
				Usage:   "delete a note from database",
				Action: func(c *cli.Context) error {
					id, _ := strconv.Atoi(c.Args().Get(1))
					err := notesRepo.Delete(ctx, int32(id))
					if err != nil {
						return err
					}
					return nil
				},
			},
			{
				Name:    "mine",
				Aliases: []string{"s"},
				Usage:   "show note from database",
				Action: func(c *cli.Context) error {
					current, _ := user.Current()
					author := current.Username
					_, err := notesRepo.ByAuthor(ctx, author)
					if err != nil {
						return err
					}
					return nil
				},
			},
			{
				Name:    "name",
				Aliases: []string{"n"},
				Usage:   "show note from database by name",
				Action: func(c *cli.Context) error {
					note, err := notesRepo.ByName(ctx, c.Args().First())
					fmt.Println(note.Text)
					if err != nil {
						return err
					}
					return nil
				},
			},
		},
	}

	return app
}

type item string

func (i item) FilterValue() string { return "" }

type itemDelegate struct{}

func (d itemDelegate) Height() int                               { return 1 }
func (d itemDelegate) Spacing() int                              { return 0 }
func (d itemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

type model struct {
	list     list.Model
	choice   string
	quitting bool
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) View() string {
	notes, _ := notesRepo.All(ctx)
	var names []string
	for _, v := range notes {
		names = append(names, v.Name)
	}

	if m.choice != "" {
		note, _ := notesRepo.ByName(ctx, m.choice)
		m.list.View()
		return quitTextStyle.Render(fmt.Sprintf("%s", note.Text))
	}

	if m.quitting {
		return quitTextStyle.Render("Not hungry? That’s cool.")
	}
	return "\n" + m.list.View()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.choice = string(i)
			}
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}
