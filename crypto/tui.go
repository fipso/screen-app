package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	table table.Model
}

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

func pullPrices() tea.Msg {
	// Block till update
	<-pricesUpdated

	return true
}

func (m model) Init() tea.Cmd { return pullPrices }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	case bool:
		var rows []table.Row

		// Sort keys
		keys := make([]string, 0)
		for k, _ := range currencies {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, key := range keys {
			currency := currencies[key]

			row := table.Row{
				currency.name,
				fmt.Sprintf("%08.2f$", currency.price),
			}
			rows = append(rows, row)
		}

		m.table.SetRows(rows)

		return m, pullPrices
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return baseStyle.Render(m.table.View()) + "\n"
}

func setupTUI() {
	columns := []table.Column{
		{Title: "Name", Width: 10},
		{Title: "Price", Width: 15},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows([]table.Row{}),
		table.WithHeight(len(currencies)+1),
	)

	s := table.DefaultStyles()
        s.Selected = lipgloss.NewStyle()
	t.SetStyles(s)

	m := model{t}
	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
