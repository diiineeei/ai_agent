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

          <!-- Sidebar: histórico + comentário + ações -->
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

            <v-spacer />

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
              {{ PIECES[turnColor + p] }}
            </span>
          </div>
        </v-card-text>
      </v-card>
    </v-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useChatStore } from '@/stores/chat'

const props = defineProps({
  configs:          { type: Array,  default: () => [] },
  initialAgentId:   { type: String, default: null },
  initialAgentName: { type: String, default: null },
})
const emit = defineEmits(['close'])

const chatStore = useChatStore()

// ── Constants ──────────────────────────────────────────────
const FILES = ['a','b','c','d','e','f','g','h']

const PIECES = {
  'wK': '♔', 'wQ': '♕', 'wR': '♖', 'wB': '♗', 'wN': '♘', 'wP': '♙',
  'bK': '♚', 'bQ': '♛', 'bR': '♜', 'bB': '♝', 'bN': '♞', 'bP': '♟',
}

const INIT_BOARD = () => [
  ['bR','bN','bB','bQ','bK','bB','bN','bR'],
  ['bP','bP','bP','bP','bP','bP','bP','bP'],
  [null,null,null,null,null,null,null,null],
  [null,null,null,null,null,null,null,null],
  [null,null,null,null,null,null,null,null],
  [null,null,null,null,null,null,null,null],
  ['wP','wP','wP','wP','wP','wP','wP','wP'],
  ['wR','wN','wB','wQ','wK','wB','wN','wR'],
]

// ── State ──────────────────────────────────────────────────
const board       = ref(INIT_BOARD())
const turn        = ref('w')          // 'w' or 'b'
const selected    = ref(null)         // [row, col]
const validMoves  = ref([])           // [{r,c}]
const lastMove    = ref(null)         // [fr,fc,tr,tc]
const moves       = ref([])           // move history in algebraic notation
const gameStatus  = ref('idle')       // 'idle' | 'playing' | 'check' | 'checkmate' | 'stalemate' | 'draw'
const aiThinking  = ref(false)
const drawLoading = ref(false)
const error       = ref(null)
const agentId     = ref(props.initialAgentId ?? null)
const agentName   = ref(props.initialAgentName ?? null)
const promoDialog = ref(false)
const promoData   = ref(null)         // {fr,fc,tr,tc} pending promotion

// Castling rights: [wK, wQ, bK, bQ] (can still castle each side)
const castleRights = ref({ wK: true, wQ: true, bK: true, bQ: true })
// En passant target square [r,c] or null
const enPassant = ref(null)

const turnColor = computed(() => turn.value === 'w' ? 'w' : 'b')

const statusLabel = computed(() => ({
  playing: 'Em jogo',
  check: 'Xeque!',
  checkmate: turn.value === 'w' ? 'Xeque-mate! Pretas vencem' : 'Xeque-mate! Brancas vencem',
  stalemate: 'Afogamento — Empate',
  draw: 'Empate',
})[gameStatus.value] ?? '')

const statusColor = computed(() => ({
  check: 'warning',
  checkmate: 'error',
  stalemate: 'info',
  draw: 'info',
})[gameStatus.value] ?? 'success')

// Se já há agente na prop, inicia o jogo ao montar
onMounted(() => {
  if (agentId.value) resetGame()
})

// ── Game management ────────────────────────────────────────
function startGame(cfg) {
  agentId.value   = cfg.id
  agentName.value = cfg.name
  resetGame()
}

function resetGame() {
  board.value        = INIT_BOARD()
  turn.value         = 'w'
  selected.value     = null
  validMoves.value   = []
  lastMove.value     = null
  moves.value        = []
  gameStatus.value   = agentId.value ? 'playing' : 'idle'
  aiThinking.value   = false
  error.value        = null
  castleRights.value = { wK: true, wQ: true, bK: true, bQ: true }
  enPassant.value    = null
}

// ── Board helpers ──────────────────────────────────────────
function inBounds(r, c) { return r >= 0 && r < 8 && c >= 0 && c < 8 }

function pieceAt(r, c, b = board.value) {
  if (!inBounds(r, c)) return null
  return b[r][c]
}

function colorOf(piece) { return piece ? piece[0] : null }

function isValidTarget(r, c) {
  return validMoves.value.some(m => m.r === r && m.c === c)
}

function findKing(color, b = board.value) {
  for (let r = 0; r < 8; r++)
    for (let c = 0; c < 8; c++)
      if (b[r][c] === color + 'K') return [r, c]
  return null
}

function cloneBoard(b) { return b.map(row => [...row]) }

function applyMove(b, fr, fc, tr, tc, promo = null) {
  const nb = cloneBoard(b)
  const piece = nb[fr][fc]
  nb[tr][tc] = promo ? (piece[0] + promo) : piece
  nb[fr][fc] = null
  // En passant capture
  if (piece[1] === 'P' && fc !== tc && !nb[tr][tc - (tc - fc)]) {
    // This shouldn't occur since we already set nb[tr][tc], handle separately
  }
  return nb
}

function isSquareAttacked(r, c, byColor, b = board.value) {
  const opp = byColor
  for (let sr = 0; sr < 8; sr++) {
    for (let sc = 0; sc < 8; sc++) {
      const p = b[sr][sc]
      if (!p || p[0] !== opp) continue
      const raw = rawMoves(sr, sc, b, null)
      if (raw.some(m => m.r === r && m.c === c)) return true
    }
  }
  return false
}

function isInCheck(color, b = board.value) {
  const king = findKing(color, b)
  if (!king) return false
  return isSquareAttacked(king[0], king[1], color === 'w' ? 'b' : 'w', b)
}

// ── Move generation ────────────────────────────────────────
// rawMoves: moves without check validation (used for attack detection)
function rawMoves(r, c, b, ep) {
  const piece = b[r][c]
  if (!piece) return []
  const color = piece[0]
  const type  = piece[1]
  const moves = []

  const add = (tr, tc) => {
    if (inBounds(tr, tc) && colorOf(b[tr][tc]) !== color)
      moves.push({ r: tr, c: tc })
  }
  const slide = (dr, dc) => {
    let tr = r + dr, tc = c + dc
    while (inBounds(tr, tc)) {
      if (b[tr][tc]) { add(tr, tc); break }
      moves.push({ r: tr, c: tc })
      tr += dr; tc += dc
    }
  }

  if (type === 'P') {
    const dir = color === 'w' ? -1 : 1
    const startRow = color === 'w' ? 6 : 1
    // Forward
    if (inBounds(r + dir, c) && !b[r + dir][c]) {
      moves.push({ r: r + dir, c })
      if (r === startRow && !b[r + 2 * dir][c])
        moves.push({ r: r + 2 * dir, c })
    }
    // Captures
    for (const dc of [-1, 1]) {
      const tr = r + dir, tc = c + dc
      if (inBounds(tr, tc) && b[tr][tc] && colorOf(b[tr][tc]) !== color)
        moves.push({ r: tr, c: tc })
      // En passant
      if (ep && ep[0] === tr && ep[1] === tc)
        moves.push({ r: tr, c: tc })
    }
  } else if (type === 'N') {
    for (const [dr, dc] of [[-2,-1],[-2,1],[-1,-2],[-1,2],[1,-2],[1,2],[2,-1],[2,1]])
      add(r + dr, c + dc)
  } else if (type === 'B') {
    for (const [dr, dc] of [[-1,-1],[-1,1],[1,-1],[1,1]]) slide(dr, dc)
  } else if (type === 'R') {
    for (const [dr, dc] of [[-1,0],[1,0],[0,-1],[0,1]]) slide(dr, dc)
  } else if (type === 'Q') {
    for (const [dr, dc] of [[-1,-1],[-1,1],[1,-1],[1,1],[-1,0],[1,0],[0,-1],[0,1]]) slide(dr, dc)
  } else if (type === 'K') {
    for (const [dr, dc] of [[-1,-1],[-1,0],[-1,1],[0,-1],[0,1],[1,-1],[1,0],[1,1]])
      add(r + dr, c + dc)
  }
  return moves
}

// legalMoves: filter rawMoves by check constraint, add castling
function legalMoves(r, c) {
  const piece = board.value[r][c]
  if (!piece) return []
  const color = piece[0]
  const raw = rawMoves(r, c, board.value, enPassant.value)
  const legal = raw.filter(({ r: tr, c: tc }) => {
    const nb = cloneBoard(board.value)
    // En passant capture
    if (piece[1] === 'P' && tc !== c && !nb[tr][tc]) {
      nb[r][tc] = null // captured pawn
    }
    nb[tr][tc] = nb[r][c]
    nb[r][c] = null
    return !isInCheck(color, nb)
  })

  // Castling
  if (piece[1] === 'K' && !isInCheck(color)) {
    const row = color === 'w' ? 7 : 0
    if (r === row && c === 4) {
      // Kingside
      if (castleRights.value[color + 'K'] &&
          !board.value[row][5] && !board.value[row][6] &&
          !isSquareAttacked(row, 5, color === 'w' ? 'b' : 'w') &&
          !isSquareAttacked(row, 6, color === 'w' ? 'b' : 'w')) {
        legal.push({ r: row, c: 6, castle: 'K' })
      }
      // Queenside
      if (castleRights.value[color + 'Q'] &&
          !board.value[row][3] && !board.value[row][2] && !board.value[row][1] &&
          !isSquareAttacked(row, 3, color === 'w' ? 'b' : 'w') &&
          !isSquareAttacked(row, 2, color === 'w' ? 'b' : 'w')) {
        legal.push({ r: row, c: 2, castle: 'Q' })
      }
    }
  }
  return legal
}

// ── User interaction ───────────────────────────────────────
function onCellClick(r, c) {
  if (turn.value !== 'w' || aiThinking.value || gameStatus.value === 'checkmate' || gameStatus.value === 'stalemate') return

  const cell  = board.value[r][c]
  const color = colorOf(cell)

  // If a piece is already selected
  if (selected.value) {
    const [sr, sc] = selected.value
    const move = validMoves.value.find(m => m.r === r && m.c === c)

    if (move) {
      // Pawn promotion check
      const piece = board.value[sr][sc]
      if (piece === 'wP' && r === 0) {
        promoData.value = { fr: sr, fc: sc, tr: r, tc: c, move }
        promoDialog.value = true
        return
      }
      executeMove(sr, sc, r, c, null, move)
      return
    }

    // Clicked another friendly piece → reselect
    if (color === 'w') {
      selected.value  = [r, c]
      validMoves.value = legalMoves(r, c)
      return
    }

    // Clicked invalid target → deselect
    selected.value  = null
    validMoves.value = []
    return
  }

  // Nothing selected — select own piece
  if (color === 'w') {
    selected.value  = [r, c]
    validMoves.value = legalMoves(r, c)
  }
}

function confirmPromotion(type) {
  promoDialog.value = false
  const { fr, fc, tr, tc, move } = promoData.value
  executeMove(fr, fc, tr, tc, type, move)
  promoData.value = null
}

function squareName(r, c) {
  return FILES[c] + (8 - r)
}

function executeMove(fr, fc, tr, tc, promo, moveObj) {
  const nb     = cloneBoard(board.value)
  const piece  = nb[fr][fc]
  const captured = nb[tr][tc]
  const newEP  = ref(null)

  // Pawn: 2-square advance → set en passant
  if (piece[1] === 'P' && Math.abs(tr - fr) === 2) {
    newEP.value = [(fr + tr) / 2, fc]
  }

  // En passant capture
  if (piece[1] === 'P' && fc !== tc && !nb[tr][tc]) {
    nb[fr][tc] = null
  }

  nb[tr][tc] = promo ? (piece[0] + promo) : piece
  nb[fr][fc] = null

  // Castling rook move
  if (moveObj?.castle) {
    const row = tr
    if (moveObj.castle === 'K') { nb[row][5] = nb[row][7]; nb[row][7] = null }
    else                        { nb[row][3] = nb[row][0]; nb[row][0] = null }
  }

  // Update castling rights
  if (piece === 'wK') { castleRights.value.wK = false; castleRights.value.wQ = false }
  if (piece === 'bK') { castleRights.value.bK = false; castleRights.value.bQ = false }
  if (fr === 7 && fc === 7) castleRights.value.wK = false
  if (fr === 7 && fc === 0) castleRights.value.wQ = false
  if (fr === 0 && fc === 7) castleRights.value.bK = false
  if (fr === 0 && fc === 0) castleRights.value.bQ = false

  board.value      = nb
  enPassant.value  = newEP.value
  lastMove.value   = [fr, fc, tr, tc]
  selected.value   = null
  validMoves.value = []

  // Record move
  const notation = squareName(fr, fc) + squareName(tr, tc) + (promo ? promo.toLowerCase() : '')
  moves.value.push(notation)

  // Switch turn
  turn.value = 'b'
  updateGameStatus()
  if (gameStatus.value === 'playing' || gameStatus.value === 'check') {
    askAI(notation)
  }
}

function updateGameStatus() {
  const color = turn.value
  const inCheck = isInCheck(color)
  let hasLegal = false
  outer: for (let r = 0; r < 8; r++) {
    for (let c = 0; c < 8; c++) {
      if (colorOf(board.value[r][c]) === color && legalMoves(r, c).length > 0) {
        hasLegal = true
        break outer
      }
    }
  }
  if (!hasLegal) {
    gameStatus.value = inCheck ? 'checkmate' : 'stalemate'
  } else {
    gameStatus.value = inCheck ? 'check' : 'playing'
  }
}

// ── Resign / Draw ─────────────────────────────────────────
function resign() {
  gameStatus.value = 'checkmate' // encerra o jogo como derrota
  chatStore.send('[Xadrez] Desisto. Você venceu!')
}

async function offerDraw() {
  drawLoading.value = true
  try {
    await chatStore.send('[Xadrez] Ofereço empate. Você aceita?')
    const last = [...chatStore.messages].reverse().find(m => m.role === 'model')
    const text = (last?.text ?? '').toLowerCase()
    const accepted = /aceito|empate|draw|concordo|sim/.test(text)
    if (accepted) {
      gameStatus.value = 'draw'
    }
  } finally {
    drawLoading.value = false
  }
}

// ── AI move ───────────────────────────────────────────────
async function askAI(userMove) {
  aiThinking.value = true
  error.value      = null

  const isFirst = moves.value.length === 1
  const fen     = boardToFen()
  const formatNote = 'OBRIGATÓRIO: primeira linha da resposta deve ser exatamente "LANCE: [origem][destino]" (exemplo: LANCE: e7e5). Sem esse formato o lance será ignorado. Após o lance, 1 frase curta de análise.'
  const prompt  = isFirst
    ? `[Xadrez] Nova partida. Você joga com as pretas, eu com as brancas. Minha abertura: ${userMove}. FEN: ${fen}. ${formatNote}`
    : `[Xadrez] Minha jogada: ${userMove}. FEN: ${fen}. ${formatNote}`

  try {
    await chatStore.send(prompt)

    // Pega a última mensagem do modelo no chat
    const msgs     = chatStore.messages
    const lastModel = [...msgs].reverse().find(m => m.role === 'model')
    const text      = lastModel?.text ?? ''
    const match     = text.match(/LANCE:\s*([a-h][1-8][a-h][1-8][qrbn]?)/i)

    if (!match) {
      error.value = `Lance não encontrado na resposta. Verifique o chat.`
      turn.value  = 'w'
      return
    }

    applyAIMove(match[1].toLowerCase())
  } catch (e) {
    error.value = 'Erro: ' + (chatStore.error ?? e.message)
    turn.value  = 'w'
  } finally {
    aiThinking.value = false
  }
}

function applyAIMove(moveStr) {
  // Parse "e7e5" or "e7e5q"
  const fc = FILES.indexOf(moveStr[0])
  const fr = 8 - parseInt(moveStr[1])
  const tc = FILES.indexOf(moveStr[2])
  const tr = 8 - parseInt(moveStr[3])
  const promo = moveStr[4] ? moveStr[4].toUpperCase() : null

  if (!inBounds(fr, fc) || !inBounds(tr, tc)) {
    error.value = `Lance inválido recebido: ${moveStr}`
    turn.value  = 'w'
    return
  }

  const piece = board.value[fr][fc]
  if (!piece || piece[0] !== 'b') {
    error.value = `Lance inválido: nenhuma peça preta em ${moveStr.slice(0,2)}`
    turn.value  = 'w'
    return
  }

  const nb = cloneBoard(board.value)

  // En passant capture
  if (piece[1] === 'P' && fc !== tc && !nb[tr][tc]) {
    nb[fr][tc] = null
  }

  // Castling
  if (piece[1] === 'K' && Math.abs(tc - fc) === 2) {
    const row = fr
    if (tc === 6) { nb[row][5] = nb[row][7]; nb[row][7] = null }
    else          { nb[row][3] = nb[row][0]; nb[row][0] = null }
  }

  // Pawn EP target
  let newEP = null
  if (piece[1] === 'P' && Math.abs(tr - fr) === 2) {
    newEP = [(fr + tr) / 2, fc]
  }

  nb[tr][tc] = promo ? ('b' + promo) : piece
  nb[fr][fc] = null

  // Update castling rights
  if (piece === 'bK') { castleRights.value.bK = false; castleRights.value.bQ = false }
  if (fr === 0 && fc === 7) castleRights.value.bK = false
  if (fr === 0 && fc === 0) castleRights.value.bQ = false

  board.value     = nb
  enPassant.value = newEP
  lastMove.value  = [fr, fc, tr, tc]
  moves.value.push(moveStr)

  turn.value = 'w'
  updateGameStatus()
}

// ── FEN generation ─────────────────────────────────────────
function boardToFen() {
  const PIECE_FEN = {
    'wP':'P','wN':'N','wB':'B','wR':'R','wQ':'Q','wK':'K',
    'bP':'p','bN':'n','bB':'b','bR':'r','bQ':'q','bK':'k',
  }
  const rows = board.value.map(row => {
    let s = '', empty = 0
    for (const cell of row) {
      if (!cell) { empty++ }
      else { if (empty) { s += empty; empty = 0 } s += PIECE_FEN[cell] }
    }
    if (empty) s += empty
    return s
  })
  const active    = turn.value
  const castle    = [
    castleRights.value.wK ? 'K' : '',
    castleRights.value.wQ ? 'Q' : '',
    castleRights.value.bK ? 'k' : '',
    castleRights.value.bQ ? 'q' : '',
  ].join('') || '-'
  const ep = enPassant.value ? (FILES[enPassant.value[1]] + (8 - enPassant.value[0])) : '-'
  return `${rows.join('/')} ${active} ${castle} ${ep} 0 ${Math.floor(moves.value.length / 2) + 1}`
}
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

.ai-comment {
  display: flex;
  align-items: flex-start;
  background: rgba(var(--v-theme-primary), 0.06);
  border-radius: 6px;
  padding: 6px 8px;
  line-height: 1.4;
}

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
