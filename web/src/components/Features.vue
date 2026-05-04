<script setup>
import { useI18n } from '../i18n'
const { t } = useI18n()

const features = [
  { key: 'loop' },
  { key: 'edits' },
  { key: 'context' },
  { key: 'safe' },
  { key: 'plug' },
  { key: 'binary' }
]
</script>

<template>
  <section id="features" class="section container">
    <div class="features-grid">
      <div v-for="feat in features" :key="feat.key" class="feat-card">
        <h3>{{ t(`features.${feat.key}.title`) }}</h3>
        <p>{{ t(`features.${feat.key}.desc`) }}</p>
      </div>
    </div>
  </section>
</template>

<style scoped>
.features-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 4rem 2rem;
  border-top: 1px solid var(--border);
  padding-top: 4rem;
}

.feat-card h3 {
  font-size: 1.1rem;
  margin-bottom: 0.75rem;
  font-weight: 700;
  color: var(--text);
}

.feat-card p {
  font-size: 1rem;
  color: var(--text-secondary);
  line-height: 1.6;
}

@media (max-width: 768px) {
  .features-grid {
    padding: 3rem 1.5rem 0;
    gap: 3rem;
  }
}
</style>

