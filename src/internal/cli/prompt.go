package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/studiowebux/restcli/internal/types"
)

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2).Bold(true)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	helpStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).MarginTop(1).MarginLeft(2)
)

type item struct {
	value    string
	alias    string
	index    int
	isActive bool
}

func (i item) FilterValue() string {
	return i.value + " " + i.alias
}

func (i item) Title() string {
	title := i.value
	if i.alias != "" {
		title += fmt.Sprintf(" (alias: %s)", i.alias)
	}
	if i.isActive {
		title += " [active]"
	}
	return title
}

func (i item) Description() string { return "" }

type selectorModel struct {
	list     list.Model
	choice   string
	quitting bool
	varName  string
}

func (m selectorModel) Init() tea.Cmd {
	return nil
}

func (m selectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			m.choice = ""
			return m, tea.Quit

		case "enter":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.choice = i.value
			}
			m.quitting = true
			return m, tea.Quit

		case "c", "C":
			// Custom input mode
			m.choice = "!CUSTOM!"
			m.quitting = true
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m selectorModel) View() string {
	if m.quitting {
		return ""
	}

	help := helpStyle.Render("↑/↓: navigate • enter: select • c: custom value • q/ctrl+c: cancel")
	return fmt.Sprintf("%s\n\n%s", m.list.View(), help)
}

// promptForMultiValueVariable shows an interactive list to select from multi-value options
func promptForMultiValueVariable(varName string, mv *types.MultiValueVariable) (string, error) {
	// Build list items
	items := make([]list.Item, 0, len(mv.Options))

	// Create reverse alias map (index -> alias names)
	indexToAliases := make(map[int][]string)
	if mv.Aliases != nil {
		for alias, idx := range mv.Aliases {
			indexToAliases[idx] = append(indexToAliases[idx], alias)
		}
	}

	for i, opt := range mv.Options {
		aliasStr := ""
		if aliases, ok := indexToAliases[i]; ok && len(aliases) > 0 {
			aliasStr = strings.Join(aliases, ", ")
		}
		items = append(items, item{
			value:    opt,
			alias:    aliasStr,
			index:    i,
			isActive: i == mv.Active,
		})
	}

	const defaultWidth = 80
	const listHeight = 14

	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.Title = fmt.Sprintf("Select value for variable: %s", varName)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = titleStyle

	// Set initial selection to active index (if valid)
	if mv.Active >= 0 && mv.Active < len(items) {
		l.Select(mv.Active)
	}

	m := selectorModel{list: l, varName: varName}

	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return "", fmt.Errorf("error running selector: %w", err)
	}

	result := finalModel.(selectorModel)
	if result.choice == "!CUSTOM!" {
		// Prompt for custom value
		return promptForCustomValue(varName)
	}

	if result.choice == "" {
		return "", fmt.Errorf("selection cancelled")
	}

	return result.choice, nil
}

// itemDelegate is a custom list item delegate
type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i.Title())

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

// promptForCustomValue prompts for a custom value input
func promptForCustomValue(varName string) (string, error) {
	fmt.Printf("\nEnter custom value for %s: ", varName)
	var value string
	_, err := fmt.Scanln(&value)
	if err != nil {
		return "", fmt.Errorf("failed to read input: %w", err)
	}
	return value, nil
}
