import { defineStore } from 'pinia'
import { ref } from 'vue'
import { skillsAPI } from '@/services/api'

export const useSkillsStore = defineStore('skills', () => {
  const skills = ref([])
  const loading = ref(false)
  const error = ref(null)

  async function fetchSkills() {
    loading.value = true
    error.value = null
    try {
      const { data } = await skillsAPI.list()
      skills.value = data || []
    } catch (e) {
      error.value = e.response?.data?.error || e.message
    } finally {
      loading.value = false
    }
  }

  async function toggleSkill(name) {
    error.value = null
    try {
      const { data } = await skillsAPI.toggle(name)
      const idx = skills.value.findIndex((s) => s.name === name)
      if (idx !== -1) skills.value[idx] = data
    } catch (e) {
      error.value = e.response?.data?.error || e.message
      throw e
    }
  }

  return { skills, loading, error, fetchSkills, toggleSkill }
})
