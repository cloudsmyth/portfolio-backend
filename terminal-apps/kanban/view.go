package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	todoColumnStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("203")).
			Padding(1, 2).
			Width(30)

	inProgressColumnStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("220")).
				Padding(1, 2).
				Width(30)

	doneColumnStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("46")).
			Padding(1, 2).
			Width(30)

	selectedColumnStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("87")).
				Padding(1, 2).
				Width(30)

	cardStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1).
			MarginBottom(1)

	selectedCardStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("170")).
				Background(lipgloss.Color("235")).
				Padding(0, 1).
				MarginBottom(1)

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("255")).
			Align(lipgloss.Center)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Padding(1, 0)

	inputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("170")).
			Padding(0, 1)
)

func (m model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	var s strings.Builder

	title := titleStyle.Width(m.width).Render("Kanban Board")
	s.WriteString(title)
	s.WriteString("\n\n")

	columns := make([]string, len(m.board.Columns))
	for i, col := range m.board.Columns {
		columns[i] = m.renderColumn(i, col)
	}

	s.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, columns...))
	s.WriteString("\n\n")

	if m.addingCard {
		prompt := "New card title: "
		input := inputStyle.Render(prompt + m.cardInput + "█")
		s.WriteString(input)
		s.WriteString("\n")
		s.WriteString(helpStyle.Render("Enter to save ∙ Esc to cancel"))
	} else {
		help := helpStyle.Render(
			"h/l or ←/→: move between columns ∙ j/k or ↓/↑: move between cards\n" +
				"H/L: move card left/right ∙ a: add card ∙ d: delete card ∙ q:quit",
		)
		s.WriteString(help)
	}
	return s.String()
}

func (m model) renderColumn(colIndex int, col Column) string {
	isSelected := colIndex == m.selectedCol

	var style lipgloss.Style
	switch col.Type {
	case ColumnTodo:
		if isSelected {
			style = selectedColumnStyle
		} else {
			style = todoColumnStyle
		}
	case ColumnInProgress:
		if isSelected {
			style = selectedColumnStyle
		} else {
			style = inProgressColumnStyle
		}
	case ColumnDone:
		if isSelected {
			style = selectedColumnStyle
		} else {
			style = doneColumnStyle
		}
	default:
		style = lipgloss.NewStyle().
			BorderForeground(lipgloss.Color("87")).
			Padding(1, 2).
			Width(30)
	}

	var content strings.Builder

	colTitle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("255")).
		Render(col.Title)
	content.WriteString(colTitle)
	content.WriteString(fmt.Sprintf(" (%d)", len(col.Cards)))
	content.WriteString("\n\n")

	if len(col.Cards) == 0 {
		emptyText := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Italic(true).
			Render("No cards")
		content.WriteString(emptyText)
	} else {
		for cardIdx, card := range col.Cards {
			cardContent := m.renderCard(colIndex, cardIdx, card)
			content.WriteString(cardContent)
		}
	}

	return style.Render(content.String())
}

func (m model) renderCard(colIndex, cardIdx int, card Card) string {
	isSelected := colIndex == m.selectedCol && cardIdx == m.selectedCard
	style := cardStyle
	if isSelected {
		style = selectedCardStyle
	}

	cardText := card.Title

	maxLen := 20
	if isSelected {
		maxLen -= 2
	}
	if len(cardText) > maxLen {
		cardText = cardText[:maxLen-3] + "..."
	}

	if isSelected {
		cardText = "▸ " + cardText
	}

	return style.Render(cardText)
}
