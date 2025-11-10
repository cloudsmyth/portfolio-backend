package main

import (
	"time"

	"github.com/google/uuid"
)

type ColumnType string

const (
	ColumnTodo       ColumnType = "todo"
	ColumnInProgress ColumnType = "in-progress"
	ColumnDone       ColumnType = "done"
)

type Card struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created-at"`
}

type Column struct {
	Type  ColumnType `json:"type"`
	Title string     `json:"title"`
	Cards []Card     `json:"cards"`
}

type Board struct {
	Columns []Column `json:"columns"`
}

func NewBoard() Board {
	return Board{
		Columns: []Column{
			{Type: ColumnTodo, Title: "To Do", Cards: []Card{}},
			{Type: ColumnInProgress, Title: "In Progress", Cards: []Card{}},
			{Type: ColumnDone, Title: "Done", Cards: []Card{}},
		},
	}
}

func NewCard(title string) Card {
	return Card{
		ID:        uuid.New().String(),
		Title:     title,
		CreatedAt: time.Now(),
	}
}

func (b *Board) AddCard(ct ColumnType, card Card) {
	for i := range b.Columns {
		if b.Columns[i].Type == ct {
			b.Columns[i].Cards = append(b.Columns[i].Cards, card)
			return
		}
	}
}

func (b *Board) DeleteCard(cardID string) {
	for i := range b.Columns {
		for j, card := range b.Columns[i].Cards {
			if card.ID == cardID {
				b.Columns[i].Cards = append(b.Columns[i].Cards[:j], b.Columns[i].Cards[j+1:]...)
				return
			}
		}
	}
}

func (b *Board) MoveCard(cardID string, toColumn ColumnType) {
	var card Card
	var fromColumnIdx int
	var fromCardIdx int
	found := false

	for i := range b.Columns {
		for j, c := range b.Columns[i].Cards {
			if c.ID == cardID {
				card = c
				fromColumnIdx = i
				fromCardIdx = j
				found = true
				break
			}
		}
		if found {
			break
		}
	}

	if !found {
		return
	}

	targetColumnIdx := -1
	for i := range b.Columns {
		if b.Columns[i].Type == toColumn {
			targetColumnIdx = i
			break
		}
	}

	if targetColumnIdx == -1 {
		return
	}

	if fromColumnIdx == targetColumnIdx {
		return
	}

	b.Columns[fromColumnIdx].Cards = append(
		b.Columns[fromColumnIdx].Cards[:fromCardIdx],
		b.Columns[fromColumnIdx].Cards[fromCardIdx+1:]...,
	)

	b.Columns[targetColumnIdx].Cards = append(b.Columns[targetColumnIdx].Cards, card)
}
