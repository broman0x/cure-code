<script setup>
import { ref, onMounted } from 'vue'
import { useI18n } from '../i18n'

const { t } = useI18n()
const contributors = ref([])

onMounted(async () => {
  try {
    const res = await fetch('https://api.github.com/repos/broman0x/cure-code/contributors')
    contributors.value = await res.json()
  } catch (e) {
    console.error('Failed to fetch contributors')
  }
})
</script>

<template>
  <section class="section container">
    <div class="contributors-header">
      <h2 class="title-text">{{ t('contributors.title') }}</h2>
      <p>{{ t('contributors.subtitle') }}</p>
    </div>
    
    <div class="contributors-list">
      <a v-for="user in contributors" :key="user.id" :href="user.html_url" target="_blank" class="contributor-card">
        <img :src="user.avatar_url" :alt="user.login">
        <span>{{ user.login }}</span>
      </a>
    </div>

    <div class="cta-box">
      <p>{{ t('contributors.cta') }}</p>
      <a href="https://github.com/broman0x/cure-code/blob/main/CONTRIBUTING.md" target="_blank" class="btn btn-outline">
        How to contribute
      </a>
    </div>
  </section>
</template>

<style scoped>
.contributors-header {
  text-align: center;
  margin-bottom: 3rem;
}

h2 { font-size: 2.5rem; margin-bottom: 1rem; }
p { color: var(--text-secondary); }

.contributors-list {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 1.5rem;
  margin-bottom: 4rem;
}

.contributor-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
  text-decoration: none;
  color: var(--text);
  transition: var(--transition);
}

.contributor-card:hover { transform: scale(1.1); }

.contributor-card img {
  width: 64px;
  height: 64px;
  border-radius: 50%;
  border: 2px solid var(--border);
}

.contributor-card span {
  font-size: 0.85rem;
  font-weight: 500;
}

.cta-box {
  background: var(--bg-secondary);
  padding: 3rem;
  border-radius: var(--radius);
  text-align: center;
  border: 1px solid var(--border);
}

.cta-box p {
  margin-bottom: 1.5rem;
  font-size: 1.1rem;
  font-weight: 500;
}

@media (max-width: 768px) {
  h2 { font-size: 2rem; }
  .contributor-card img { width: 48px; height: 48px; }
}
</style>
