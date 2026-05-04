<template>
  <div class="docs-page container">
    <aside class="sidebar">
      <nav>
        <ul>
          <li><a href="#intro">Introduction</a></li>
          <li><a href="#install">Installation</a></li>
          <li><a href="#usage">Usage</a></li>
          <li><a href="#tools">Tools</a></li>
          <li><a href="#providers">AI Providers</a></li>
          <li><a href="#uninstall">Uninstallation</a></li>
        </ul>
      </nav>
    </aside>
    
    <main class="docs-content">
      <section id="intro">
        <h1 class="title-text">Documentation</h1>
        <p>CuRe Code is an autonomous AI coding agent that lives in your terminal. It uses an agentic loop to diagnose, fix, and refactor your code by executing specialized tools.</p>
      </section>

      <section id="install">
        <h2>Installation</h2>
        <p>You can download and install CuRe Code with a single command (Linux/macOS):</p>
        <pre><code>curl -L https://github.com/broman0x/cure-code/releases/latest/download/curecode-linux-amd64 -o curecode && chmod +x curecode && ./curecode --install</code></pre>
        <p>For Windows, download the <code>.exe</code> from the releases page and run <code>.\curecode.exe --install</code> in your terminal.</p>
      </section>

      <section id="usage">
        <h2>Usage</h2>
        <p>Launch the interactive session:</p>
        <pre><code>curecode</code></pre>
        <p>Run a specific task (One-shot mode):</p>
        <pre><code>curecode "refactor main.go to use interfaces"</code></pre>
      </section>

      <section id="tools">
        <h2>Built-in Tools</h2>
        <div class="tool-list">
          <div class="tool-card">
            <strong>read_file / read_many</strong>
            <p>Read contents of one or multiple files with line numbers.</p>
          </div>
          <div class="tool-card">
            <strong>edit_file</strong>
            <p>Precision search-and-replace using target blocks.</p>
          </div>
          <div class="tool-card">
            <strong>shell</strong>
            <p>Execute terminal commands in your local environment.</p>
          </div>
          <div class="tool-card">
            <strong>project_summary</strong>
            <p>Get a high-level overview of the codebase structure.</p>
          </div>
        </div>
      </section>

      <section id="providers">
        <h2>AI Providers</h2>
        <p>CuRe Code supports multiple providers via environment variables:</p>
        <ul>
          <li><strong>Gemini:</strong> <code>GEMINI_API_KEY</code></li>
          <li><strong>OpenAI:</strong> <code>OPENAI_API_KEY</code></li>
          <li><strong>Anthropic:</strong> <code>ANTHROPIC_API_KEY</code></li>
          <li><strong>Ollama:</strong> Runs locally (no key required)</li>
        </ul>
      </section>

      <section id="uninstall">
        <h2>Uninstallation</h2>
        <p>To remove CuRe Code and all its configuration:</p>
        <pre><code>curecode --uninstall</code></pre>
      </section>
    </main>
  </div>
</template>

<style scoped>
.docs-page {
  display: flex;
  gap: 4rem;
  padding: 6rem 0;
  text-align: left;
}

.sidebar {
  width: 220px;
  position: sticky;
  top: 100px;
  height: fit-content;
}

.sidebar ul { list-style: none; }
.sidebar li { margin-bottom: 0.75rem; }
.sidebar a {
  color: var(--text-secondary);
  text-decoration: none;
  font-size: 0.95rem;
  font-weight: 600;
}
.sidebar a:hover { color: var(--primary); }

.docs-content { flex: 1; max-width: 800px; }
section { margin-bottom: 5rem; }

h1 { font-size: 3rem; margin-bottom: 1.5rem; }
h2 { font-size: 1.75rem; margin-bottom: 1.25rem; color: var(--text); margin-top: 3rem; }
h2::after { content: ''; display: block; width: 40px; height: 3px; background: var(--primary); margin-top: 0.5rem; border-radius: 2px; }

p { color: var(--text-secondary); margin-bottom: 1.25rem; line-height: 1.7; }

pre {
  background: var(--bg-secondary);
  padding: 1.5rem;
  border-radius: var(--radius);
  border: 1px solid var(--border);
  margin: 1.5rem 0;
  overflow-x: auto;
}

pre code { font-family: 'JetBrains Mono', monospace; font-size: 0.9rem; color: var(--text); }

.tool-list { display: grid; gap: 1.25rem; }
.tool-card {
  padding: 1.5rem;
  background: var(--bg-secondary);
  border: 1px solid var(--border);
  border-radius: var(--radius);
}
.tool-card strong { display: block; margin-bottom: 0.5rem; font-size: 1rem; color: var(--text); }
.tool-card p { font-size: 0.95rem; margin-bottom: 0; }

ul { padding-left: 1.5rem; color: var(--text-secondary); }
li { margin-bottom: 0.75rem; }

@media (max-width: 768px) {
  .docs-page {
    flex-direction: column;
    gap: 3rem;
    padding: 3rem 1.5rem;
  }
  
  .sidebar {
    width: 100%;
    position: relative;
    top: 0;
    border-bottom: 1px solid var(--border);
    padding-bottom: 1.5rem;
  }
  
  .sidebar ul {
    display: flex;
    overflow-x: auto;
    gap: 1.5rem;
    padding-bottom: 0.5rem;
    -webkit-overflow-scrolling: touch;
  }
  
  .sidebar ul::-webkit-scrollbar { display: none; }
  
  .sidebar li { margin-bottom: 0; white-space: nowrap; }
  
  h1 { font-size: 2.5rem; }
  h2 { font-size: 1.5rem; }
}
</style>
