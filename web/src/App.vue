<template>
  <v-app :theme="theme">
    <v-navigation-drawer
      v-model="drawer"
      :permanent="!mobile"
      :temporary="mobile"
      :rail="!mobile && rail"
    >
      <v-list-item prepend-icon="mdi-robot" title="AI Agent" nav>
        <template #append>
          <v-btn
            v-if="!mobile"
            :icon="rail ? 'mdi-chevron-right' : 'mdi-chevron-left'"
            variant="text"
            size="small"
            @click="rail = !rail"
          />
        </template>
      </v-list-item>

      <v-divider />

      <v-list density="compact" nav class="mt-1">
        <v-list-item
          v-for="item in navItems"
          :key="item.to"
          :prepend-icon="item.icon"
          :title="item.title"
          :to="item.to"
          exact
          rounded="lg"
        />
      </v-list>

      <template #append>
        <v-divider />
        <v-list density="compact" nav>
          <v-list-item
            :prepend-icon="theme === 'dark' ? 'mdi-weather-sunny' : 'mdi-weather-night'"
            :title="theme === 'dark' ? 'Modo Claro' : 'Modo Escuro'"
            rounded="lg"
            @click="toggleTheme"
          />
        </v-list>
      </template>
    </v-navigation-drawer>

    <v-app-bar v-if="mobile" flat border="b">
      <v-app-bar-nav-icon @click="drawer = !drawer" />
      <v-app-bar-title>
        <v-icon color="primary" class="mr-1">mdi-robot</v-icon>
        AI Agent
      </v-app-bar-title>
      <template #append>
        <v-btn
          :icon="theme === 'dark' ? 'mdi-weather-sunny' : 'mdi-weather-night'"
          variant="text"
          @click="toggleTheme"
        />
      </template>
    </v-app-bar>

    <v-main>
      <router-view />
    </v-main>
  </v-app>
</template>

<script setup>
import { ref } from 'vue'
import { useDisplay } from 'vuetify'

const { mobile } = useDisplay()

const drawer = ref(true)
const rail = ref(false)
const theme = ref('light')

const navItems = [
  { title: 'Chat', icon: 'mdi-chat', to: '/' },
  { title: 'Arquivos', icon: 'mdi-file-multiple', to: '/files' },
  { title: 'Skills', icon: 'mdi-puzzle', to: '/skills' },
  { title: 'Configurações', icon: 'mdi-cog', to: '/settings' },
]

function toggleTheme() {
  theme.value = theme.value === 'light' ? 'dark' : 'light'
}
</script>
