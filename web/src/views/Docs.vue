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
          <a href="#memory">Memory &amp; Context</a>
          <a href="#planning">Planning</a>
          <a href="#uninstall">Uninstallation</a>
        </nav>
      </aside>

      <main class="docs-main">
        <section id="intro">
          <div class="section-label">Overview</div>
          <h1>Documentation</h1>
          <p>CuRe Code is an autonomous AI coding agent that runs in your terminal. It can read files, edit code, run commands, and iterate using tool calls.</p>
          <p>Built in Go as a single binary, with no Node.js or Python runtime dependency for normal usage.</p>
          <div class="callout">
            <span class="callout-icon">&rarr;</span>
            <span>Primary command: <code>curecode</code></span>
          </div>
        </section>

        <section id="install">
          <div class="section-label">Getting Started</div>
          <h2>Installation</h2>

          <h3>Linux / macOS</h3>
          <pre><code>curl -fsSL https://raw.githubusercontent.com/broman0x/cure-code/main/install.sh | bash</code></pre>

          <h3>Windows (PowerShell)</h3>
          <pre><code>iex (irm https://raw.githubusercontent.com/broman0x/cure-code/main/install.ps1)</code></pre>

          <h3>Manual (from Releases)</h3>
          <pre><code># Linux example
chmod +x curecode-linux-amd64
sudo mv curecode-linux-amd64 /usr/local/bin/curecode</code></pre>
        </section>

        <section id="config">
          <div class="section-label">Setup</div>
          <h2>Configuration</h2>
          <p>On first run, CuRe Code can guide provider setup and save your selection. Runtime config paths:</p>
          <pre><code>Linux/macOS: ~/.config/curecode/config.json
Linux/macOS: ~/.config/curecode/.env
Windows: %APPDATA%\CuReCode\config.json
Windows: %APPDATA%\CuReCode\.env</code></pre>

          <h3>Environment variables</h3>
          <p>CuRe Code auto-detects provider keys from environment or <code>.env</code>:</p>
          <pre><code>GEMINI_API_KEY
OPENAI_API_KEY
ANTHROPIC_API_KEY
NVIDIA_API_KEY
XAI_API_KEY
DEEPSEEK_API_KEY
OPENROUTER_API_KEY</code></pre>

          <h3>Config JSON fields</h3>
          <div class="config-table">
            <div class="cfg-row header">
              <span>Field</span>
              <span>Description</span>
              <span>Example</span>
            </div>
            <div class="cfg-row"><span><code>language</code></span><span>UI language preference</span><span><code>en</code></span></div>
            <div class="cfg-row"><span><code>first_run</code></span><span>First-run setup state</span><span><code>false</code></span></div>
            <div class="cfg-row"><span><code>last_provider</code></span><span>Last selected provider</span><span><code>gemini</code></span></div>
            <div class="cfg-row"><span><code>last_model</code></span><span>Last selected model</span><span><code>gemini-2.5-flash</code></span></div>
            <div class="cfg-row"><span><code>install_path</code></span><span>Installer metadata</span><span><code>/home/.../.local/bin</code></span></div>
            <div class="cfg-row"><span><code>version</code></span><span>Config schema/app version</span><span><code>2.0.0</code></span></div>
          </div>
        </section>

        <section id="usage">
          <div class="section-label">Basics</div>
          <h2>Usage</h2>

          <h3>Interactive REPL</h3>
          <pre><code>curecode</code></pre>

          <h3>One-shot mode</h3>
          <pre><code>curecode "refactor main.go to use dependency injection"</code></pre>

          <h3>Resume a saved session</h3>
          <pre><code>curecode --resume session-1234567890</code></pre>

          <h3>Skip command confirmations</h3>
          <pre><code>curecode --yolo "run tests and fix failures"</code></pre>

          <h3>Pipe input</h3>
          <pre><code>echo "summarize this repo" | curecode</code></pre>
        </section>

        <section id="commands">
          <div class="section-label">CLI Reference</div>
          <h2>Flags &amp; Slash Commands</h2>

          <h3>CLI flags</h3>
          <div class="cmd-table">
            <div class="cmd-row header">
              <span>Flag</span>
              <span>Description</span>
            </div>
            <div class="cmd-row"><span><code>--version</code></span><span>Show version</span></div>
            <div class="cmd-row"><span><code>--install</code></span><span>Install to user path</span></div>
            <div class="cmd-row"><span><code>--uninstall</code></span><span>Remove binary and config</span></div>
            <div class="cmd-row"><span><code>--resume &lt;id&gt;</code></span><span>Resume a saved session</span></div>
            <div class="cmd-row"><span><code>--yolo</code></span><span>Skip tool execution confirmations</span></div>
          </div>

          <h3>REPL slash commands</h3>
          <div class="cmd-table">
            <div class="cmd-row header">
              <span>Command</span>
              <span>Description</span>
            </div>
            <div class="cmd-row"><span><code>/help</code></span><span>Show help</span></div>
            <div class="cmd-row"><span><code>/clear</code></span><span>Clear screen</span></div>
            <div class="cmd-row"><span><code>/compact</code></span><span>Clear conversation history</span></div>
            <div class="cmd-row"><span><code>/model</code></span><span>Switch provider/model</span></div>
            <div class="cmd-row"><span><code>/usage</code></span><span>Show token usage</span></div>
            <div class="cmd-row"><span><code>/save</code></span><span>Save current session</span></div>
            <div class="cmd-row"><span><code>/resume</code></span><span>Open session picker</span></div>
            <div class="cmd-row"><span><code>/ps</code></span><span>List or stop background processes</span></div>
            <div class="cmd-row"><span><code>/version</code></span><span>Show version info</span></div>
            <div class="cmd-row"><span><code>/exit</code></span><span>Exit app</span></div>
          </div>
          <p>Aliases: <code>/h</code>, <code>/cls</code>, <code>/quit</code>, <code>/q</code>.</p>
        </section>

        <section id="tools">
          <div class="section-label">Capabilities</div>
          <h2>Built-in Tools</h2>
          <p>Registered tool names currently exposed to the agent:</p>

          <div class="tool-table">
            <div class="tool-row header">
              <span>Tool</span>
              <span>Description</span>
            </div>
            <div class="tool-row"><span><code>read_file</code></span><span>Read one file (optionally by line range)</span></div>
            <div class="tool-row"><span><code>read_many_files</code></span><span>Read multiple files in one call</span></div>
            <div class="tool-row"><span><code>write_file</code></span><span>Create or overwrite file content</span></div>
            <div class="tool-row"><span><code>edit_file</code></span><span>Search-and-replace edits in a file</span></div>
            <div class="tool-row"><span><code>run_command</code></span><span>Run shell commands</span></div>
            <div class="tool-row"><span><code>list_directory</code></span><span>List files/directories in a path</span></div>
            <div class="tool-row"><span><code>grep_search</code></span><span>Regex search across the project</span></div>
            <div class="tool-row"><span><code>glob</code></span><span>Match files using glob patterns</span></div>
            <div class="tool-row"><span><code>ask_user</code></span><span>Ask user clarification questions</span></div>
            <div class="tool-row"><span><code>web_fetch</code></span><span>Fetch and read a web page</span></div>
            <div class="tool-row"><span><code>web_search</code></span><span>Search web results</span></div>
            <div class="tool-row"><span><code>get_project_summary</code></span><span>Generate high-level project overview</span></div>
            <div class="tool-row"><span><code>get_git_info</code></span><span>Show branch/status/diff summary</span></div>
            <div class="tool-row"><span><code>search_symbol</code></span><span>Find functions/types/symbols</span></div>
            <div class="tool-row"><span><code>write_todos</code></span><span>Maintain the internal task list</span></div>
            <div class="tool-row"><span><code>enter_plan_mode</code></span><span>Enable planning-only mode</span></div>
            <div class="tool-row"><span><code>exit_plan_mode</code></span><span>Return to execution mode</span></div>
          </div>
        </section>

        <section id="providers">
          <div class="section-label">Integrations</div>
          <h2>AI Providers</h2>
          <p>Provider support includes cloud and local backends:</p>

          <div class="provider-table">
            <div class="prov-row header">
              <span>Provider</span>
              <span>Env Variable</span>
              <span>Example Model</span>
            </div>
            <div class="prov-row"><span>Gemini</span><span><code>GEMINI_API_KEY</code></span><span><code>gemini-2.5-flash</code></span></div>
            <div class="prov-row"><span>OpenAI</span><span><code>OPENAI_API_KEY</code></span><span><code>gpt-4o-mini</code></span></div>
            <div class="prov-row"><span>Anthropic Claude</span><span><code>ANTHROPIC_API_KEY</code></span><span><code>claude-sonnet-4-20250514</code></span></div>
            <div class="prov-row"><span>NVIDIA NIM</span><span><code>NVIDIA_API_KEY</code></span><span><code>nvidia/nemotron-3-super-120b-a12b</code></span></div>
            <div class="prov-row"><span>xAI</span><span><code>XAI_API_KEY</code></span><span><code>grok-2-1212</code></span></div>
            <div class="prov-row"><span>DeepSeek</span><span><code>DEEPSEEK_API_KEY</code></span><span><code>deepseek-coder</code></span></div>
            <div class="prov-row"><span>OpenRouter</span><span><code>OPENROUTER_API_KEY</code></span><span><code>anthropic/claude-3.5-sonnet</code></span></div>
            <div class="prov-row"><span>Ollama (local)</span><span><em>none required</em></span><span><code>llama3</code></span></div>
          </div>

          <p>If no cloud key is detected, CuRe Code attempts local Ollama fallback.</p>
        </section>

        <section id="memory">
          <div class="section-label">Architecture</div>
          <h2>Memory &amp; Context</h2>
          <p>The agent maintains conversation history, task state, and recent context to stay coherent on long tasks.</p>
          <p>Use <code>/save</code> to persist the current session and <code>/resume</code> to continue it later.</p>
          <p>Saved sessions are stored under the <code>sessions</code> directory near the config file.</p>
        </section>

        <section id="planning">
          <div class="section-label">Advanced</div>
          <h2>Planning</h2>
          <p>Planning is exposed as tool-level controls for the agent via <code>enter_plan_mode</code> and <code>exit_plan_mode</code>.</p>
          <p>In plan mode, file modifications are restricted while the agent explores and designs an approach.</p>
          <p>The agent can synchronize internal tasks into <code>PLAN.md</code> for human-readable progress.</p>
        </section>

        <section id="uninstall">
          <div class="section-label">Cleanup</div>
          <h2>Uninstallation</h2>
          <pre><code>curecode --uninstall</code></pre>
          <p>This removes the installed binary and attempts to remove local CuRe Code configuration data.</p>
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
