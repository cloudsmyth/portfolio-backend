package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ScryfallResponse struct {
	Object     string `json:"object"`
	TotalCards int    `json:"total_cards"`
	Data       []Card `json:"data"`
}

type Card struct {
	Name       string   `json:"name"`
	ManaCost   string   `json:"mana_cost"`
	TypeLine   string   `json:"type_line"`
	OracleText string   `json:"oracle_text"`
	Power      string   `json:"power"`
	Toughness  string   `json:"toughness"`
	Colors     []string `json:"colors"`
	SetName    string   `json:"set_name"`
	Rarity     string   `json:"rarity"`
}

const (
	scryfallAPI    = "https://api.scryfall.com/cards/search"
	rateLimitDelay = 100 * time.Millisecond
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("170")).
			Padding(0, 1)

	inputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205"))

	cardTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("86"))

	cardDetailStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))
)

type viewMode int

const (
	searchView viewMode = iota
	resultsView
	detailView
)

type model struct {
	textInput    textinput.Model
	cards        []Card
	list         list.Model
	selectedCard *Card
	mode         viewMode
	searching    bool
	err          error
	width        int
	height       int
}

type searchResultMsg struct {
	cards []Card
	err   error
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Enter card name..."
	ti.Focus()
	ti.CharLimit = 100
	ti.Width = 50

	return model{
		textInput: ti,
		mode:      searchView,
		width:     80,
		height:    24,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if msg.Width != m.width || msg.Height != m.height {
			m.width = msg.Width
			m.height = msg.Height
			if m.mode == resultsView {
				m.list.SetSize(m.width, m.height-10)
			}
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.mode == searchView {
				return m, tea.Quit
			}
			m.mode = searchView
			m.err = nil
			m.textInput.Focus()
			return m, nil

		case "esc":
			if m.mode == detailView {
				m.mode = resultsView
				return m, nil
			} else if m.mode == resultsView {
				m.mode = searchView
				m.textInput.Focus()
				return m, nil
			}

		case "enter":
			if m.mode == searchView && !m.searching {
				query := m.textInput.Value()
				if query != "" {
					m.searching = true
					m.err = nil
					return m, searchCards(query)
				}
			} else if m.mode == resultsView {
				if len(m.cards) > 0 {
					selected := m.list.Index()
					if selected < len(m.cards) {
						m.selectedCard = &m.cards[selected]
						m.mode = detailView
					}
				}
				return m, nil
			}
		}

	case searchResultMsg:
		m.searching = false
		m.err = msg.err
		if msg.err == nil && len(msg.cards) > 0 {
			m.cards = msg.cards
			m.mode = resultsView
			items := make([]list.Item, len(msg.cards))
			for i, card := range msg.cards {
				items[i] = cardItem{card: card}
			}
			m.list = list.New(items, list.NewDefaultDelegate(), m.width, m.height-10)
			m.list.Title = fmt.Sprintf("Found %d cards", len(msg.cards))
		}
		return m, nil
	}

	var cmd tea.Cmd
	if m.mode == searchView {
		m.textInput, cmd = m.textInput.Update(msg)
	} else if m.mode == resultsView {
		m.list, cmd = m.list.Update(msg)
	}
	return m, cmd
}

func (m model) View() string {
	switch m.mode {
	case searchView:
		return m.searchView()
	case resultsView:
		return m.resultsView()
	case detailView:
		return m.detailView()
	}
	return ""
}

func (m model) searchView() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("🃏 MTG Card Search"))
	b.WriteString("\n\n")

	b.WriteString(inputStyle.Render(m.textInput.View()))
	b.WriteString("\n\n")

	if m.searching {
		b.WriteString("Searching...\n")
	}

	if m.err != nil {
		b.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v\n", m.err)))
		b.WriteString("\n")
	}

	b.WriteString(helpStyle.Render("Press Enter to search • q to quit"))

	return b.String()
}

func (m model) resultsView() string {
	var b strings.Builder

	if len(m.cards) == 0 {
		b.WriteString(titleStyle.Render("No cards found"))
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render("Press esc to search again • q to quit"))
		return b.String()
	}

	b.WriteString(m.list.View())
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("↑/↓: navigate • Enter: view details • esc: back • q: quit"))

	return b.String()
}

func (m model) detailView() string {
	if m.selectedCard == nil {
		return "No card selected"
	}

	var b strings.Builder
	card := m.selectedCard

	b.WriteString(cardTitleStyle.Render(fmt.Sprintf("%s %s", card.Name, card.ManaCost)))
	b.WriteString("\n\n")

	b.WriteString(cardDetailStyle.Render("Type: "))
	b.WriteString(card.TypeLine)
	b.WriteString("\n\n")

	if card.OracleText != "" {
		b.WriteString(cardDetailStyle.Render("Text:\n"))
		b.WriteString(wrapText(card.OracleText, 70))
		b.WriteString("\n\n")
	}

	if card.Power != "" && card.Toughness != "" {
		b.WriteString(cardDetailStyle.Render("Power/Toughness: "))
		b.WriteString(fmt.Sprintf("%s/%s\n\n", card.Power, card.Toughness))
	}

	b.WriteString(cardDetailStyle.Render("Set: "))
	b.WriteString(fmt.Sprintf("%s (%s)\n", card.SetName, card.Rarity))

	if len(card.Colors) > 0 {
		b.WriteString(cardDetailStyle.Render("Colors: "))
		b.WriteString(strings.Join(card.Colors, ", "))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("Press esc to go back • q to quit"))

	return b.String()
}

func wrapText(text string, width int) string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return text
	}

	var lines []string
	var currentLine strings.Builder

	for _, word := range words {
		if currentLine.Len()+len(word)+1 > width {
			lines = append(lines, currentLine.String())
			currentLine.Reset()
		}
		if currentLine.Len() > 0 {
			currentLine.WriteString(" ")
		}
		currentLine.WriteString(word)
	}

	if currentLine.Len() > 0 {
		lines = append(lines, currentLine.String())
	}

	return strings.Join(lines, "\n")
}

type cardItem struct {
	card Card
}

func (i cardItem) Title() string       { return fmt.Sprintf("%s %s", i.card.Name, i.card.ManaCost) }
func (i cardItem) Description() string { return i.card.TypeLine }
func (i cardItem) FilterValue() string { return i.card.Name }

func searchCards(query string) tea.Cmd {
	return func() tea.Msg {
		params := url.Values{}
		params.Add("q", query)
		params.Add("order", "name")

		reqURL := fmt.Sprintf("%s?%s", scryfallAPI, params.Encode())

		resp, err := http.Get(reqURL)
		if err != nil {
			return searchResultMsg{err: fmt.Errorf("failed to make request: %w", err)}
		}
		defer resp.Body.Close()

		if resp.StatusCode == 429 {
			return searchResultMsg{err: fmt.Errorf("rate limited by Scryfall API")}
		}

		if resp.StatusCode != 200 {
			body, _ := io.ReadAll(resp.Body)
			return searchResultMsg{err: fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))}
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return searchResultMsg{err: fmt.Errorf("failed to read response: %w", err)}
		}

		var result ScryfallResponse
		if err := json.Unmarshal(body, &result); err != nil {
			return searchResultMsg{err: fmt.Errorf("failed to parse JSON: %w", err)}
		}

		time.Sleep(rateLimitDelay)

		return searchResultMsg{cards: result.Data}
	}
}

func main() {
	opts := []tea.ProgramOption{
		tea.WithAltScreen(),
		tea.WithInput(os.Stdin),
	}
	p := tea.NewProgram(initialModel(), opts...)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
