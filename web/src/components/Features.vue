<script setup>
import { useI18n } from '../i18n'
const { t } = useI18n()

const features = [
  { key: 'loop', num: '01' },
  { key: 'edits', num: '02' },
  { key: 'context', num: '03' },
  { key: 'safe', num: '04' },
  { key: 'plug', num: '05' },
  { key: 'binary', num: '06' },
]
</script>

<template>
  <section id="features" class="features">
    <div class="container">
      <div class="feat-header">
        <span class="feat-label">Capabilities</span>
      </div>
      <div class="feat-list">
        <div v-for="f in features" :key="f.key" class="feat-row">
          <span class="feat-num">{{ f.num }}</span>
          <div class="feat-body">
            <h3>{{ t(`features.${f.key}.title`) }}</h3>
            <p>{{ t(`features.${f.key}.desc`) }}</p>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>

<style scoped>
.features {
  padding: 80px 0;
  border-bottom: 1px solid var(--line);
}

.feat-header {
  margin-bottom: 40px;
}

.feat-label {
  font-size: 0.72rem;
  font-family: var(--mono);
  color: var(--text-3);
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.feat-list {
  display: flex;
  flex-direction: column;
}

.feat-row {
  display: grid;
  grid-template-columns: 48px 1fr;
  gap: 24px;
  padding: 24px 0;
  border-top: 1px solid var(--line);
}

.feat-num {
  font-size: 0.7rem;
  font-family: var(--mono);
  color: var(--text-3);
  padding-top: 3px;
}

.feat-body h3 {
  font-size: 0.95rem;
  font-weight: 600;
  color: var(--text);
  margin-bottom: 6px;
}

.feat-body p {
  font-size: 0.875rem;
  color: var(--text-2);
  line-height: 1.65;
  max-width: 560px;
}
</style>
