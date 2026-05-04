<template>
  <footer class="footer">
    <div class="container footer-content">
      <div class="footer-left">
        <span class="logo">CuRe Code</span>
        <p>Built by bromanprjkt</p>
      </div>
      <div class="footer-right">
        <p>© 2026 CuRe Code. MIT License.</p>
      </div>
    </div>
  </footer>
</template>

<style scoped>
.footer {
  padding: 4rem 0;
  border-top: 1px solid var(--border);
  background: var(--bg-secondary);
}

.footer-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.logo {
  font-weight: 700;
  font-size: 0.9rem;
  display: block;
  margin-bottom: 0.25rem;
}

p {
  font-size: 0.85rem;
  color: var(--text-secondary);
}

@media (max-width: 600px) {
  .footer-content {
    flex-direction: column;
    gap: 1rem;
    text-align: center;
  }
}
</style>

