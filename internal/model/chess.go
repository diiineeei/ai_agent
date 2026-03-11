package model

import "time"

// ChessGame armazena o estado persistente de uma partida de xadrez.
type ChessGame struct {
	SessionID string    `bson:"session_id" json:"session_id"`
	AgentID   string    `bson:"agent_id"   json:"agent_id"`
	FEN       string    `bson:"fen"        json:"fen"`       // posição atual após o último lance
	Moves     []string  `bson:"moves"      json:"moves"`     // histórico completo de lances
	Status    string    `bson:"status"     json:"status"`    // playing | checkmate | stalemate | draw
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}
