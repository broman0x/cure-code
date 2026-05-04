<template>
  <footer class="footer">
    <div class="container foot-inner">
      <span class="foot-brand">CuRe Code</span>
      <span class="foot-sep">·</span>
      <span class="foot-copy">Built by bromanprjkt</span>
      <span class="foot-right">© 2026 · GNU GPL v3</span>
    </div>
  </footer>
</template>

<style scoped>
.footer {
  border-top: 1px solid var(--line);
}

.foot-inner {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 20px 24px;
  font-size: 0.78rem;
}

.foot-brand {
  font-weight: 700;
  color: var(--text);
}

.foot-sep, .foot-copy {
  color: var(--text-3);
}

.foot-right {
  margin-left: auto;
  color: var(--text-3);
  font-family: var(--mono);
  font-size: 0.7rem;
}

@media (max-width: 500px) {
  .foot-inner { flex-wrap: wrap; gap: 6px; }
  .foot-right { margin-left: 0; }
}
</style>
