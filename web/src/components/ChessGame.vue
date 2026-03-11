<template>
  <!-- Painel inline — renderizado diretamente no chat, sem modal -->
  <div class="chess-panel">

    <!-- Barra de título -->
    <div class="chess-panel-header">
      <v-icon size="18" color="primary" class="mr-2">mdi-chess-knight</v-icon>
      <span class="text-body-2 font-weight-bold flex-grow-1">
        Xadrez<template v-if="agentName"> · vs {{ agentName }}</template>
      </span>
      <v-chip v-if="gameStatus !== 'idle'" size="x-small" :color="statusColor" variant="tonal" class="mr-2">
        {{ statusLabel }}
      </v-chip>
      <v-btn icon size="x-small" variant="text" @click="$emit('close')">
        <v-icon size="16">mdi-close</v-icon>
      </v-btn>
    </div>

    <!-- Conteúdo -->
    <div class="chess-panel-body">

      <!-- Seletor de agente -->
      <template v-if="!agentId">
        <p class="text-caption text-medium-emphasis mb-2">Escolha o agente para jogar contra:</p>
        <div class="agent-list">
          <div
            v-for="cfg in configs"
            :key="cfg.id"
            class="agent-row cursor-pointer"
            @click="startGame(cfg)"
          >
            <v-avatar color="primary" variant="tonal" size="32" class="flex-shrink-0">
              <v-img v-if="cfg.avatar" :src="cfg.avatar" cover />
              <span v-else class="text-caption font-weight-bold">{{ cfg.name[0].toUpperCase() }}</span>
            </v-avatar>
            <div class="overflow-hidden flex-grow-1">
              <div class="text-body-2 font-weight-medium text-truncate">{{ cfg.name }}</div>
              <div class="text-caption text-medium-emphasis text-truncate">{{ cfg.model }}</div>
            </div>
            <v-icon size="16" color="medium-emphasis">mdi-chevron-right</v-icon>
          </div>
          <p v-if="!configs.length" class="text-caption text-disabled">Nenhum agente disponível.</p>
        </div>
      </template>

      <!-- Tabuleiro + sidebar -->
      <template v-else>
        <div class="chess-layout">

          <!-- Tabuleiro -->
          <div class="chess-board-wrap">
            <div class="chess-coords-left">
              <span v-for="r in 8" :key="r">{{ 9 - r }}</span>
            </div>
            <div class="chess-board-inner">
              <div class="chess-board">
                <div v-for="(row, ri) in board" :key="ri" class="chess-row">
                  <div
                    v-for="(cell, ci) in row"
                    :key="ci"
                    class="chess-cell"
                    :class="{
                      'cell-light': (ri + ci) % 2 === 0,
                      'cell-dark':  (ri + ci) % 2 !== 0,
                      'cell-selected':  selected && selected[0] === ri && selected[1] === ci,
                      'cell-valid':     isValidTarget(ri, ci),
                      'cell-last-from': lastMove && lastMove[0] === ri && lastMove[1] === ci,
                      'cell-last-to':   lastMove && lastMove[2] === ri && lastMove[3] === ci,
                    }"
                    @click="onCellClick(ri, ci)"
                  >
                    <span v-if="cell" class="chess-piece" :class="cell[0] === 'w' ? 'piece-white' : 'piece-black'">
                      {{ PIECES[cell] }}
                    </span>
                    <div v-else-if="isValidTarget(ri, ci)" class="valid-dot" />
                  </div>
                </div>
              </div>
              <div class="chess-coords-bottom">
                <span v-for="f in FILES" :key="f">{{ f }}</span>
              </div>
            </div>
          </div>

          <!-- Sidebar: histórico + ações -->
          <div class="chess-sidebar">

            <!-- Indicador de turno -->
            <div class="turn-indicator mb-2">
              <v-progress-circular v-if="aiThinking" size="12" width="2" indeterminate color="primary" class="mr-1" />
              <v-icon v-else size="12" :color="turn === 'w' ? 'grey-lighten-2' : 'grey-darken-4'" class="mr-1">mdi-circle</v-icon>
              <span class="text-caption text-medium-emphasis">
                {{ aiThinking ? agentName + ' pensando…' : (turn === 'w' ? 'Sua vez' : agentName + ' jogando…') }}
              </span>
            </div>

            <!-- Histórico de lances -->
            <div class="moves-scroll">
              <template v-for="(m, i) in moves" :key="i">
                <span v-if="i % 2 === 0" class="move-number">{{ Math.floor(i / 2) + 1 }}.</span>
                <span class="move-entry" :class="{ 'move-last': i === moves.length - 1 }">{{ m }}</span>
              </template>
              <span v-if="!moves.length" class="text-caption text-disabled">Nenhum lance ainda</span>
            </div>

            <!-- Erro -->
            <v-alert v-if="error" type="error" variant="tonal" density="compact" rounded="lg" class="mt-2 text-caption" closable @click:close="error = null">
              {{ error }}
            </v-alert>

            <!-- Ações -->
            <div class="d-flex gap-1 mt-2 flex-wrap">
              <v-btn size="x-small" variant="tonal" prepend-icon="mdi-restart" @click="resetGame">Nova partida</v-btn>
              <v-btn
                v-if="gameStatus === 'playing' || gameStatus === 'check'"
                size="x-small" variant="tonal" color="error"
                prepend-icon="mdi-flag-outline"
                @click="resign"
              >Desistir</v-btn>
              <v-btn
                v-if="gameStatus === 'playing' || gameStatus === 'check'"
                size="x-small" variant="tonal" color="warning"
                prepend-icon="mdi-handshake-outline"
                :loading="drawLoading"
                @click="offerDraw"
              >Empate</v-btn>
              <v-btn size="x-small" variant="text" @click="agentId = null; agentName = null; resetGame()">Trocar agente</v-btn>
            </div>
          </div>
        </div>
      </template>

    </div>

    <!-- Promoção de peão (dentro do root único para evitar fragmento) -->
    <v-dialog v-model="promoDialog" max-width="280" persistent>
      <v-card rounded="lg">
        <v-card-title class="pt-4 text-center text-body-1">Promover peão</v-card-title>
        <v-card-text>
          <div class="d-flex justify-center gap-4">
            <span v-for="p in ['Q','R','B','N']" :key="p" class="promo-piece" @click="confirmPromotion(p)">
              {{ PIECES['w' + p] }}
            </span>
          </div>
        </v-card-text>
      </v-card>
    </v-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { chessAPI } from '@/services/api'

const props = defineProps({
  configs:          { type: Array,  default: () => [] },
  initialAgentId:   { type: String, default: null },
  initialAgentName: { type: String, default: null },
})
const emit = defineEmits(['close', 'ai-move', 'game-start', 'game-end'])

const chessSessionId = ref('chess-' + Math.random().toString(36).substring(2, 10))

// ── Constants ─────────────────────────────────────────────
const FILES = ['a','b','c','d','e','f','g','h']

const PIECES = {
  'wK': '♔', 'wQ': '♕', 'wR': '♖', 'wB': '♗', 'wN': '♘', 'wP': '♙',
  'bK': '♚', 'bQ': '♛', 'bR': '♜', 'bB': '♝', 'bN': '♞', 'bP': '♟',
}

const PIECE_MAP = {
  'P':'wP','N':'wN','B':'wB','R':'wR','Q':'wQ','K':'wK',
  'p':'bP','n':'bN','b':'bB','r':'bR','q':'bQ','k':'bK',
}

// ── State ─────────────────────────────────────────────────
const board              = ref(fenToBoard('rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1'))
const turn               = ref('w')
const selected           = ref(null)      // [row, col]
const validMoves         = ref([])        // [{r,c}] computed from backendLegalMoves
const backendLegalMoves  = ref([])        // ["e2e4", ...] from server
const lastMove           = ref(null)      // [fr,fc,tr,tc]
const moves              = ref([])
const gameStatus         = ref('idle')
const aiThinking         = ref(false)
const drawLoading        = ref(false)
const error              = ref(null)
const agentId            = ref(props.initialAgentId ?? null)
const agentName          = ref(props.initialAgentName ?? null)
const promoDialog        = ref(false)
const promoData          = ref(null)      // { fromSq, toSq, options }
const currentFEN         = ref('rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1')

const isGameOver = computed(() =>
  ['checkmate_white','checkmate_black','stalemate','draw'].includes(gameStatus.value)
)

const statusLabel = computed(() => ({
  playing:          'Em jogo',
  checkmate_white:  'Xeque-mate! Brancas vencem',
  checkmate_black:  'Xeque-mate! Pretas vencem',
  stalemate:        'Afogamento — Empate',
  draw:             'Empate',
})[gameStatus.value] ?? '')

const statusColor = computed(() => ({
  checkmate_white: 'success',
  checkmate_black: 'error',
  stalemate:       'info',
  draw:            'info',
})[gameStatus.value] ?? 'success')

onMounted(() => { if (agentId.value) startGame({ id: agentId.value, name: agentName.value }) })

// ── FEN ───────────────────────────────────────────────────
function fenToBoard(fen) {
  return fen.split(' ')[0].split('/').map(row => {
    const cells = []
    for (const ch of row) {
      if (/\d/.test(ch)) for (let i = 0; i < +ch; i++) cells.push(null)
      else cells.push(PIECE_MAP[ch] ?? null)
    }
    return cells
  })
}

function squareName(r, c) { return FILES[c] + (8 - r) }
function colorOf(piece)   { return piece ? piece[0] : null }
function isValidTarget(r, c) { return validMoves.value.some(m => m.r === r && m.c === c) }

// ── Game management ───────────────────────────────────────
async function startGame(cfg) {
  agentId.value        = cfg.id
  agentName.value      = cfg.name
  chessSessionId.value = 'chess-' + Math.random().toString(36).substring(2, 10)
  board.value          = fenToBoard('rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1')
  turn.value           = 'w'
  selected.value       = null
  validMoves.value     = []
  lastMove.value       = null
  moves.value          = []
  gameStatus.value     = 'idle'
  aiThinking.value     = false
  error.value          = null
  currentFEN.value     = 'rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1'

  try {
    const { data } = await chessAPI.start(chessSessionId.value, cfg.id)
    backendLegalMoves.value = data.legal_moves || []
    gameStatus.value        = 'playing'
    emit('game-start', { agentId: cfg.id, agentName: cfg.name, agentColor: 'black', playerColor: 'white' })
  } catch (e) {
    error.value = 'Erro ao iniciar partida: ' + (e.response?.data?.error || e.message)
  }
}

async function resetGame() {
  if (agentId.value) await startGame({ id: agentId.value, name: agentName.value })
}

// ── User interaction ──────────────────────────────────────
function onCellClick(r, c) {
  if (turn.value !== 'w' || aiThinking.value || isGameOver.value) return

  const sq    = squareName(r, c)
  const cell  = board.value[r][c]
  const color = colorOf(cell)

  if (selected.value) {
    const [sr, sc] = selected.value
    const fromSq   = squareName(sr, sc)
    const move     = validMoves.value.find(m => m.r === r && m.c === c)

    if (move) {
      const promoOptions = backendLegalMoves.value.filter(m => m.startsWith(fromSq + sq) && m.length === 5)
      if (promoOptions.length > 0) {
        // Promoção: mostrar diálogo
        promoData.value  = { fromSq, toSq: sq, options: promoOptions }
        promoDialog.value = true
        return
      }
      sendMove(fromSq + sq)
      return
    }
    if (color === 'w') { selectPiece(r, c); return }
    selected.value  = null
    validMoves.value = []
    return
  }

  if (color === 'w') selectPiece(r, c)
}

function selectPiece(r, c) {
  const sq = squareName(r, c)
  selected.value   = [r, c]
  // Calcula destinos a partir dos lances legais do backend
  const seen = new Set()
  validMoves.value = backendLegalMoves.value
    .filter(m => m.startsWith(sq))
    .map(m => ({ r: 8 - parseInt(m[3]), c: FILES.indexOf(m[2]) }))
    .filter(m => { const k = `${m.r},${m.c}`; if (seen.has(k)) return false; seen.add(k); return true })
}

function confirmPromotion(promoChar) {
  promoDialog.value = false
  const { fromSq, toSq } = promoData.value
  promoData.value = null
  sendMove(fromSq + toSq + promoChar)
}

// ── Send move to backend ──────────────────────────────────
async function sendMove(uci) {
  aiThinking.value = true
  error.value      = null
  selected.value   = null
  validMoves.value = []

  // Aplica lance humano otimisticamente
  applyUCIToBoard(uci)
  moves.value.push(uci)
  turn.value = 'b'

  try {
    const { data } = await chessAPI.move(chessSessionId.value, uci)

    // Sincroniza com estado autoritativo do backend (inclui lance da IA)
    board.value             = fenToBoard(data.fen)
    moves.value             = data.moves
    backendLegalMoves.value = data.legal_moves || []
    currentFEN.value        = data.fen

    if (data.ai_move) {
      const m = data.ai_move
      lastMove.value = [8 - parseInt(m[1]), FILES.indexOf(m[0]), 8 - parseInt(m[3]), FILES.indexOf(m[2])]
      emit('ai-move', { move: data.ai_move, analysis: data.analysis || '' })
    }

    const newStatus = mapStatus(data.status)
    gameStatus.value = newStatus
    turn.value       = 'w'

    if (newStatus !== 'playing') {
      emit('game-end', { status: newStatus, moves: data.moves, fen: data.fen })
    }
  } catch (e) {
    error.value = e.response?.data?.error || e.message
    // Ressincroniza com o backend para desfazer o estado otimista
    syncState()
  } finally {
    aiThinking.value = false
  }
}

async function syncState() {
  try {
    const { data } = await chessAPI.state(chessSessionId.value)
    if (!data) return
    board.value             = fenToBoard(data.fen)
    moves.value             = data.moves || []
    backendLegalMoves.value = data.legal_moves || []
    gameStatus.value        = mapStatus(data.status)
    turn.value              = 'w'
  } catch { /* silencia */ }
}

// Aplica um lance UCI no board local (visual apenas, sem validação)
function applyUCIToBoard(uci) {
  const fc = FILES.indexOf(uci[0]), fr = 8 - parseInt(uci[1])
  const tc = FILES.indexOf(uci[2]), tr = 8 - parseInt(uci[3])
  const promoChar = uci[4]
  const nb    = board.value.map(row => [...row])
  const piece = nb[fr][fc]
  if (!piece) return
  // En passant
  if (piece[1] === 'P' && fc !== tc && !nb[tr][tc]) nb[fr][tc] = null
  // Roque
  if (piece[1] === 'K' && Math.abs(tc - fc) === 2) {
    if (tc === 6) { nb[fr][5] = nb[fr][7]; nb[fr][7] = null }
    else          { nb[fr][3] = nb[fr][0]; nb[fr][0] = null }
  }
  nb[tr][tc] = promoChar ? (piece[0] + promoChar.toUpperCase()) : piece
  nb[fr][fc] = null
  board.value    = nb
  lastMove.value = [fr, fc, tr, tc]
}

function mapStatus(s) {
  if (!s) return 'playing'
  if (s === 'playing') return 'playing'
  return s // checkmate_white, checkmate_black, stalemate, draw
}

// ── Resign / Draw ─────────────────────────────────────────
function resign() {
  gameStatus.value = 'checkmate_black'
  emit('game-end', { status: 'checkmate_black', moves: moves.value, fen: currentFEN.value })
  chessAPI.reset(chessSessionId.value).catch(() => {})
}

function offerDraw() {
  gameStatus.value = 'draw'
  emit('game-end', { status: 'draw', moves: moves.value, fen: currentFEN.value })
  chessAPI.reset(chessSessionId.value).catch(() => {})
}

defineExpose({
  get chessSessionId() { return chessSessionId.value },
  get agentId()        { return agentId.value },
  get currentFEN()     { return currentFEN.value },
  get moves()          { return moves.value },
})
</script>

<style scoped>
/* ── Painel ───────────────────────────────────────────── */
.chess-panel {
  border-top: 1px solid rgba(var(--v-theme-on-surface), 0.12);
  background: rgba(var(--v-theme-surface), 1);
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
}
.chess-panel-header {
  display: flex;
  align-items: center;
  padding: 8px 12px;
  border-bottom: 1px solid rgba(var(--v-theme-on-surface), 0.08);
}
.chess-panel-body {
  padding: 12px;
  overflow-y: auto;
  max-height: 420px;
}

/* ── Layout: tabuleiro + sidebar ─────────────────────── */
.chess-layout {
  display: flex;
  gap: 12px;
  align-items: flex-start;
}
.chess-board-wrap {
  display: flex;
  gap: 4px;
  flex-shrink: 0;
}
.chess-coords-left {
  display: flex;
  flex-direction: column;
  justify-content: space-around;
  font-size: 10px;
  color: rgba(var(--v-theme-on-surface), 0.4);
  padding-bottom: 16px;
  user-select: none;
  width: 12px;
  text-align: center;
}
.chess-board-inner { display: flex; flex-direction: column; }
.chess-board {
  display: grid;
  grid-template-rows: repeat(8, 1fr);
  border: 1px solid rgba(var(--v-theme-on-surface), 0.2);
  border-radius: 3px;
  overflow: hidden;
  aspect-ratio: 1;
  width: min(300px, calc(100vw - 180px));
}
.chess-row { display: grid; grid-template-columns: repeat(8, 1fr); }
.chess-cell {
  aspect-ratio: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: background 0.08s;
}
.cell-light { background: #f0d9b5; }
.cell-dark  { background: #b58863; }
.cell-selected          { background: #7fc97f !important; }
.cell-valid.cell-light  { background: #cde6a0 !important; }
.cell-valid.cell-dark   { background: #8eb866 !important; }
.cell-last-from.cell-light, .cell-last-to.cell-light { background: #cdd16f !important; }
.cell-last-from.cell-dark,  .cell-last-to.cell-dark  { background: #a9a535 !important; }

.chess-piece {
  font-size: clamp(14px, 3vw, 28px);
  line-height: 1;
  user-select: none;
}
.piece-white { color: #fff; text-shadow: 0 0 2px #000, 0 1px 2px rgba(0,0,0,.6); }
.piece-black { color: #111; text-shadow: 0 1px 1px rgba(255,255,255,.15); }

.valid-dot {
  width: 28%;
  height: 28%;
  border-radius: 50%;
  background: rgba(0,0,0,.2);
}
.chess-coords-bottom {
  display: flex;
  justify-content: space-around;
  font-size: 10px;
  color: rgba(var(--v-theme-on-surface), 0.4);
  padding-top: 2px;
  user-select: none;
}

/* ── Sidebar ─────────────────────────────────────────── */
.chess-sidebar {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  min-height: min(300px, calc(100vw - 180px));
}
.turn-indicator { display: flex; align-items: center; }
.moves-scroll {
  flex: 1;
  overflow-y: auto;
  font-family: monospace;
  font-size: 11px;
  line-height: 1.8;
  max-height: 180px;
  background: rgba(var(--v-theme-on-surface), 0.03);
  border-radius: 6px;
  padding: 4px 6px;
}
.move-number { color: rgba(var(--v-theme-on-surface), 0.4); margin-right: 3px; }
.move-entry { margin-right: 6px; }
.move-last { font-weight: 600; color: rgb(var(--v-theme-primary)); }

/* ── Agent picker ────────────────────────────────────── */
.agent-list { display: flex; flex-direction: column; gap: 6px; }
.agent-row {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 10px;
  border-radius: 10px;
  border: 1px solid rgba(var(--v-theme-on-surface), 0.1);
  transition: background 0.12s;
}
.agent-row:hover { background: rgba(var(--v-theme-primary), 0.06); }

/* ── Promoção ────────────────────────────────────────── */
.promo-piece {
  font-size: 38px;
  cursor: pointer;
  padding: 6px;
  border-radius: 8px;
  transition: background 0.12s;
  color: #1a1a1a;
}
.promo-piece:hover { background: rgba(var(--v-theme-primary), 0.12); }
</style>
