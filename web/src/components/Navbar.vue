<script setup>
import { ref } from 'vue'
import { useI18n } from '../i18n'

const { lang, setLang, t } = useI18n()
const isMenuOpen = ref(false)

const toggleLang = () => {
  setLang(lang.value === 'en' ? 'id' : 'en')
}
</script>

<template>
  <nav class="navbar">
    <div class="container nav-content">
      <router-link to="/" class="logo" @click="isMenuOpen = false">
        <img src="/logo.png" alt="Logo" class="nav-logo">
        CuRe Code
      </router-link>

      <div class="nav-right">
        <div :class="['nav-links', { 'is-active': isMenuOpen }]">
          <router-link to="/" @click="isMenuOpen = false">{{ t('nav.home') }}</router-link>
          <router-link to="/docs" @click="isMenuOpen = false">{{ t('nav.docs') }}</router-link>
          <a href="https://github.com/broman0x/cure-code" target="_blank" class="github-btn">GitHub</a>
        </div>

        <button class="lang-switch" @click="toggleLang">
          {{ lang === 'en' ? 'ID' : 'EN' }}
        </button>

        <button class="menu-toggle" @click="isMenuOpen = !isMenuOpen" :class="{ 'is-active': isMenuOpen }">
          <span></span>
          <span></span>
        </button>
      </div>
    </div>
  </nav>
</template>

<style scoped>
.navbar {
  position: sticky;
  top: 0;
  background: var(--bg);
  border-bottom: 1px solid var(--border);
  z-index: 1000;
  height: 72px;
  display: flex;
  align-items: center;
}

.nav-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.nav-right {
  display: flex;
  align-items: center;
  gap: 1.5rem;
}

.logo {
  font-weight: 900;
  font-size: 1.25rem;
  text-decoration: none;
  color: #000;
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.nav-logo {
  width: 48px;
  height: 48px;
  object-fit: contain;
}

.nav-links {
  display: flex;
  gap: 2rem;
  align-items: center;
}

.nav-links a {
  text-decoration: none;
  font-size: 0.95rem;
  color: var(--text-secondary);
  font-weight: 600;
  transition: var(--transition);
}

.nav-links a:hover, .nav-links a.router-link-active {
  color: var(--primary);
}

.github-btn {
  background: var(--accent);
  color: var(--bg) !important;
  padding: 0.5rem 1rem;
  border-radius: 8px;
  font-size: 0.9rem;
  font-weight: 600;
  text-decoration: none;
}

.lang-switch {
  background: var(--bg-secondary);
  border: 1px solid var(--border);
  padding: 0.3rem 0.6rem;
  border-radius: 6px;
  font-size: 0.75rem;
  font-weight: 700;
  cursor: pointer;
  color: var(--text);
}

.menu-toggle {
  display: none;
  flex-direction: column;
  gap: 6px;
  background: transparent;
  border: none;
  cursor: pointer;
}

.menu-toggle span {
  width: 24px;
  height: 2px;
  background: var(--text);
  transition: var(--transition);
}

@media (max-width: 768px) {
  .menu-toggle { display: flex; }
  
  .nav-links {
    position: fixed;
    top: 72px;
    left: 0;
    width: 100%;
    background: var(--bg);
    flex-direction: column;
    padding: 3rem 1.5rem;
    gap: 2.5rem;
    border-bottom: 1px solid var(--border);
    transform: translateY(-100%);
    opacity: 0;
    transition: var(--transition);
    pointer-events: none;
  }

  .nav-links.is-active {
    transform: translateY(0);
    opacity: 1;
    pointer-events: all;
  }
}
</style>
