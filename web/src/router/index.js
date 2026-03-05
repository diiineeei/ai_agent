import { createRouter, createWebHistory } from 'vue-router'
import ChatView from '@/views/ChatView.vue'
import FilesView from '@/views/FilesView.vue'
import SkillsView from '@/views/SkillsView.vue'
import AgentsView from '@/views/AgentsView.vue'

export default createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', name: 'chat', component: ChatView },
    { path: '/files', name: 'files', component: FilesView },
    { path: '/skills', name: 'skills', component: SkillsView },
    { path: '/agents', name: 'agents', component: AgentsView },
  ],
})
