<script setup>
import { ref, onMounted } from 'vue'

const banner = [
  `  ██████╗██╗   ██╗██████╗ ███████╗    ██████╗ ██████╗ ██████╗ ███████╗`,
  ` ██╔════╝██║   ██║██╔══██╗██╔════╝    ██╔════╝██╔═══██╗██╔══██╗██╔════╝`,
  ` ██║     ██║   ██║██████╔╝█████╗      ██║     ██║   ██║██║  ██║█████╗  `,
  ` ██║     ██║   ██║██╔══██╗██╔══╝      ██║     ██║   ██║██║  ██║██╔══╝  `,
  ` ╚██████╗╚██████╔╝██║  ██║███████╗    ╚██████╗╚██████╔╝██████╔╝███████╗`,
  `  ╚═════╝ ╚═════╝ ╚═╝  ╚═╝╚══════╝     ╚═════╝ ╚═════╝ ╚═════╝ ╚══════╝`,
]

const lines = ref([])
const fullSequence = [
  { type: 'banner', text: banner },
  { type: 'info', text: '  v1.0.1 | AI Coding Agent by bromanprjkt' },
  { type: 'info', text: '  Provider: Gemini (gemini-2.0-flash)' },
  { type: 'input', text: 'cure > add error handling to auth/login.go' },
  { type: 'thought', text: 'Analyzing auth/login.go...' },
  { type: 'tool', text: 'read_file auth/login.go', status: '✔' },
  { type: 'thought', text: 'Found unhandled error. Applying fix...' },
  { type: 'tool', text: 'edit_file auth/login.go', status: '✔' },
  { type: 'ai', text: '✦ CURE CODE: Added error validation. Logic is now robust.' }
]

onMounted(() => {
  let i = 0
  const interval = setInterval(() => {
    if (i < fullSequence.length) {
      lines.value.push(fullSequence[i])
      i++
    } else {
      clearInterval(interval)
    }
  }, 800)
})
</script>

<template>
  <div class="terminal-container container">
    <div class="terminal">
      <div class="terminal-header">
        <div class="terminal-dots">
          <span></span><span></span><span></span>
        </div>
        <div class="terminal-title">curecode — bash</div>
      </div>
      <div class="terminal-body">
        <div v-for="(line, idx) in lines" :key="idx" :class="['line', line.type]">
          <template v-if="line.type === 'banner'">
            <div class="banner-text">
              <div v-for="(bLine, bIdx) in line.text" :key="bIdx">{{ bLine }}</div>
            </div>
          </template>
          <template v-else-if="line.type === 'input'">
            <span class="prompt">{{ line.text }}</span>
          </template>
          <template v-else-if="line.type === 'thought'">
            <div class="thought-header">┌─ THOUGHT</div>
            <div class="thought-body">│ {{ line.text }}</div>
          </template>
          <template v-else-if="line.type === 'tool'">
            <span class="tool-icon">{{ line.status }}</span> {{ line.text }}
          </template>
          <template v-else-if="line.type === 'ai'">
            <div class="ai-msg">{{ line.text }}</div>
          </template>
          <template v-else>
            {{ line.text }}
          </template>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.terminal-container {
  padding: 2rem 0 6rem;
}

.terminal {
  background: #0d1117;
  border: 1px solid #30363d;
  border-radius: var(--radius);
  overflow: hidden;
  max-width: 850px;
  margin: 0 auto;
  box-shadow: 0 20px 50px rgba(0,0,0,0.15);
}

.terminal-header {
  background: #161b22;
  padding: 0.75rem 1rem;
  border-bottom: 1px solid #30363d;
  display: flex;
  align-items: center;
}

.terminal-dots {
  display: flex;
  gap: 8px;
}

.terminal-dots span {
  width: 12px;
  height: 12px;
  background: #30363d;
  border-radius: 50%;
}

.terminal-title {
  flex: 1;
  text-align: center;
  font-size: 0.8rem;
  color: #8b949e;
}

.terminal-body {
  padding: 1.5rem;
  font-family: 'JetBrains Mono', monospace;
  font-size: 0.9rem;
  min-height: 400px;
  text-align: left;
  color: #c9d1d9;
  overflow-x: hidden;
}

.line {
  margin-bottom: 0.4rem;
  word-break: break-all;
}

.banner-text {
  color: var(--primary);
  font-size: 0.65rem;
  line-height: 1.2;
  margin-bottom: 1rem;
  font-weight: bold;
  white-space: pre;
  overflow-x: auto;
}

.banner-text::-webkit-scrollbar { display: none; }

.info { color: #8b949e; }
.input { margin-top: 1rem; }
.prompt { color: #fff; font-weight: bold; }
.thought-header { color: #8b949e; font-weight: bold; margin-top: 1rem; }
.thought-body { color: #8b949e; border-left: 2px solid #30363d; padding-left: 12px; }
.tool-icon { color: var(--primary); }
.ai-msg { color: var(--primary); font-weight: bold; margin-top: 1.5rem; }


@media (max-width: 768px) {
  .terminal-container { padding: 1rem; }
  .banner-text { font-size: 0.35rem; }
  .terminal-body { font-size: 0.75rem; padding: 1rem; }
  .terminal-title { display: none; }
}
</style>

