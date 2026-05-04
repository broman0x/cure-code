<script setup>
import { useI18n } from '../i18n'
const { t } = useI18n()
</script>

<template>
  <section class="hero">
    <div class="container hero-inner">
      <div class="hero-text">
        <div class="hero-eyebrow">Open Source · AI Coding Agent · v1.0.2</div>
        <h1>{{ t('hero.title') }}</h1>
        <p>{{ t('hero.subtitle') }}</p>
        <div class="hero-cta">
          <router-link to="/docs" class="cta-primary">{{ t('hero.start') }}</router-link>
          <a href="#install" class="cta-secondary">{{ t('hero.install') }}</a>
        </div>
      </div>
    </div>
  </section>
</template>

<style scoped>
.hero {
  padding: 72px 0 80px;
  border-bottom: 1px solid var(--line);
}

.hero-inner {
  max-width: 640px;
}

.hero-eyebrow {
  font-size: 0.75rem;
  font-family: var(--mono);
  color: var(--text-3);
  margin-bottom: 20px;
  letter-spacing: 0.02em;
}

h1 {
  font-size: clamp(2rem, 5vw, 3rem);
  font-weight: 700;
  color: var(--text);
  margin-bottom: 16px;
  line-height: 1.15;
}

p {
  font-size: 1rem;
  color: var(--text-2);
  line-height: 1.7;
  max-width: 520px;
  margin-bottom: 32px;
}

.hero-cta {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.cta-primary {
  display: inline-block;
  padding: 9px 20px;
  background: var(--text);
  color: var(--bg);
  font-size: 0.875rem;
  font-weight: 600;
  border-radius: 5px;
  transition: opacity 0.15s;
}

.cta-primary:hover { opacity: 0.85; }

.cta-secondary {
  display: inline-block;
  padding: 9px 20px;
  background: transparent;
  color: var(--text-2);
  font-size: 0.875rem;
  font-weight: 500;
  border: 1px solid var(--line);
  border-radius: 5px;
  transition: color 0.15s, border-color 0.15s;
}

.cta-secondary:hover {
  color: var(--text);
  border-color: var(--text-3);
}
</style>
