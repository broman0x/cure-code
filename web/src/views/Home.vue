<script setup>
import { ref, onMounted } from 'vue'
import { useI18n } from '../i18n'
import Hero from '../components/Hero.vue'
import Terminal from '../components/Terminal.vue'
import Features from '../components/Features.vue'
import Contributors from '../components/Contributors.vue'

const { t } = useI18n()
const os = ref('linux')

const recentSymbols = ref(['Agent', 'SpawnSubAgent', 'saveState', 'checkLoopDetection', 'IntelligenceService'])
const tasks = ref([
  { description: 'Phase 1: Memory Compaction', status: 'completed' },
  { description: 'Phase 2: Code Intelligence', status: 'completed' },
  { description: 'Phase 3: Task Orchestration', status: 'completed' },
  { description: 'Phase 4: Web Integration', status: 'in_progress' },
])

onMounted(() => {
  const p = window.navigator.platform.toLowerCase()
  if (p.includes('win')) os.value = 'windows'
  else if (p.includes('mac')) os.value = 'macos'
})
</script>

<template>
  <div class="home">
    <Hero />
    <section class="demo-section">
      <div class="container">
        <Terminal />
      </div>
    </section>

    <Features />

    <section id="install" class="install-section">
      <div class="container">
        <div class="install-label">{{ t('install.title') }}</div>
        <div class="install-os">{{ os === 'windows' ? 'Windows (PowerShell)' : os === 'macos' ? 'macOS / Linux' : 'Linux / macOS' }}</div>
        <div class="install-cmd">
          <code v-if="os === 'windows'">iex (irm https://raw.githubusercontent.com/broman0x/cure-code/main/install.ps1)</code>
          <code v-else>curl -fsSL https://raw.githubusercontent.com/broman0x/cure-code/main/install.sh | bash</code>
        </div>
      </div>
    </section>

    <Contributors />
  </div>
</template>

<style scoped>
.demo-section {
  padding: 48px 0;
  border-bottom: 1px solid var(--line);
}



.install-section {
  padding: 64px 0;
  border-bottom: 1px solid var(--line);
}

.install-label {
  font-size: 0.72rem;
  font-family: var(--mono);
  color: var(--text-3);
  text-transform: uppercase;
  letter-spacing: 0.06em;
  margin-bottom: 8px;
}

.install-os {
  font-size: 1.1rem;
  font-weight: 600;
  color: var(--text);
  margin-bottom: 20px;
}

.install-cmd {
  display: inline-flex;
  background: var(--surface);
  border: 1px solid var(--line);
  border-radius: 5px;
  padding: 12px 20px;
  max-width: 100%;
  overflow-x: auto;
}

.install-cmd code {
  font-family: var(--mono);
  font-size: 0.875rem;
  color: var(--text);
  white-space: nowrap;
}

@media (max-width: 768px) {
  .demo-grid {
    grid-template-columns: 1fr;
  }
}
</style>
