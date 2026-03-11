package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"ai_agent/internal/agent"
	"ai_agent/internal/model"
	"ai_agent/internal/repository"

	chess "github.com/corentings/chess/v2"
	"google.golang.org/genai"
)

type ChessHandler struct {
	geminiClient    *genai.Client
	chessRepo       repository.ChessRepository
	agentConfigRepo repository.AgentConfigRepository
	sessionRepo     repository.SessionRepository
}

func NewChessHandler(
	geminiClient *genai.Client,
	chessRepo repository.ChessRepository,
	agentConfigRepo repository.AgentConfigRepository,
	sessionRepo repository.SessionRepository,
) *ChessHandler {
	return &ChessHandler{
		geminiClient:    geminiClient,
		chessRepo:       chessRepo,
		agentConfigRepo: agentConfigRepo,
		sessionRepo:     sessionRepo,
	}
}

// ── request / response types ──────────────────────────────────────────────────

type chessStartRequest struct {
	SessionID     string `json:"session_id"`
	AgentConfigID string `json:"agent_config_id"`
}

type chessMoveRequest struct {
	SessionID string `json:"session_id"`
	Move      string `json:"move"` // UCI: "e2e4"
}

type chessStateResponse struct {
	FEN        string   `json:"fen"`
	Moves      []string `json:"moves"`
	LegalMoves []string `json:"legal_moves"`
	Status     string   `json:"status"`
	AIMove     string   `json:"ai_move,omitempty"`
	Analysis   string   `json:"analysis,omitempty"`
}

// ── handlers ──────────────────────────────────────────────────────────────────

// Start handles POST /chess/start — cria nova partida e retorna estado inicial.
func (h *ChessHandler) Start(w http.ResponseWriter, r *http.Request) {
	var req chessStartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "body inválido: "+err.Error(), http.StatusBadRequest)
		return
	}
	if req.SessionID == "" || req.AgentConfigID == "" {
		jsonError(w, "session_id e agent_config_id são obrigatórios", http.StatusBadRequest)
		return
	}

	game := chess.NewGame()
	legalMoves := encodeMoves(game.ValidMoves())

	state := model.ChessGame{
		SessionID: req.SessionID,
		AgentID:   req.AgentConfigID,
		FEN:       game.FEN(),
		Moves:     []string{},
		Status:    "playing",
	}
	if err := h.chessRepo.Save(r.Context(), state); err != nil {
		jsonError(w, "erro ao salvar jogo: "+err.Error(), http.StatusInternalServerError)
		return
	}

	jsonResponse(w, http.StatusOK, chessStateResponse{
		FEN:        state.FEN,
		Moves:      state.Moves,
		LegalMoves: legalMoves,
		Status:     state.Status,
	})
}

// Move handles POST /chess/move — valida lance humano, chama agente para as pretas, retorna novo estado.
func (h *ChessHandler) Move(w http.ResponseWriter, r *http.Request) {
	var req chessMoveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "body inválido: "+err.Error(), http.StatusBadRequest)
		return
	}
	if req.SessionID == "" || req.Move == "" {
		jsonError(w, "session_id e move são obrigatórios", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	state, err := h.chessRepo.Load(ctx, req.SessionID)
	if err != nil || state == nil {
		jsonError(w, "partida não encontrada; inicie com POST /chess/start", http.StatusNotFound)
		return
	}

	// Reconstrói posição a partir do FEN armazenado
	fenOpt, err := chess.FEN(state.FEN)
	if err != nil {
		jsonError(w, "FEN inválido no banco: "+err.Error(), http.StatusInternalServerError)
		return
	}
	game := chess.NewGame(fenOpt)

	// Valida e aplica lance humano
	if err := game.PushNotationMove(req.Move, chess.UCINotation{}, nil); err != nil {
		jsonError(w, "lance inválido: "+err.Error(), http.StatusUnprocessableEntity)
		return
	}
	state.Moves = append(state.Moves, req.Move)

	// Verifica fim de jogo após lance humano
	if game.Outcome() != chess.NoOutcome {
		state.FEN = game.FEN()
		state.Status = outcomeToStatus(game.Outcome(), game.Method())
		_ = h.chessRepo.Save(ctx, *state)
		jsonResponse(w, http.StatusOK, chessStateResponse{
			FEN:    state.FEN,
			Moves:  state.Moves,
			Status: state.Status,
		})
		return
	}

	// Lances legais para as pretas
	legalForBlack := encodeMoves(game.ValidMoves())

	// Carrega configuração do agente
	cfg, err := h.agentConfigRepo.GetByID(ctx, state.AgentID)
	if err != nil || cfg == nil {
		jsonError(w, "agente não encontrado", http.StatusBadRequest)
		return
	}

	// Chama agente sem skills e sem histórico de chat
	var a agent.Agent
	if cfg.Provider == "ollama" {
		a = agent.NewOllama(cfg.BaseURL, cfg.Model, cfg.SystemInstruction, nil)
	} else {
		a = agent.NewWithRepo(h.geminiClient, cfg.Model, cfg.SystemInstruction, nil)
	}

	prompt := buildChessPrompt(state.Moves, game.FEN(), legalForBlack)
	aiResponse, _, err := a.Send(ctx, "", prompt)
	if err != nil {
		jsonError(w, "erro ao consultar agente: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Extrai e aplica lance da IA
	aiMove := extractChessMove(aiResponse, req.Move, legalForBlack)
	if aiMove == "" {
		// Fallback: primeiro lance legal
		if len(legalForBlack) > 0 {
			aiMove = legalForBlack[0]
		} else {
			jsonError(w, "sem lances legais para as pretas", http.StatusUnprocessableEntity)
			return
		}
	}

	if err := game.PushNotationMove(aiMove, chess.UCINotation{}, nil); err != nil {
		// Lance extraído inválido — usa primeiro legal
		aiMove = legalForBlack[0]
		_ = game.PushNotationMove(aiMove, chess.UCINotation{}, nil)
	}
	state.Moves = append(state.Moves, aiMove)
	state.FEN = game.FEN()

	// Verifica fim de jogo após lance da IA
	if game.Outcome() != chess.NoOutcome {
		state.Status = outcomeToStatus(game.Outcome(), game.Method())
	} else {
		state.Status = "playing"
	}

	legalForWhite := encodeMoves(game.ValidMoves())

	if err := h.chessRepo.Save(ctx, *state); err != nil {
		jsonError(w, "erro ao salvar jogo: "+err.Error(), http.StatusInternalServerError)
		return
	}

	jsonResponse(w, http.StatusOK, chessStateResponse{
		FEN:        state.FEN,
		Moves:      state.Moves,
		LegalMoves: legalForWhite,
		Status:     state.Status,
		AIMove:     aiMove,
		Analysis:   extractAnalysis(aiResponse),
	})
}

// State handles GET /chess/state — retorna estado atual com lances legais.
func (h *ChessHandler) State(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		jsonError(w, "session_id obrigatório", http.StatusBadRequest)
		return
	}

	state, err := h.chessRepo.Load(r.Context(), sessionID)
	if err != nil {
		jsonError(w, "erro ao carregar jogo: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if state == nil {
		jsonResponse(w, http.StatusOK, nil)
		return
	}

	fenOpt, err := chess.FEN(state.FEN)
	if err != nil {
		jsonResponse(w, http.StatusOK, state)
		return
	}
	game := chess.NewGame(fenOpt)

	jsonResponse(w, http.StatusOK, chessStateResponse{
		FEN:        state.FEN,
		Moves:      state.Moves,
		LegalMoves: encodeMoves(game.ValidMoves()),
		Status:     state.Status,
	})
}

// Reset handles DELETE /chess/game — remove partida salva.
func (h *ChessHandler) Reset(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		jsonError(w, "session_id obrigatório", http.StatusBadRequest)
		return
	}
	if err := h.chessRepo.Delete(r.Context(), sessionID); err != nil {
		jsonError(w, "erro ao remover jogo: "+err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, http.StatusOK, map[string]string{"message": "partida removida"})
}

// ── helpers ───────────────────────────────────────────────────────────────────

func encodeMoves(moves []chess.Move) []string {
	result := make([]string, len(moves))
	for i, m := range moves {
		result[i] = m.String()
	}
	return result
}

func outcomeToStatus(o chess.Outcome, m chess.Method) string {
	switch o {
	case chess.WhiteWon:
		if m == chess.Checkmate {
			return "checkmate_white"
		}
		return "white_won"
	case chess.BlackWon:
		if m == chess.Checkmate {
			return "checkmate_black"
		}
		return "black_won"
	case chess.Draw:
		switch m {
		case chess.Stalemate:
			return "stalemate"
		default:
			return "draw"
		}
	}
	return "playing"
}

func buildChessPrompt(moves []string, fen string, legalMoves []string) string {
	history := buildHistory(moves)
	legal := strings.Join(legalMoves, ", ")
	humanMove := ""
	if len(moves) > 0 {
		humanMove = moves[len(moves)-1]
	}
	if len(moves) == 1 {
		return fmt.Sprintf(
			"[Xadrez] Nova partida. Você joga com as PRETAS. Minha abertura: %s. FEN: %s. "+
				"Seus lances válidos: %s. "+
				"Escolha UM lance e responda com \"LANCE: [lance]\" (ex: LANCE: e7e5). "+
				"Em seguida, escreva UMA frase de análise. "+
				"Se a posição for muito desfavorável para as brancas, você pode sugerir que o jogador proponha empate ou desista.",
			humanMove, fen, legal,
		)
	}
	return fmt.Sprintf(
		"[Xadrez] Você joga com as PRETAS. Histórico: %s. Minha jogada: %s. FEN: %s. "+
			"Seus lances válidos: %s. "+
			"Escolha UM lance e responda com \"LANCE: [lance]\" (ex: LANCE: e7e5). "+
			"Em seguida, escreva UMA frase de análise. "+
			"Se a posição for muito desfavorável para as brancas, você pode sugerir que o jogador proponha empate ou desista.",
		history, humanMove, fen, legal,
	)
}

func buildHistory(moves []string) string {
	var sb strings.Builder
	for i, m := range moves {
		if i%2 == 0 {
			fmt.Fprintf(&sb, "%d.", i/2+1)
		}
		sb.WriteString(m)
		if i < len(moves)-1 {
			sb.WriteByte(' ')
		}
	}
	return sb.String()
}

var (
	moveRegex    = regexp.MustCompile(`(?i)\*{0,2}lance\*{0,2}\s*:?\*{0,2}\s*([a-h][1-8][a-h][1-8][qrbn]?)`)
	anyMoveRegex = regexp.MustCompile(`\b([a-h][1-8][a-h][1-8][qrbn]?)\b`)
)

func extractChessMove(text, lastHumanMove string, legalMoves []string) string {
	lowerText := strings.ToLower(text)
	lowerHuman := strings.ToLower(lastHumanMove)

	var candidates []string
	if m := moveRegex.FindStringSubmatch(text); len(m) > 1 {
		candidates = append(candidates, strings.ToLower(m[1]))
	}
	for _, m := range anyMoveRegex.FindAllStringSubmatch(text, -1) {
		mv := strings.ToLower(m[1])
		if mv != lowerHuman {
			candidates = append(candidates, mv)
		}
	}

	// 1) Lance exato na lista legal
	for _, mv := range candidates {
		if isInList(mv, legalMoves) {
			return mv
		}
	}
	// 2) Fallback por destino: modelo errou a origem mas acertou a casa destino
	for _, mv := range candidates {
		if len(mv) >= 4 {
			dest := mv[2:4]
			for _, legal := range legalMoves {
				if len(legal) >= 4 && strings.ToLower(legal[2:4]) == dest {
					return strings.ToLower(legal)
				}
			}
		}
	}
	// 3) Qualquer lance legal mencionado no texto
	for _, legal := range legalMoves {
		if strings.Contains(lowerText, strings.ToLower(legal)) {
			return strings.ToLower(legal)
		}
	}
	return ""
}

func extractAnalysis(text string) string {
	lanceRe := regexp.MustCompile(`(?i)lance\s*:`)
	var result []string
	for _, l := range strings.Split(strings.TrimSpace(text), "\n") {
		if !lanceRe.MatchString(l) && strings.TrimSpace(l) != "" {
			result = append(result, strings.TrimSpace(l))
		}
	}
	return strings.Join(result, " ")
}

func isInList(move string, list []string) bool {
	for _, m := range list {
		if strings.ToLower(m) == move {
			return true
		}
	}
	return false
}
