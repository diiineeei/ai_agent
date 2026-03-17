<template>
  <div class="d-flex flex-column" style="height: calc(100vh - var(--v-layout-top, 0px))">

    <!-- ── Header ──────────────────────────────────────── -->
    <div class="d-flex align-center gap-1 px-4" style="height: 64px; flex-shrink: 0">
      <v-avatar color="primary" size="40" class="mr-1">
        <v-img v-if="currentAgentConfig?.avatar" :src="currentAgentConfig.avatar" cover />
        <v-icon v-else>mdi-robot-happy</v-icon>
      </v-avatar>

      <div class="flex-grow-1 overflow-hidden">
        <div class="text-subtitle-2 font-weight-bold text-truncate" style="line-height:1.2">
          {{ store.sessionName || store.agentName || 'Playground AI' }}
        </div>
        <div class="d-flex align-center gap-1 flex-wrap" style="line-height:1">
          <v-chip v-if="currentAgentConfig?.model" size="x-small" variant="tonal" color="secondary" class="mt-1">
            <v-icon start size="10">mdi-chip</v-icon>
            {{ currentAgentConfig.model }}
          </v-chip>
          <v-chip v-if="currentAgentConfig?.provider === 'ollama'" size="x-small" variant="tonal" color="orange" class="mt-1">
            <v-icon start size="10">mdi-server-outline</v-icon>
            Ollama
          </v-chip>
        </div>
      </div>

      <v-btn icon size="small" variant="text" color="primary" title="Nova sessão" @click="newSessionDialog = true">
        <v-icon size="22">mdi-square-edit-outline</v-icon>
      </v-btn>

      <v-menu location="bottom end" :close-on-content-click="false" min-width="260">
        <template #activator="{ props: menuProps }">
          <v-btn icon size="small" variant="text" v-bind="menuProps">
            <v-icon size="20">mdi-dots-vertical</v-icon>
          </v-btn>
        </template>

        <v-card rounded="lg">
          <v-list density="compact" nav>

            <!-- Sessão -->
            <v-list-subheader>Sessão</v-list-subheader>
            <v-list-item
              prepend-icon="mdi-history"
              title="Carregar histórico"
              rounded="lg"
              :disabled="loadingHistory"
              @click="loadHistory"
            />
            <v-list-item
              prepend-icon="mdi-folder-open-outline"
              title="Sessões"
              rounded="lg"
              @click="openSessionsDialog"
            />
            <v-list-item
              prepend-icon="mdi-plus-circle-outline"
              title="Nova sessão"
              rounded="lg"
              base-color="primary"
              @click="newSessionDialog = true"
            />

            <!-- Agente (só quando há agente selecionado) -->
            <template v-if="currentAgentConfig">
              <v-divider class="my-1" />
              <v-list-subheader>Agente</v-list-subheader>
              <v-list-item
                prepend-icon="mdi-cog-outline"
                title="Editar agente"
                rounded="lg"
                @click="router.push({ name: 'agents', query: { edit: currentAgentConfig.id } })"
              />
              <v-list-item rounded="lg" class="py-1">
                <template #prepend>
                  <v-icon class="mr-3">mdi-puzzle-outline</v-icon>
                </template>
                <v-list-item-title>Skills ativas</v-list-item-title>
                <template #append>
                  <div class="d-flex flex-wrap gap-1 justify-end" style="max-width:140px">
                    <v-chip
                      v-for="s in currentAgentConfig.enabled_skills"
                      :key="s"
                      size="x-small"
                      variant="tonal"
                      color="primary"
                      :prepend-icon="skillIcon(s)"
                    >{{ skillLabel(s) }}</v-chip>
                    <v-chip
                      v-for="mid in currentAgentConfig.mcp_server_ids"
                      :key="mid"
                      size="x-small"
                      variant="tonal"
                      color="secondary"
                      prepend-icon="mdi-connection"
                    >{{ mcpName(mid) }}</v-chip>
                    <span v-if="!currentAgentConfig.enabled_skills?.length && !currentAgentConfig.mcp_server_ids?.length" class="text-caption text-disabled">Nenhuma</span>
                  </div>
                </template>
              </v-list-item>
            </template>

            <!-- Configurações -->
            <v-divider class="my-1" />
            <v-list-subheader>Configurações</v-list-subheader>
            <v-list-item rounded="lg" @click="toggleTts">
              <template #prepend>
                <v-icon class="mr-3">{{ ttsEnabled ? 'mdi-volume-high' : 'mdi-volume-off' }}</v-icon>
              </template>
              <v-list-item-title>Voz</v-list-item-title>
              <template #append>
                <v-switch :model-value="ttsEnabled" color="primary" hide-details density="compact" @click.stop="toggleTts" />
              </template>
            </v-list-item>
            <v-list-item rounded="lg" @click="showTokens = !showTokens">
              <template #prepend>
                <v-icon class="mr-3">mdi-counter</v-icon>
              </template>
              <v-list-item-title>Mostrar tokens</v-list-item-title>
              <template #append>
                <v-switch :model-value="showTokens" color="primary" hide-details density="compact" @click.stop="showTokens = !showTokens" />
              </template>
            </v-list-item>

            <!-- Perigo -->
            <template v-if="store.messages.length > 0">
              <v-divider class="my-1" />
              <v-list-item
                prepend-icon="mdi-delete-outline"
                title="Limpar histórico"
                rounded="lg"
                base-color="error"
                @click="clearDialog = true"
              />
            </template>

          </v-list>
        </v-card>
      </v-menu>
    </div>

    <v-divider />

    <!-- ── Messages ─────────────────────────────────────── -->
    <div ref="messagesEl" class="flex-grow-1 overflow-y-auto messages-area">

      <!-- Empty state -->
      <div
        v-if="store.messages.length === 0 && !store.loading"
        class="d-flex flex-column align-center justify-center h-100 pa-6"
      >
        <!-- Agent already selected -->
        <template v-if="store.agentConfigId">
          <v-avatar color="primary" size="80" class="mb-5" style="opacity:.15">
            <v-icon size="48">mdi-robot-happy</v-icon>
          </v-avatar>
          <p class="text-h6 mb-1 font-weight-regular">Olá! Sou {{ store.agentName }}</p>
          <p class="text-body-2 text-medium-emphasis">Como posso ajudar você hoje?</p>
        </template>

        <!-- No agent selected: show agent picker -->
        <template v-else>
          <v-icon size="48" color="primary" style="opacity:.3" class="mb-4">mdi-robot-happy</v-icon>
          <p class="text-h6 font-weight-regular mb-1">Escolha um agente para começar</p>
          <p class="text-body-2 text-medium-emphasis mb-6">Selecione com qual assistente deseja conversar</p>

          <div class="agent-picker-grid">
            <v-card
              v-for="cfg in agentConfigsStore.configs"
              :key="cfg.id"
              rounded="xl"
              variant="outlined"
              class="agent-pick-card cursor-pointer"
              @click="pickAgent(cfg)"
            >
              <div class="pa-4 d-flex align-center gap-3">
                <v-avatar color="primary" variant="tonal" size="44" class="flex-shrink-0">
                  <v-img v-if="cfg.avatar" :src="cfg.avatar" cover />
                  <span v-else class="text-body-1 font-weight-bold">{{ cfg.name[0].toUpperCase() }}</span>
                </v-avatar>
                <div class="overflow-hidden flex-grow-1">
                  <div class="text-body-2 font-weight-bold text-truncate">{{ cfg.name }}</div>
                  <div class="d-flex align-center gap-1 mt-1 flex-wrap">
                    <v-chip size="x-small" variant="tonal" color="secondary">
                      <v-icon start size="10">mdi-chip</v-icon>
                      {{ cfg.model }}
                    </v-chip>
                    <v-chip v-if="cfg.provider === 'ollama'" size="x-small" variant="tonal" color="orange">
                      <v-icon start size="10">mdi-server-outline</v-icon>
                      Ollama
                    </v-chip>
                  </div>
                </div>
                <v-icon size="18" color="medium-emphasis">mdi-chevron-right</v-icon>
              </div>

              <!-- Skills reveladas no hover -->
              <div class="agent-card-skills px-4 pb-3">
                <v-divider class="mb-2" />
                <div class="d-flex flex-wrap gap-1">
                  <v-chip
                    v-for="s in cfg.enabled_skills"
                    :key="s"
                    size="x-small"
                    variant="tonal"
                    color="primary"
                    :prepend-icon="skillIcon(s)"
                  >{{ skillLabel(s) }}</v-chip>
                  <v-chip
                    v-for="mid in cfg.mcp_server_ids"
                    :key="mid"
                    size="x-small"
                    variant="tonal"
                    color="secondary"
                    prepend-icon="mdi-connection"
                  >{{ mcpName(mid) }}</v-chip>
                  <span v-if="!cfg.enabled_skills?.length && !cfg.mcp_server_ids?.length" class="text-caption text-disabled">Nenhuma skill</span>
                </div>
              </div>
            </v-card>

            <div v-if="!agentConfigsStore.configs.length" class="text-center text-medium-emphasis">
              <p class="text-body-2">Nenhum agente cadastrado.</p>
            </div>
          </div>
        </template>
      </div>

      <template v-for="(msg, i) in store.messages" :key="i">

        <!-- System notification -->
        <div v-if="msg.role === 'system'" class="d-flex justify-center my-3">
          <v-chip size="x-small" variant="tonal" prepend-icon="mdi-paperclip">
            {{ msg.text }}
          </v-chip>
        </div>

        <!-- Chat bubbles -->
        <div
          v-else
          class="d-flex mb-2"
          :class="msg.role === 'user' ? 'justify-end' : 'align-end gap-2'"
        >
          <v-avatar v-if="msg.role !== 'user'" color="primary" size="28">
            <v-icon size="14">mdi-robot</v-icon>
          </v-avatar>

          <!-- User -->
          <div v-if="msg.role === 'user'" class="d-flex flex-column align-end bubble-user">
            <v-sheet color="primary" rounded="xl" class="px-4 py-3 w-100" style="border-bottom-right-radius:4px">
              <p class="text-body-1 text-on-primary mb-0" style="white-space:pre-wrap;line-height:1.55">{{ msg.text }}</p>
            </v-sheet>
            <span v-if="msg.createdAt" class="text-caption text-disabled mt-1 mr-1">{{ formatTime(msg.createdAt) }}</span>
          </div>

          <!-- Model -->
          <div v-else class="d-flex flex-column bubble-model">
            <v-sheet rounded="xl" class="border px-4 py-3">
              <div class="markdown-body text-body-1" v-html="renderMarkdown(msg.text)" />
            </v-sheet>
            <!-- Rating buttons + copy + token info + timestamp -->
            <div class="d-flex align-center gap-1 mt-1 pl-1">
              <v-btn
                icon size="x-small" variant="text"
                @click="copyMessage(msg.text, i)"
              >
                <v-icon size="15">{{ copiedIdx === i ? 'mdi-check' : 'mdi-content-copy' }}</v-icon>
                <v-tooltip activator="parent" location="top">Copiar resposta</v-tooltip>
              </v-btn>
              <v-btn
                icon size="x-small" variant="text"
                :color="ratings[modelSeqOf(i)] === 'up' ? 'success' : undefined"
                :style="ratings[modelSeqOf(i)] && ratings[modelSeqOf(i)] !== 'up' ? 'opacity:.3' : ''"
                @click="rateMessage(i, 'up')"
              >
                <v-icon size="15">{{ ratings[modelSeqOf(i)] === 'up' ? 'mdi-thumb-up' : 'mdi-thumb-up-outline' }}</v-icon>
              </v-btn>
              <v-btn
                icon size="x-small" variant="text"
                :color="ratings[modelSeqOf(i)] === 'down' ? 'error' : undefined"
                :style="ratings[modelSeqOf(i)] && ratings[modelSeqOf(i)] !== 'down' ? 'opacity:.3' : ''"
                @click="rateMessage(i, 'down')"
              >
                <v-icon size="15">{{ ratings[modelSeqOf(i)] === 'down' ? 'mdi-thumb-down' : 'mdi-thumb-down-outline' }}</v-icon>
              </v-btn>
              <span v-if="msg.createdAt" class="text-caption text-disabled ml-1">{{ formatTime(msg.createdAt) }}</span>
              <v-chip
                v-if="showTokens && msg.tokenUsage"
                size="x-small" variant="text"
                class="token-chip text-disabled ml-1"
              >
                <v-icon start size="11">mdi-lightning-bolt</v-icon>
                {{ msg.tokenUsage.total_tokens.toLocaleString() }} tokens
                <v-tooltip activator="parent" location="top">
                  Entrada: {{ msg.tokenUsage.prompt_tokens.toLocaleString() }} &nbsp;·&nbsp;
                  Saída: {{ msg.tokenUsage.response_tokens.toLocaleString() }}
                </v-tooltip>
              </v-chip>
            </div>
          </div>
        </div>

      </template>

      <!-- Thinking indicator -->
      <div v-if="store.loading" class="d-flex align-end gap-2 mb-2">
        <v-avatar color="primary" size="28" class="thinking-avatar">
          <v-icon size="14">mdi-robot</v-icon>
        </v-avatar>
        <v-sheet rounded="xl" class="border px-4 py-3" style="border-bottom-left-radius:4px">
          <div class="d-flex align-center gap-2">
            <div class="typing-dots"><span /><span /><span /></div>
            <transition name="thinking-msg" mode="out-in">
              <span :key="thinkingText" class="text-caption text-medium-emphasis">{{ thinkingText }}</span>
            </transition>
          </div>
        </v-sheet>
      </div>

    </div>

    <!-- ── Chess panel ──────────────────────────────────── -->
    <ChessGame
      v-if="chessOpen"
      ref="chessRef"
      :configs="agentConfigsStore.configs.filter(c => c.enabled_skills?.includes('chess'))"
      :initial-agent-id="store.agentConfigId"
      :initial-agent-name="store.agentName"
      @close="chessOpen = false"
      @ai-move="onChessAiMove"
      @game-start="onChessGameStart"
      @game-end="onChessGameEnd"
    />

    <!-- Error -->
    <v-alert
      v-if="store.error"
      type="error" variant="tonal" density="compact"
      class="mx-4 mb-2" closable
      @click:close="store.error = null"
    >{{ store.error }}</v-alert>

    <v-divider />

    <!-- ── Input ──────────────────────────────────────── -->
    <div v-if="store.agentConfigId" class="pa-4" style="flex-shrink:0">

      <!-- Upload status -->
      <v-slide-y-reverse-transition>
        <div v-if="uploadStatus" class="mb-2">
          <v-chip
            size="small" variant="tonal"
            :color="uploadStatus.state === 'success' ? 'success' : uploadStatus.state === 'error' ? 'error' : undefined"
            :prepend-icon="uploadStatus.state === 'success' ? 'mdi-check-circle' : uploadStatus.state === 'error' ? 'mdi-alert-circle' : undefined"
          >
            <template v-if="uploadStatus.state === 'uploading'" #prepend>
              <v-progress-circular size="13" width="2" indeterminate class="mr-1" />
            </template>
            {{ uploadStatus.name }}
          </v-chip>
        </div>
      </v-slide-y-reverse-transition>

      <!-- Sugestões de perguntas -->
      <v-slide-y-reverse-transition>
        <div v-if="(loadingSuggestions || suggestions.length > 0) && !store.loading" class="mb-2 suggestions-row">
          <!-- Skeletons enquanto carrega -->
          <template v-if="loadingSuggestions">
            <v-skeleton-loader
              v-for="n in 3" :key="n"
              type="chip"
              class="flex-shrink-0"
              style="width:120px"
            />
          </template>
          <!-- Chips com as sugestões -->
          <template v-else>
            <v-chip
              v-for="(q, idx) in suggestions"
              :key="idx"
              size="small"
              variant="tonal"
              color="primary"
              class="suggestion-chip flex-shrink-0"
              @click="applySuggestion(q)"
            >{{ q }}</v-chip>
            <v-btn icon size="x-small" variant="text" class="flex-shrink-0 ml-1" @click="suggestions = []">
              <v-icon size="14">mdi-close</v-icon>
              <v-tooltip activator="parent" location="top">Fechar sugestões</v-tooltip>
            </v-btn>
          </template>
        </div>
      </v-slide-y-reverse-transition>

      <!-- Chess mode badge -->
      <v-slide-y-reverse-transition>
        <div v-if="chessOpen && chessRef?.agentId" class="mb-2 d-flex align-center gap-1">
          <v-chip size="small" color="primary" variant="tonal" prepend-icon="mdi-chess-knight">
            Conversando sobre o jogo
          </v-chip>
        </div>
      </v-slide-y-reverse-transition>

      <!-- Input card -->
      <input ref="fileInput" type="file" accept=".txt,.pdf" style="display:none" @change="onFileSelected" />
      <v-card
        variant="outlined"
        rounded="xl"
        :color="inputFocused ? 'primary' : undefined"
        flat
      >
        <v-textarea
          v-model="input"
          :placeholder="listening ? 'Ouvindo…' : (chessOpen && chessRef?.agentId ? 'Pergunte sobre o jogo…' : 'Digite sua mensagem…')"
          variant="plain"
          rows="1"
          auto-grow
          max-rows="6"
          hide-details
          density="compact"
          class="px-2 pt-2 pb-0 messages-input"
          @keydown.enter.exact.prevent="sendMessage"
          @keydown.shift.enter.exact="input += '\n'"
          @focus="inputFocused = true"
          @blur="inputFocused = false"
        />

        <div class="d-flex align-center px-2 pb-2 pt-1">
          <v-btn
            icon size="small" variant="text"
            :disabled="uploading"
            @click="triggerFileInput"
          >
            <v-icon size="20">mdi-paperclip</v-icon>
            <v-tooltip activator="parent" location="top">Anexar arquivo (.txt, .pdf)</v-tooltip>
          </v-btn>
          <v-btn
            v-if="currentAgentConfig?.enabled_skills?.includes('chess')"
            icon size="small" variant="text"
            title="Jogar xadrez"
            @click="chessOpen = !chessOpen"
          >
            <v-icon size="20" :color="chessOpen ? 'primary' : undefined">mdi-chess-knight</v-icon>
          </v-btn>
          <v-btn
            v-if="speechSupported"
            icon size="small" variant="text"
            :color="listening ? 'error' : holdRecording ? 'warning' : undefined"
            :class="{ 'mic-pulse': listening || holdRecording }"
            @pointerdown.prevent="onMicDown"
            @pointerup="onMicUp"
            @pointercancel="onMicCancel"
            @contextmenu.prevent
          >
            <v-icon size="20">
              {{ listening ? 'mdi-stop' : holdRecording ? 'mdi-microphone' : 'mdi-microphone-outline' }}
            </v-icon>
            <v-tooltip activator="parent" location="top">
              {{ listening ? 'Clique para parar' : 'Clique para gravar · Segure para enviar automaticamente' }}
            </v-tooltip>
          </v-btn>
          <v-spacer />
          <span class="text-caption text-disabled mr-3 d-none d-sm-inline">Shift+Enter = nova linha</span>
          <v-btn
            icon color="primary" size="small" variant="flat" rounded="lg"
            :loading="store.loading"
            :disabled="!input.trim()"
            @click="sendMessage"
          >
            <v-icon size="18">mdi-send</v-icon>
          </v-btn>
        </div>
      </v-card>
    </div>

  </div>

  <!-- ── Sessions dialog ─────────────────────────────── -->
  <v-dialog v-model="sessionDialog" max-width="560" scrollable>
    <v-card rounded="lg">
      <v-card-title class="pt-4 d-flex align-center">
        Sessões
        <v-spacer />
        <v-btn size="small" variant="tonal" color="primary" prepend-icon="mdi-plus"
          @click="sessionDialog = false; newSessionDialog = true">
          Nova sessão
        </v-btn>
      </v-card-title>
      <v-divider />
      <v-card-text class="pa-0" style="max-height:420px">
        <v-progress-linear v-if="loadingSessions" indeterminate />
        <v-list v-if="sessions.length > 0" lines="two">
          <v-list-item
            v-for="s in sessions" :key="s.session_id"
            :active="s.session_id === store.sessionId"
            color="primary" rounded="lg" class="mx-2 my-1"
            @click="selectSession(s)"
          >
            <template #title>
              <!-- Inline rename -->
              <div class="d-flex align-center" @click.stop>
                <template v-if="renamingId === s.session_id">
                  <v-text-field
                    v-model="renameValue"
                    density="compact"
                    variant="underlined"
                    hide-details
                    autofocus
                    class="flex-grow-1"
                    style="min-width:0"
                    @keydown.enter="submitRename(s)"
                    @keydown.esc="renamingId = null"
                    @click.stop
                  />
                  <v-btn icon size="x-small" variant="text" color="primary" class="ml-1" @click.stop="submitRename(s)">
                    <v-icon size="16">mdi-check</v-icon>
                  </v-btn>
                  <v-btn icon size="x-small" variant="text" class="ml-1" @click.stop="renamingId = null">
                    <v-icon size="16">mdi-close</v-icon>
                  </v-btn>
                </template>
                <template v-else>
                  <span class="text-body-2 font-weight-medium text-truncate" @click="selectSession(s)">
                    {{ s.name || s.session_id }}
                  </span>
                  <v-btn
                    icon size="x-small" variant="text"
                    class="ml-1 flex-shrink-0"
                    style="opacity:.5"
                    @click.stop="startRename(s)"
                  >
                    <v-icon size="14">mdi-pencil-outline</v-icon>
                  </v-btn>
                </template>
              </div>
            </template>
            <template #subtitle>
              <span v-if="s.name" class="text-caption" style="font-family:monospace">{{ s.session_id }}</span>
              {{ getAgentName(s.agent_config_id) }}
              <span class="mx-1 text-disabled">·</span>{{ s.message_count }} msg
              <template v-if="s.updated_at">
                <span class="mx-1 text-disabled">·</span>{{ formatDate(s.updated_at) }}
              </template>
            </template>
            <template #append>
              <v-icon v-if="s.session_id === store.sessionId" size="16" color="primary">mdi-check-circle</v-icon>
            </template>
          </v-list-item>
        </v-list>
        <div v-else-if="!loadingSessions" class="text-center text-medium-emphasis py-10">
          <v-icon size="40" style="opacity:.25">mdi-chat-outline</v-icon>
          <p class="mt-2 text-body-2">Nenhuma sessão anterior.</p>
        </div>
      </v-card-text>
      <v-divider />
      <v-card-text class="pt-3 pb-2">
        <v-row no-gutters align="center" class="gap-2">
          <v-col>
            <v-text-field v-model="manualSessionId" label="Ou digitar um ID de sessão"
              variant="outlined" density="compact" hide-details @keydown.enter="applyManualSession" />
          </v-col>
          <v-col cols="auto">
            <v-btn color="primary" variant="tonal" :disabled="!manualSessionId.trim()" @click="applyManualSession">Ir</v-btn>
          </v-col>
        </v-row>
      </v-card-text>
      <v-card-actions class="pt-0">
        <v-spacer />
        <v-btn @click="sessionDialog = false">Fechar</v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>

  <!-- ── New session dialog ──────────────────────────── -->
  <v-dialog v-model="newSessionDialog" max-width="480" scrollable>
    <v-card rounded="lg">
      <v-card-title class="pt-4 d-flex align-center">
        Nova sessão
        <v-spacer />
        <v-btn icon size="small" variant="text" @click="newSessionDialog = false">
          <v-icon size="18">mdi-close</v-icon>
        </v-btn>
      </v-card-title>
      <v-divider />
      <v-card-text class="pa-3" style="max-height: 420px">
        <p class="text-caption text-medium-emphasis mb-3 px-1">Escolha o agente para iniciar a conversa:</p>
        <div class="agent-pick-grid">
          <v-card
            v-for="cfg in agentConfigsStore.configs"
            :key="cfg.id"
            rounded="xl"
            variant="outlined"
            class="agent-pick-card cursor-pointer"
            @click="quickStartSession(cfg)"
          >
            <div class="pa-3 d-flex align-center gap-3">
              <v-avatar color="primary" variant="tonal" size="40" class="flex-shrink-0">
                <v-img v-if="cfg.avatar" :src="cfg.avatar" cover />
                <span v-else class="text-body-2 font-weight-bold">{{ cfg.name[0].toUpperCase() }}</span>
              </v-avatar>
              <div class="overflow-hidden flex-grow-1">
                <div class="text-body-2 font-weight-bold text-truncate">{{ cfg.name }}</div>
                <div class="d-flex align-center gap-1 mt-1 flex-wrap">
                  <v-chip size="x-small" variant="tonal" color="secondary">
                    <v-icon start size="10">mdi-chip</v-icon>{{ cfg.model }}
                  </v-chip>
                  <v-chip v-if="cfg.provider === 'ollama'" size="x-small" variant="tonal" color="orange">
                    <v-icon start size="10">mdi-server-outline</v-icon>Ollama
                  </v-chip>
                </div>
              </div>
              <v-icon size="16" color="medium-emphasis">mdi-chevron-right</v-icon>
            </div>
          </v-card>
          <div v-if="!agentConfigsStore.configs.length" class="text-center text-medium-emphasis py-6">
            <v-icon size="36" style="opacity:.2">mdi-robot-outline</v-icon>
            <p class="text-body-2 mt-2">Nenhum agente cadastrado.</p>
          </div>
        </div>
      </v-card-text>
    </v-card>
  </v-dialog>



  <!-- ── Clear confirm dialog ────────────────────────── -->
  <v-dialog v-model="clearDialog" max-width="380">
    <v-card rounded="lg">
      <v-card-title class="pt-4">Limpar histórico</v-card-title>
      <v-card-text>Apagará todo o histórico desta sessão no servidor. Tem certeza?</v-card-text>
      <v-card-actions>
        <v-spacer />
        <v-btn @click="clearDialog = false">Cancelar</v-btn>
        <v-btn color="error" @click="doClear">Limpar</v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup>
import { ref, computed, watch, nextTick, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useChatStore } from '@/stores/chat'
import { useAgentConfigsStore } from '@/stores/agent_configs'
import { useMcpServersStore } from '@/stores/mcp_servers'
import { chatAPI, filesAPI, feedbackAPI, suggestAPI } from '@/services/api'
import { renderMarkdown } from '@/utils/markdown'
import ChessGame from '@/components/ChessGame.vue'

const router = useRouter()
const route  = useRoute()
const store = useChatStore()
const agentConfigsStore = useAgentConfigsStore()
const mcpStore = useMcpServersStore()

const currentAgentConfig = computed(() =>
  agentConfigsStore.configs.find((c) => c.id === store.agentConfigId) ?? null
)

const mcpName = (id) => mcpStore.servers.find((s) => s.id === id)?.name ?? id.slice(0, 8) + '…'

const SKILL_META = {
  weather:           { label: 'Clima',              icon: 'mdi-weather-partly-cloudy' },
  search_documents:  { label: 'Busca em Documentos', icon: 'mdi-text-search' },
  suggest_questions: { label: 'Sugestões',           icon: 'mdi-help-circle-outline' },
  chess:             { label: 'Xadrez',              icon: 'mdi-chess-knight' },
}
const skillLabel = (name) => SKILL_META[name]?.label ?? name
const skillIcon  = (name) => SKILL_META[name]?.icon  ?? 'mdi-puzzle-outline'

const chessOpen    = ref(false)
const chessRef     = ref(null)

// Fecha o xadrez se o agente atual não tiver a skill chess
watch(currentAgentConfig, (cfg) => {
  if (chessOpen.value && !cfg?.enabled_skills?.includes('chess')) {
    chessOpen.value = false
  }
})

const input        = ref('')
const inputFocused = ref(false)
const messagesEl   = ref(null)
const fileInput    = ref(null)
const loadingHistory = ref(false)
const uploading    = ref(false)
const uploadStatus = ref(null)

// ratings: { [messageIndex]: 'up' | 'down' }
const ratings    = ref({})
const showTokens = ref(false)
const copiedIdx  = ref(null)

const THINKING_MESSAGES = [
  'Pensando…',
  'Analisando sua pergunta…',
  'Verificando ferramentas disponíveis…',
  'Consultando base de conhecimento…',
  'Processando contexto…',
  'Elaborando resposta…',
  'Revisando informações…',
]
const thinkingText  = ref(THINKING_MESSAGES[0])
let thinkingTimer   = null
let thinkingIdx     = 0

watch(() => store.loading, (loading) => {
  if (loading) {
    thinkingIdx = 0
    thinkingText.value = THINKING_MESSAGES[0]
    thinkingTimer = setInterval(() => {
      thinkingIdx = (thinkingIdx + 1) % THINKING_MESSAGES.length
      thinkingText.value = THINKING_MESSAGES[thinkingIdx]
    }, 2000)
  } else {
    clearInterval(thinkingTimer)
    thinkingText.value = THINKING_MESSAGES[0]
  }
})

// ── Voice ──────────────────────────────────────────────
const listening       = ref(false)   // toggle mode: recording, click to stop
const holdRecording   = ref(false)   // hold mode: recording while held
const ttsEnabled      = ref(false)
const speechSupported = !!(window.SpeechRecognition || window.webkitSpeechRecognition)

let currentAudio = null

let recognition = null
let holdTimer    = null
let isHoldMode   = false
let finalTranscript = ''
const HOLD_MS = 350

if (speechSupported) {
  const SR = window.SpeechRecognition || window.webkitSpeechRecognition
  recognition = new SR()
  recognition.lang = 'pt-BR'
  recognition.continuous = false
  recognition.interimResults = true

  recognition.onresult = (e) => {
    let interim = ''
    finalTranscript = ''
    for (let i = e.resultIndex; i < e.results.length; i++) {
      const t = e.results[i][0].transcript
      if (e.results[i].isFinal) finalTranscript += t
      else interim = t
    }
    input.value = finalTranscript || interim
  }

  recognition.onend = () => {
    if (isHoldMode) {
      holdRecording.value = false
      if (finalTranscript.trim()) sendMessage()
    } else {
      listening.value = false
    }
    isHoldMode = false
  }

  recognition.onerror = () => {
    listening.value = false
    holdRecording.value = false
    isHoldMode = false
  }
}

function onMicDown() {
  if (!recognition) return
  isHoldMode = false
  holdTimer = setTimeout(() => {
    isHoldMode = true
    holdRecording.value = true
    finalTranscript = ''
    try { recognition.start() } catch {}
  }, HOLD_MS)
}

function onMicUp() {
  clearTimeout(holdTimer)
  if (isHoldMode) {
    // hold mode: stop and auto-send (handled in onend)
    recognition.stop()
  } else {
    // quick tap: toggle listening
    if (listening.value) {
      recognition.stop()
    } else {
      listening.value = true
      finalTranscript = ''
      try { recognition.start() } catch {}
    }
  }
}

function onMicCancel() {
  clearTimeout(holdTimer)
  if (holdRecording.value || listening.value) recognition?.stop()
  holdRecording.value = false
  listening.value = false
  isHoldMode = false
}

const suggestions        = ref([])
const loadingSuggestions = ref(false)

const clearDialog      = ref(false)
const sessionDialog    = ref(false)
const newSessionDialog = ref(false)
const manualSessionId  = ref('')
const selectedConfigId = ref(null)
const newSessionName   = ref('')
const sessions         = ref([])
const loadingSessions  = ref(false)

// inline rename
const renamingId  = ref(null)
const renameValue = ref('')

onMounted(async () => {
  await Promise.all([agentConfigsStore.fetchAll(), mcpStore.fetchAll()])
  if (route.query.agent) {
    const cfg = agentConfigsStore.configs.find((c) => c.id === route.query.agent)
    if (cfg) pickAgent(cfg)
  }
  selectedConfigId.value = store.agentConfigId
  scrollToBottom()
})

watch(() => store.messages.length, () => {
  scrollToBottom()
  const msgs = store.messages
  if (msgs.length > 0 && msgs[msgs.length - 1].role === 'model') {
    fetchSuggestions()
    speak(msgs[msgs.length - 1].text)
  }
})
watch(() => store.sessionId, () => {
  ratings.value = {}
  suggestions.value = []
  loadingSuggestions.value = false
  if (currentAudio) { currentAudio.pause(); currentAudio = null }
  onMicCancel()
})

// ── Messaging ──────────────────────────────────────────
async function sendMessage() {
  const text = input.value.trim()
  if (!text || store.loading) return
  input.value = ''

  // Se o xadrez está aberto e com agente selecionado, roteia para a sessão de xadrez
  if (chessOpen.value && chessRef.value?.agentId) {
    const chess = chessRef.value
    const gameCtx = chess.moves?.length
      ? `[Xadrez | FEN: ${chess.currentFEN} | Lances: ${chess.moves.join(' ')}] `
      : '[Xadrez - nova partida] '
    store.messages.push({ role: 'user', text, createdAt: new Date().toISOString() })
    store.loading = true
    try {
      const { data } = await chatAPI.sendPrompt(chess.chessSessionId, gameCtx + text, chess.agentId, [])
      store.messages.push({ role: 'model', text: data.response, tokenUsage: data.token_usage ?? null, createdAt: new Date().toISOString() })
    } catch (e) {
      store.messages.pop()
      store.error = e.response?.data?.error || e.message
    } finally {
      store.loading = false
    }
    return
  }

  await store.send(text)
}

function onChessAiMove({ move, analysis }) {
  if (!analysis) return
  store.messages.push({
    role: 'model',
    text: `♟ ${move} — ${analysis}`,
    createdAt: new Date().toISOString(),
  })
}

async function onChessGameStart({ agentId, agentColor, playerColor }) {
  if (!chessRef.value) return
  const chess = chessRef.value
  const agentSide  = agentColor  === 'black' ? 'pretas' : 'brancas'
  const playerSide = playerColor === 'white'  ? 'brancas' : 'pretas'
  store.loading = true
  try {
    const { data } = await chatAPI.sendPrompt(
      chess.chessSessionId,
      `[Xadrez] Nova partida iniciada. Você joga com as ${agentSide} e o jogador humano joga com as ${playerSide} (ele move primeiro). Apresente-se brevemente, confirme qual cor é a sua e deseje boa sorte ao jogador.`,
      agentId,
      [],
    )
    store.messages.push({ role: 'model', text: data.response, createdAt: new Date().toISOString() })
  } catch { /* silencia */ } finally {
    store.loading = false
  }
}

async function onChessGameEnd({ status, moves, fen }) {
  if (!chessRef.value) return
  const chess = chessRef.value
  const statusMsg = {
    checkmate_white: 'Xeque-mate — brancas venceram',
    checkmate_black: 'Xeque-mate — pretas venceram',
    stalemate: 'Afogamento — empate',
    draw: 'Empate aceito pelo jogador',
    checkmate_black_resign: 'Jogador desistiu',
  }[status] ?? 'Fim de jogo'

  const playedMoves = moves?.length ? moves : []
  const moveList = playedMoves.length ? playedMoves.join(' ') : 'nenhum lance foi realizado'

  const analysisInstruction = playedMoves.length >= 4
    ? `Comente brevemente o resultado e destaque no máximo duas jogadas interessantes dentre as seguintes (e APENAS essas): ${moveList}.`
    : `Comente brevemente o resultado. Não analise jogadas pois a partida foi muito curta (${playedMoves.length} lance(s)).`

  store.loading = true
  try {
    const { data } = await chatAPI.sendPrompt(
      chess.chessSessionId,
      `[Xadrez] Fim de partida. Resultado: ${statusMsg}. Lances jogados: ${moveList}. FEN final: ${fen}. ${analysisInstruction} Sugira ao jogador uma nova partida.`,
      chess.agentId,
      [],
    )
    store.messages.push({ role: 'model', text: data.response, createdAt: new Date().toISOString() })
  } catch { /* silencia */ } finally {
    store.loading = false
  }
}

async function loadHistory() {
  loadingHistory.value = true
  await store.loadHistory()
  await loadRatings()
  scrollToBottom()
  loadingHistory.value = false
}

async function doClear() {
  await store.clearHistory()
  ratings.value = {}
  suggestions.value = []
  clearDialog.value = false
}

// ── File upload ────────────────────────────────────────
function triggerFileInput() { fileInput.value?.click() }

async function onFileSelected(e) {
  const file = e.target.files?.[0]
  if (!file) return
  e.target.value = ''
  uploading.value  = true
  uploadStatus.value = { state: 'uploading', name: file.name }
  try {
    await filesAPI.upload(file)
    uploadStatus.value = { state: 'success', name: file.name }
    store.messages.push({ role: 'system', text: `${file.name} enviado com sucesso` })
    setTimeout(() => { uploadStatus.value = null }, 3000)
  } catch {
    uploadStatus.value = { state: 'error', name: `Erro ao enviar ${file.name}` }
    setTimeout(() => { uploadStatus.value = null }, 4000)
  } finally {
    uploading.value = false
  }
}

// ── Sessions ───────────────────────────────────────────
async function openSessionsDialog() {
  renamingId.value = null
  sessionDialog.value = true
  manualSessionId.value = ''
  loadingSessions.value = true
  try {
    const { data } = await chatAPI.listSessions()
    sessions.value = data ?? []
  } catch { sessions.value = [] }
  finally { loadingSessions.value = false }
}

function getAgentName(id) {
  if (!id) return '—'
  return agentConfigsStore.configs.find((c) => c.id === id)?.name ?? id
}

function formatTime(dateStr) {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  if (isNaN(d.getTime())) return ''
  return d.toLocaleTimeString('pt-BR', { hour: '2-digit', minute: '2-digit' })
}

function formatDate(dateStr) {
  const d = new Date(dateStr)
  if (isNaN(d.getTime()) || d.getFullYear() < 2000) return ''
  return d.toLocaleString('pt-BR', { dateStyle: 'short', timeStyle: 'short' })
}

function selectSession(s) {
  store.setSession(s.session_id, s.name || null)
  store.agentConfigId = s.agent_config_id || null
  store.agentName     = getAgentName(s.agent_config_id)
  sessionDialog.value = false
  loadHistory() // loadHistory já chama loadRatings internamente
}

function applyManualSession() {
  const id = manualSessionId.value.trim()
  if (!id) return
  store.setSession(id)
  store.agentConfigId = null
  store.agentName     = null
  sessionDialog.value = false
}

function startNewSession() {
  const cfg = agentConfigsStore.configs.find((c) => c.id === selectedConfigId.value)
  store.newSession(selectedConfigId.value, newSessionName.value.trim() || null)
  if (cfg) store.agentName = cfg.name
  newSessionDialog.value = false
  newSessionName.value = ''
}

function quickStartSession(cfg) {
  store.newSession(cfg.id)
  store.agentName = cfg.name
  newSessionDialog.value = false
}

function pickAgent(cfg) {
  store.newSession(cfg.id)
  store.agentName = cfg.name
}

// ── Inline rename ──────────────────────────────────────
function startRename(s) {
  renamingId.value  = s.session_id
  renameValue.value = s.name || ''
}

async function submitRename(s) {
  const name = renameValue.value.trim()
  try {
    await chatAPI.renameSession(s.session_id, name)
    s.name = name
    // sync store if it's the active session
    if (s.session_id === store.sessionId) store.sessionName = name || null
  } catch { /* ignore */ }
  renamingId.value = null
}

// ── Feedback ───────────────────────────────────────────
// Use the sequential index among model-only messages as a stable key
// (immune to interleaved system messages from file uploads).
function modelSeqOf(flatIdx) {
  let n = -1
  for (let j = 0; j <= flatIdx; j++) {
    if (store.messages[j]?.role === 'model') n++
  }
  return n
}

async function copyMessage(text, idx) {
  await navigator.clipboard.writeText(text)
  copiedIdx.value = idx
  setTimeout(() => { copiedIdx.value = null }, 2000)
}

async function rateMessage(flatIdx, rating) {
  const seq = modelSeqOf(flatIdx)
  if (ratings.value[seq] === rating) return
  ratings.value[seq] = rating
  feedbackAPI.submit(store.sessionId, seq, store.agentConfigId ?? '', rating).catch(() => {})
}

async function loadRatings() {
  try {
    const { data } = await feedbackAPI.forSession(store.sessionId)
    const map = {}
    for (const f of (data ?? [])) map[f.message_index] = f.rating
    ratings.value = map
  } catch { /* non-critical */ }
}

function scrollToBottom() {
  nextTick(() => {
    if (messagesEl.value) messagesEl.value.scrollTop = messagesEl.value.scrollHeight
  })
}

function toggleTts() {
  ttsEnabled.value = !ttsEnabled.value
  if (!ttsEnabled.value && currentAudio) { currentAudio.pause(); currentAudio = null }
}

async function speak(text) {
  if (!ttsEnabled.value || !text) return
  if (currentAudio) { currentAudio.pause(); currentAudio = null }

  try {
    const resp = await fetch('/api/tts', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ text }),
    })
    if (!resp.ok) return
    const blob = await resp.blob()
    const url = URL.createObjectURL(blob)
    currentAudio = new Audio(url)
    currentAudio.onended = () => { URL.revokeObjectURL(url); currentAudio = null }
    currentAudio.play()
  } catch { /* silencioso */ }
}

// ── Sugestões de perguntas ──────────────────────────────
async function fetchSuggestions() {
  if (!store.sessionId) return
  suggestions.value = []
  loadingSuggestions.value = true
  try {
    const { data } = await suggestAPI.getQuestions(store.sessionId)
    suggestions.value = data?.questions ?? []
  } catch { /* non-critical */ }
  finally { loadingSuggestions.value = false }
}

function applySuggestion(question) {
  input.value = question
  suggestions.value = []
  nextTick(() => {
    const textarea = document.querySelector('.messages-input textarea')
    textarea?.focus()
  })
}
</script>

<style scoped>
.messages-area {
  padding: 24px 20px 12px;
  scroll-behavior: smooth;
}

/* Bubbles */
.bubble-user  { border-bottom-right-radius: 4px !important; max-width: 74%; }
.bubble-model { border-bottom-left-radius:  4px !important; max-width: 74%; }

/* Typing animation */
.typing-dots { display: flex; gap: 5px; align-items: center; height: 18px; }
.typing-dots span {
  width: 7px; height: 7px;
  border-radius: 50%;
  background: currentColor;
  opacity: .5;
  animation: bounce 1.4s infinite;
}
.typing-dots span:nth-child(2) { animation-delay: .2s; }
.typing-dots span:nth-child(3) { animation-delay: .4s; }
@keyframes bounce {
  0%, 80%, 100% { transform: scale(.8); opacity: .4; }
  40%           { transform: scale(1.2); opacity: 1;  }
}

/* Thinking message transition */
.thinking-msg-enter-active, .thinking-msg-leave-active { transition: opacity .3s, transform .3s; }
.thinking-msg-enter-from { opacity: 0; transform: translateY(4px); }
.thinking-msg-leave-to   { opacity: 0; transform: translateY(-4px); }

/* Thinking avatar pulse */
@keyframes thinking-pulse {
  0%, 100% { box-shadow: 0 0 0 0 rgba(var(--v-theme-primary), .4); }
  50%       { box-shadow: 0 0 0 6px rgba(var(--v-theme-primary), 0); }
}
.thinking-avatar { animation: thinking-pulse 1.4s ease-in-out infinite; }

/* Thinking text fade */
@keyframes thinking-fade {
  0%, 100% { opacity: .4; }
  50%       { opacity: 1; }
}
.thinking-text { animation: thinking-fade 1.4s ease-in-out infinite; }

/* Mic pulse animation */
@keyframes mic-pulse {
  0%, 100% { box-shadow: 0 0 0 0 rgba(var(--v-theme-error), .4); }
  50%       { box-shadow: 0 0 0 6px rgba(var(--v-theme-error), 0); }
}
.mic-pulse { animation: mic-pulse 1.2s ease-in-out infinite; border-radius: 50%; }

/* Token chip */
.token-chip { font-size: 10px !important; opacity: .7; }

/* Agent picker grid (empty state) */
.agent-picker-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
  gap: 12px;
  width: 100%;
  max-width: 680px;
}

/* Agent pick grid (new session dialog) */
.agent-pick-grid {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.agent-pick-card { transition: all .15s ease; }
.agent-pick-card:hover { transform: translateY(-2px); box-shadow: 0 4px 16px rgba(0,0,0,.1); }

.agent-card-skills {
  overflow: hidden;
  max-height: 0;
  opacity: 0;
  transition: max-height .2s ease, opacity .2s ease;
}
.agent-pick-card:hover .agent-card-skills {
  max-height: 80px;
  opacity: 1;
}

/* Suggestion chips row */
.suggestions-row {
  display: flex;
  align-items: center;
  gap: 6px;
  overflow-x: auto;
  scrollbar-width: none;
  padding-bottom: 2px;
}
.suggestions-row::-webkit-scrollbar { display: none; }
.suggestion-chip { cursor: pointer; white-space: nowrap; flex-shrink: 0; }

/* Markdown inside model bubble */
:deep(.markdown-body) {
  line-height: 1.6;
  > *:first-child { margin-top: 0; }
  > *:last-child  { margin-bottom: 0; }
  p  { margin: .45em 0; }
  h1,h2,h3,h4,h5,h6 { margin: .8em 0 .3em; font-weight: 600; line-height: 1.3; }
  h1 { font-size: 1.35em; } h2 { font-size: 1.15em; } h3 { font-size: 1.02em; }
  ul, ol { margin: .4em 0; padding-left: 1.5em; }
  li { margin: .15em 0; }
  code {
    font-family: 'Roboto Mono', monospace;
    font-size: .875em;
    padding: .15em .4em;
    border-radius: 4px;
    background: rgba(var(--v-theme-on-surface), 0.08);
  }
  pre {
    margin: .6em 0; padding: .75em 1em;
    border-radius: 10px;
    background: rgba(var(--v-theme-on-surface), 0.06);
    overflow-x: auto;
    code { padding: 0; background: none; font-size: .85em; }
  }
  blockquote {
    margin: .5em 0; padding: .3em .75em;
    border-left: 3px solid currentColor; opacity: .75;
  }
  table {
    border-collapse: collapse; margin: .5em 0; width: 100%; font-size: .9em;
    th, td {
      border: 1px solid rgba(var(--v-theme-on-surface), 0.15);
      padding: .3em .6em;
    }
    th { font-weight: 600; background: rgba(var(--v-theme-on-surface), 0.06); }
  }
  hr { border: none; border-top: 1px solid rgba(var(--v-theme-on-surface), 0.15); margin: .8em 0; }
  a  { color: inherit; text-decoration: underline; opacity: .8; }
  a:hover { opacity: 1; }
}
</style>
