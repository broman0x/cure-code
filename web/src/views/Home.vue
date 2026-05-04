<script setup>
import { ref, onMounted } from 'vue'
import { useI18n } from '../i18n'
import Hero from '../components/Hero.vue'
import Terminal from '../components/Terminal.vue'
import Features from '../components/Features.vue'
import Contributors from '../components/Contributors.vue'

const { t } = useI18n()
const os = ref('linux')

onMounted(() => {
  const platform = window.navigator.platform.toLowerCase()
  if (platform.includes('win')) {
    os.value = 'windows'
  } else if (platform.includes('mac')) {
    os.value = 'macos'
  } else {
    os.value = 'linux'
  }
})
</script>

<template>
  <div class="home-page">
    <Hero />
    <Terminal />
    <Features />
    
    <section id="install" class="section container">
      <div class="install-box">
        <template v-if="os === 'windows'">
          <h2>{{ t('install.title') }} (Windows)</h2>
          <p>{{ t('install.win') }}</p>
          <div class="cmd-line">
            <code>.\curecode.exe --install</code>
          </div>
          <p class="mt-4">
            <a href="https://github.com/broman0x/cure-code/releases/latest" target="_blank" class="text-link">
              {{ t('install.win_link') }}
            </a>
          </p>
        </template>
        <template v-else>
          <h2>{{ t('install.title') }} ({{ os === 'macos' ? 'macOS' : 'Linux' }})</h2>
          <p>{{ t('install.subtitle') }}</p>
          <div class="cmd-line">
            <code>curl -L https://github.com/broman0x/cure-code/releases/latest/download/curecode-{{ os === 'macos' ? 'darwin' : 'linux' }}-amd64 -o curecode && chmod +x curecode && ./curecode --install</code>
          </div>
        </template>
      </div>
    </section>

    <Contributors />
  </div>
</template>

<style scoped>
.install-box {
  background: var(--bg-secondary);
  border: 1px solid var(--border);
  padding: 4rem 2rem;
  border-radius: var(--radius);
  text-align: center;
}

h2 {
  font-size: 2rem;
  margin-bottom: 1rem;
}

p {
  color: var(--text-secondary);
  margin-bottom: 2rem;
}

.cmd-line {
  background: var(--bg);
  padding: 1rem 1.5rem;
  border: 1px solid var(--border);
  border-radius: var(--radius);
  display: inline-block;
  font-family: 'JetBrains Mono', monospace;
  font-size: 0.95rem;
  max-width: 100%;
  overflow-x: auto;
}

.cmd-line code {
  color: var(--primary);
}

@media (max-width: 768px) {
  .cmd-line {
    font-size: 0.8rem;
    padding: 1rem;
    width: 100%;
  }
}
</style>
