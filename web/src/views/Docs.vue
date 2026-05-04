<template>
  <div class="docs-page">
    <div class="container docs-wrap">
      <aside class="sidebar">
        <div class="sidebar-label">On this page</div>
        <nav>
          <a href="#intro">Introduction</a>
          <a href="#install">Installation</a>
          <a href="#config">Configuration</a>
          <a href="#usage">Usage</a>
          <a href="#commands">Commands</a>
          <a href="#tools">Built-in Tools</a>
          <a href="#providers">AI Providers</a>
          <a href="#memory">Memory & Context</a>
          <a href="#planning">Planning Mode</a>
          <a href="#uninstall">Uninstallation</a>
        </nav>
      </aside>

      <main class="docs-main">
        <section id="intro">
          <div class="section-label">Overview</div>
          <h1>Documentation</h1>
          <p>CuRe Code (v1.0.2 &mdash; Galileo) is an autonomous AI coding agent that runs entirely in your terminal. It uses an agentic loop powered by function-calling to read, edit, and reason about your codebase without leaving the CLI.</p>
          <p>Built in Go as a single static binary. No Node.js, no Python runtime &mdash; just download and run.</p>
          <div class="callout">
            <span class="callout-icon">&rarr;</span>
            <span>Supports Gemini, OpenAI, Claude, xAI, NVIDIA, DeepSeek, and local Ollama models.</span>
          </div>
        </section>

        <section id="install">
          <div class="section-label">Getting Started</div>
          <h2>Installation</h2>

          <h3>Linux / macOS</h3>
          <p>Run the installer script with a single command:</p>
          <pre><code>curl -fsSL https://raw.githubusercontent.com/broman0x/cure-code/main/install.sh | bash</code></pre>

          <h3>Windows (PowerShell)</h3>
          <p>Run this in an elevated PowerShell window:</p>
          <pre><code>iex (irm https://raw.githubusercontent.com/broman0x/cure-code/main/install.ps1)</code></pre>

          <h3>Manual (from Releases)</h3>
          <p>Download the binary for your platform from the <a href="https://github.com/broman0x/cure-code/releases" target="_blank">GitHub Releases</a> page, make it executable, and place it in your <code>$PATH</code>.</p>
          <pre><code># Linux example
chmod +x curecode-linux-amd64
sudo mv curecode-linux-amd64 /usr/local/bin/cure</code></pre>
        </section>

        <section id="config">
          <div class="section-label">Setup</div>
          <h2>Configuration</h2>
          <p>On first run, CuRe Code will prompt you to configure your preferred AI provider and API key. Configuration is stored at <code>~/.curecode/.env</code>.</p>
          <p>You can also create a <code>.env</code> file manually in the project directory or set environment variables directly:</p>
          <pre><code># ~/.curecode/.env

PROVIDER=gemini
GEMINI_API_KEY=your_key_here
MODEL=gemini-2.5-flash

# Optional overrides
MAX_TOKENS=8192
TEMPERATURE=0.2</code></pre>

          <h3>Available config keys</h3>
          <div class="config-table">
            <div class="cfg-row header">
              <span>Key</span>
              <span>Description</span>
              <span>Default</span>
            </div>
            <div class="cfg-row"><span><code>PROVIDER</code></span><span>AI provider to use</span><span><code>gemini</code></span></div>
            <div class="cfg-row"><span><code>MODEL</code></span><span>Model name/ID</span><span>provider default</span></div>
            <div class="cfg-row"><span><code>MAX_TOKENS</code></span><span>Max output tokens per turn</span><span><code>8192</code></span></div>
            <div class="cfg-row"><span><code>TEMPERATURE</code></span><span>Sampling temperature (0â€“2)</span><span><code>0.2</code></span></div>
            <div class="cfg-row"><span><code>SYSTEM_PROMPT</code></span><span>Custom system prompt override</span><span>â€”</span></div>
          </div>
        </section>

        <section id="usage">
          <div class="section-label">Basics</div>
          <h2>Usage</h2>

          <h3>Interactive REPL</h3>
          <p>Start an interactive session in your project directory:</p>
          <pre><code>cure</code></pre>

          <h3>One-shot mode</h3>
          <p>Run a specific task and exit immediately:</p>
          <pre><code>cure "refactor main.go to use dependency injection"</code></pre>

          <h3>With context files</h3>
          <p>Attach files directly to your prompt using <code>@</code> syntax:</p>
          <pre><code>cure "add unit tests for @internal/agent/agent.go"</code></pre>

          <h3>Plan mode</h3>
          <p>Force the agent into structured planning before executing:</p>
          <pre><code>cure --plan "migrate database layer to use pgx/v5"</code></pre>
        </section>

        <section id="commands">
          <div class="section-label">CLI Reference</div>
          <h2>Slash Commands</h2>
          <p>While in the interactive REPL, these slash commands are available:</p>

          <div class="cmd-table">
            <div class="cmd-row header">
              <span>Command</span>
              <span>Description</span>
            </div>
            <div class="cmd-row"><span><code>/help</code></span><span>Show all available commands</span></div>
            <div class="cmd-row"><span><code>/clear</code></span><span>Clear the conversation history</span></div>
            <div class="cmd-row"><span><code>/compact</code></span><span>Manually trigger context compaction</span></div>
            <div class="cmd-row"><span><code>/plan</code></span><span>Toggle planning mode on/off</span></div>
            <div class="cmd-row"><span><code>/ps</code></span><span>List all background processes</span></div>
            <div class="cmd-row"><span><code>/kill &lt;id&gt;</code></span><span>Kill a background process by ID</span></div>
            <div class="cmd-row"><span><code>/todos</code></span><span>Show the current task plan</span></div>
            <div class="cmd-row"><span><code>/version</code></span><span>Print version info</span></div>
            <div class="cmd-row"><span><code>/exit</code></span><span>Exit the session</span></div>
          </div>
        </section>

        <section id="tools">
          <div class="section-label">Capabilities</div>
          <h2>Built-in Tools</h2>
          <p>The agent has access to these tools, which it can call automatically during a task:</p>

          <div class="tool-table">
            <div class="tool-row header">
              <span>Tool</span>
              <span>Description</span>
            </div>
            <div class="tool-row"><span><code>read_file</code></span><span>Read a file with line numbers</span></div>
            <div class="tool-row"><span><code>read_many_files</code></span><span>Read multiple files at once</span></div>
            <div class="tool-row"><span><code>edit_file</code></span><span>Search-and-replace within a file</span></div>
            <div class="tool-row"><span><code>write_file</code></span><span>Create or overwrite a file</span></div>
            <div class="tool-row"><span><code>delete_file</code></span><span>Delete a file from the filesystem</span></div>
            <div class="tool-row"><span><code>shell</code></span><span>Execute a shell command (with confirmation)</span></div>
            <div class="tool-row"><span><code>grep</code></span><span>Search across files by pattern</span></div>
            <div class="tool-row"><span><code>search_symbol</code></span><span>Find functions, types, and methods by name</span></div>
            <div class="tool-row"><span><code>get_project_summary</code></span><span>Get a tree overview of the codebase</span></div>
            <div class="tool-row"><span><code>write_todos</code></span><span>Write and update the task plan</span></div>
            <div class="tool-row"><span><code>delegate_task</code></span><span>Spawn a sub-agent for a subtask</span></div>
            <div class="tool-row"><span><code>web_search</code></span><span>Search the web for information</span></div>
          </div>
        </section>

        <section id="providers">
          <div class="section-label">Integrations</div>
          <h2>AI Providers</h2>
          <p>Set your provider and corresponding API key in <code>~/.curecode/.env</code>:</p>

          <div class="provider-table">
            <div class="prov-row header">
              <span>Provider</span>
              <span>Env Variable</span>
              <span>Example Model</span>
            </div>
            <div class="prov-row"><span>Gemini</span><span><code>GEMINI_API_KEY</code></span><span><code>gemini-2.5-flash</code></span></div>
            <div class="prov-row"><span>OpenAI</span><span><code>OPENAI_API_KEY</code></span><span><code>gpt-4o</code></span></div>
            <div class="prov-row"><span>Anthropic</span><span><code>ANTHROPIC_API_KEY</code></span><span><code>claude-sonnet-4-5</code></span></div>
            <div class="prov-row"><span>xAI (Grok)</span><span><code>XAI_API_KEY</code></span><span><code>grok-2</code></span></div>
            <div class="prov-row"><span>NVIDIA NIM</span><span><code>NVIDIA_API_KEY</code></span><span><code>meta/llama-3.3-70b</code></span></div>
            <div class="prov-row"><span>DeepSeek</span><span><code>DEEPSEEK_API_KEY</code></span><span><code>deepseek-coder</code></span></div>
            <div class="prov-row"><span>Ollama</span><span><em>none required</em></span><span><code>llama3.2</code></span></div>
          </div>

          <p>For Ollama, make sure the Ollama server is running locally (<code>ollama serve</code>) before starting CuRe Code.</p>
        </section>

        <section id="memory">
          <div class="section-label">Architecture</div>
          <h2>Memory &amp; Context</h2>
          <p>CuRe Code v1.0.2 (Galileo) introduces <strong>Agentic Memory</strong> &mdash; a system that actively manages the context window across long sessions.</p>

          <h3>How it works</h3>
          <p>When the token count approaches the model's context limit, the agent automatically compacts the conversation: summarizing past turns into a dense memory block and clearing old messages. This allows sessions to run indefinitely without degrading quality.</p>

          <h3>Spatial Memory</h3>
          <p>The agent tracks the last 20 unique code symbols it has encountered (functions, types, methods). These symbols are injected into every prompt as part of the system context, giving the agent persistent awareness of the codebase structure.</p>

          <h3>Manual compaction</h3>
          <pre><code>/compact</code></pre>
        </section>

        <section id="planning">
          <div class="section-label">Advanced</div>
          <h2>Planning Mode</h2>
          <p>Planning mode forces the agent to decompose a task into structured steps before executing any code changes. A <code>PLAN.md</code> file is created and kept in sync as tasks are completed.</p>
          <pre><code># Enable at session start
cure --plan "add OAuth2 login to the API"

# Toggle during a session
/plan</code></pre>
          <p>The agent will not write any files until the plan is confirmed. Each step is marked as completed automatically as the agent works through the task.</p>
        </section>

        <section id="uninstall">
          <div class="section-label">Cleanup</div>
          <h2>Uninstallation</h2>
          <p>Remove the binary and all configuration:</p>
          <pre><code>cure --uninstall</code></pre>
          <p>This removes the binary from <code>/usr/local/bin</code> (or <code>%USERPROFILE%\AppData\Local\Programs</code> on Windows) and deletes <code>~/.curecode</code>.</p>
        </section>

      </main>
    </div>
  </div>
</template>

<style scoped>
.docs-page {
  width: 100%;
  overflow-x: hidden;
}

.docs-wrap {
  display: flex;
  flex-direction: column;
  padding: 0;
  width: 100%;
  max-width: 100%;
}

.sidebar {
  position: sticky;
  top: 56px;
  z-index: 50;
  background: var(--bg);
  border-bottom: 1px solid var(--line);
  padding: 10px 16px;
  width: 100%;
  flex-shrink: 0;
}

.sidebar-label {
  display: none;
}

.sidebar nav {
  display: flex;
  flex-direction: row;
  flex-wrap: nowrap;
  gap: 6px;
  overflow-x: auto;
  -webkit-overflow-scrolling: touch;
  scrollbar-width: none;
  padding-bottom: 2px;
}

.sidebar nav::-webkit-scrollbar { display: none; }

.sidebar nav a {
  white-space: nowrap;
  font-size: 0.72rem;
  color: var(--text-3);
  text-decoration: none;
  padding: 4px 10px;
  border: 1px solid var(--line);
  border-radius: 99px;
  flex-shrink: 0;
  transition: color 0.1s;
}

.sidebar nav a:hover { color: var(--text); }

.docs-main {
  padding: 24px 16px 64px;
  min-width: 0;
  width: 100%;
}

.docs-main section {
  padding-bottom: 40px;
  border-bottom: 1px solid var(--line);
  margin-bottom: 40px;
}

.docs-main section:last-child {
  border-bottom: none;
  margin-bottom: 0;
}

.section-label {
  font-size: 0.62rem;
  font-family: var(--mono);
  color: var(--text-3);
  text-transform: uppercase;
  letter-spacing: 0.08em;
  margin-bottom: 10px;
}

h1 {
  font-size: 1.6rem;
  font-weight: 700;
  color: var(--text);
  margin-bottom: 16px;
  line-height: 1.2;
}

h2 {
  font-size: 1.15rem;
  font-weight: 600;
  color: var(--text);
  margin-bottom: 12px;
}

h3 {
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--text);
  margin: 20px 0 8px;
}

p {
  font-size: 0.875rem;
  color: var(--text-2);
  line-height: 1.75;
  margin-bottom: 12px;
  word-break: break-word;
}

a {
  color: var(--blue);
  text-decoration: none;
}

a:hover { text-decoration: underline; }

pre {
  background: var(--surface);
  border: 1px solid var(--line);
  border-radius: 5px;
  padding: 12px 14px;
  margin: 12px 0 18px;
  overflow-x: auto;
  -webkit-overflow-scrolling: touch;
  max-width: 100%;
}

pre code {
  font-family: var(--mono);
  font-size: 0.75rem;
  color: var(--text);
  white-space: pre;
  display: block;
}

code {
  font-family: var(--mono);
  font-size: 0.78rem;
  color: var(--text);
  background: var(--surface);
  padding: 1px 5px;
  border-radius: 3px;
  border: 1px solid var(--line);
  word-break: break-all;
}

pre code {
  background: none;
  border: none;
  padding: 0;
  word-break: normal;
}

.callout {
  display: flex;
  gap: 10px;
  align-items: flex-start;
  padding: 10px 14px;
  background: var(--surface);
  border: 1px solid var(--line);
  border-left: 3px solid var(--blue);
  font-size: 0.82rem;
  color: var(--text-2);
  margin: 16px 0;
  border-radius: 0 5px 5px 0;
}

.callout-icon { color: var(--blue); flex-shrink: 0; }

.config-table,
.cmd-table,
.tool-table,
.provider-table {
  margin: 14px 0 20px;
  border: 1px solid var(--line);
  border-radius: 5px;
  overflow-x: auto;
  -webkit-overflow-scrolling: touch;
  width: 100%;
  display: block;
}

.cfg-row,
.cmd-row,
.tool-row,
.prov-row {
  display: flex;
  flex-wrap: wrap;
  padding: 8px 12px;
  border-bottom: 1px solid var(--line);
  font-size: 0.78rem;
  gap: 4px 12px;
  min-width: 0;
}

.cfg-row:last-child,
.cmd-row:last-child,
.tool-row:last-child,
.prov-row:last-child { border-bottom: none; }

.cfg-row.header,
.cmd-row.header,
.tool-row.header,
.prov-row.header {
  background: var(--surface);
  font-size: 0.65rem;
  font-weight: 600;
  color: var(--text-3);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.cfg-row span,
.cmd-row span,
.tool-row span,
.prov-row span {
  color: var(--text-2);
  flex: 1;
  min-width: 100px;
}

.cfg-row span:first-child,
.cmd-row span:first-child,
.tool-row span:first-child,
.prov-row span:first-child {
  color: var(--text);
  min-width: 80px;
}

@media (min-width: 769px) {
  .docs-wrap {
    display: grid;
    grid-template-columns: 200px 1fr;
    gap: 0 48px;
    padding: 48px 24px 96px;
    max-width: 1080px;
    margin: 0 auto;
    align-items: start;
  }

  .sidebar {
    position: sticky;
    top: 72px;
    background: transparent;
    border-bottom: none;
    padding: 0;
    height: fit-content;
  }

  .sidebar-label {
    display: block;
    font-size: 0.65rem;
    font-family: var(--mono);
    color: var(--text-3);
    text-transform: uppercase;
    letter-spacing: 0.08em;
    margin-bottom: 12px;
  }

  .sidebar nav {
    flex-direction: column;
    flex-wrap: nowrap;
    overflow-x: visible;
    gap: 2px;
    padding-bottom: 0;
  }

  .sidebar nav a {
    border: none;
    border-radius: 4px;
    padding: 5px 8px;
    font-size: 0.8rem;
    white-space: normal;
  }

  .docs-main {
    padding: 0;
  }

  .docs-main section {
    padding-bottom: 64px;
    margin-bottom: 64px;
  }

  h1 { font-size: 2rem; margin-bottom: 20px; }
  h2 { font-size: 1.3rem; margin-bottom: 16px; }
  h3 { font-size: 0.9rem; margin: 24px 0 8px; }
  p { font-size: 0.9rem; margin-bottom: 14px; }

  pre {
    padding: 16px 20px;
    margin: 14px 0 20px;
  }

  pre code { font-size: 0.8rem; }

  .cfg-row { display: grid; grid-template-columns: 1.2fr 1.5fr 1fr; flex-wrap: unset; }
  .cmd-row { display: grid; grid-template-columns: 1fr 1.5fr; flex-wrap: unset; }
  .tool-row { display: grid; grid-template-columns: 1fr 1.5fr; flex-wrap: unset; }
  .prov-row { display: grid; grid-template-columns: 1fr 1fr 1fr; flex-wrap: unset; }

  .cfg-row span,
  .cmd-row span,
  .tool-row span,
  .prov-row span { min-width: 0; align-self: center; }
}
</style>
