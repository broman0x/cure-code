<script setup>
import { useI18n } from '../i18n'
const { t } = useI18n()
</script>

<template>
  <section class="hero animate-fade">
    <div class="container">
      <div class="hero-content">
        <h1 class="hero-title">{{ t('hero.title') }}</h1>
        <p class="hero-subtitle">{{ t('hero.subtitle') }}</p>
        <div class="hero-actions">
          <router-link to="/docs" class="btn btn-primary">{{ t('hero.start') }}</router-link>
          <a href="#install" class="btn btn-outline">{{ t('hero.install') }}</a>
        </div>
      </div>
    </div>
  </section>
</template>

<style scoped>
.hero {
  padding: 8rem 0 4rem;
  text-align: center;
}

.hero-content {
  max-width: 800px;
  margin: 0 auto;
}

.hero-title {
  font-size: 4rem;
  margin-bottom: 1.5rem;
  color: var(--text);
}

.hero-subtitle {
  font-size: 1.25rem;
  color: var(--text-secondary);
  max-width: 600px;
  margin: 0 auto 2.5rem;
  line-height: 1.6;
}

.hero-actions {
  display: flex;
  justify-content: center;
  gap: 1rem;
}

@media (max-width: 768px) {
  .hero {
    padding: 4rem 1.5rem 2rem;
  }
  
  .hero-title {
    font-size: 2.5rem;
  }
  
  .hero-subtitle {
    font-size: 1.1rem;
  }

  .hero-actions {
    flex-direction: column;
    gap: 0.75rem;
  }
  
  .btn {
    width: 100%;
  }
}
</style>

