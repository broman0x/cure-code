<script setup>
import { ref } from 'vue'
import { useI18n } from '../i18n'

const { lang, setLang, t } = useI18n()
const menuOpen = ref(false)
const toggleLang = () => setLang(lang.value === 'en' ? 'id' : 'en')
</script>

<template>
  <nav class="nav">
    <div class="container nav-inner">
      <router-link to="/" class="nav-brand" @click="menuOpen = false">
        <img src="/logo.png" alt="CuRe Code" class="brand-icon" />
        <span class="brand-name">CuRe Code</span>
      </router-link>

      <div :class="['nav-links', { open: menuOpen }]">
        <router-link to="/" @click="menuOpen = false">{{ t('nav.home') }}</router-link>
        <router-link to="/docs" @click="menuOpen = false">{{ t('nav.docs') }}</router-link>
        <a href="https://github.com/broman0x/cure-code" target="_blank" rel="noopener">GitHub</a>
        <button class="lang-btn" @click="toggleLang">{{ lang === 'en' ? 'ID' : 'EN' }}</button>
      </div>

      <button class="hamburger" @click="menuOpen = !menuOpen" :aria-expanded="menuOpen">
        <span></span>
        <span></span>
      </button>
    </div>
  </nav>
</template>

<style scoped>
.nav {
  position: sticky;
  top: 0;
  z-index: 100;
  background: var(--bg);
  border-bottom: 1px solid var(--line);
}

.nav-inner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 56px;
}

.nav-brand {
  display: flex;
  align-items: center;
  gap: 8px;
  text-decoration: none;
}

.brand-icon {
  width: 32px;
  height: 32px;
  object-fit: contain;
  display: block;
}

.brand-name {
  font-size: 0.9rem;
  font-weight: 700;
  color: var(--text);
  letter-spacing: -0.01em;
}



.nav-links {
  display: flex;
  align-items: center;
  gap: 28px;
}

.nav-links a {
  font-size: 0.85rem;
  font-weight: 500;
  color: var(--text-2);
  transition: color 0.15s;
}

.nav-links a:hover,
.nav-links a.router-link-active {
  color: var(--text);
}

.lang-btn {
  font-size: 0.75rem;
  font-weight: 700;
  font-family: var(--mono);
  background: var(--surface);
  border: 1px solid var(--line);
  color: var(--text-2);
  padding: 3px 8px;
  border-radius: 4px;
  cursor: pointer;
  transition: color 0.15s;
}

.lang-btn:hover { color: var(--text); }

.hamburger {
  display: none;
  flex-direction: column;
  gap: 5px;
  background: none;
  border: none;
  cursor: pointer;
  padding: 4px;
}

.hamburger span {
  display: block;
  width: 20px;
  height: 1.5px;
  background: var(--text-2);
}

@media (max-width: 680px) {
  .hamburger { display: flex; }

  .nav-links {
    display: none;
    position: fixed;
    inset: 56px 0 0 0;
    flex-direction: column;
    justify-content: flex-start;
    padding: 32px 24px;
    gap: 24px;
    background: var(--bg);
    border-top: 1px solid var(--line);
  }

  .nav-links.open { display: flex; }

  .nav-links a { font-size: 1rem; }
}
</style>
