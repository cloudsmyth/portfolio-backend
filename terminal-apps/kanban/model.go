package main

import tea "github.com/charmbracelet/bubbletea"

type model struct {
	board        Board
	selectedCol  int
	selectedCard int
	width        int
	height       int
	addingCard   bool
	cardInput    string
	cursor       int
}

func initialModel() model {
	board := NewBoard()
	board.AddCard(ColumnTodo, NewCard("Clean up mobile view"))
	board.AddCard(ColumnTodo, NewCard("Zig cli practice"))
	board.AddCard(ColumnInProgress, NewCard("Add Pokemon to trading card cli"))
	board.AddCard(ColumnDone, NewCard("Deploy portfolio"))

	return model{
		board:        board,
		selectedCol:  0,
		selectedCard: 0,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		if m.addingCard {
			return m.handleCardInput(msg)
		}

		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "left", "h":
			if m.selectedCol > 0 {
				m.selectedCol--
				m.selectedCard = 0
			}

		case "right", "l":
			if m.selectedCol < len(m.board.Columns)-1 {
				m.selectedCol++
				m.selectedCard = 0
			}

		case "up", "k":
			if m.selectedCard > 0 {
				m.selectedCard--
			}

		case "down", "j":
			if m.selectedCard < len(m.board.Columns[m.selectedCol].Cards)-1 {
				m.selectedCard++
			}

		case "a":
			m.addingCard = true
			m.cardInput = ""

		case "d":
			m.deleteSelectedCard()

		case "H":
			m.moveSelectedCard(-1)

		case "L":
			m.moveSelectedCard(1)
		}
	}

	return m, nil
}

func (m *model) handleCardInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if m.cardInput != "" {
			card := NewCard(m.cardInput)
			m.board.AddCard(m.board.Columns[m.selectedCol].Type, card)
		}
		m.addingCard = false
		m.cardInput = ""
	case "esc":
		m.addingCard = false
		m.cardInput = ""
	case "backspace":
		if len(m.cardInput) > 0 {
			m.cardInput = m.cardInput[:len(m.cardInput)-1]
		}
	default:
		m.cardInput += msg.String()
	}

	return m, nil
}

func (m *model) deleteSelectedCard() {
	col := &m.board.Columns[m.selectedCol]
	if len(col.Cards) == 0 {
		return
	}

	if m.selectedCard >= len(col.Cards) {
		m.selectedCard = len(col.Cards) - 1
	}

	cardID := col.Cards[m.selectedCard].ID
	m.board.DeleteCard(cardID)

	if m.selectedCard >= len(col.Cards) && m.selectedCard > 0 {
		m.selectedCard--
	}
}

func (m *model) moveSelectedCard(direction int) {
	col := &m.board.Columns[m.selectedCol]
	if len(col.Cards) == 0 {
		return
	}

	targetCol := m.selectedCol + direction
	if targetCol < 0 || targetCol >= len(m.board.Columns) {
		return
	}

	cardID := col.Cards[m.selectedCard].ID
	m.board.MoveCard(cardID, m.board.Columns[targetCol].Type)
	m.selectedCol = targetCol
	m.selectedCard = len(m.board.Columns[targetCol].Cards) - 1
}
