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
  { type: 'meta', text: 'v1.0.2 · Galileo · bromanprjkt' },
  { type: 'blank' },
  { type: 'input', text: 'cure > add error handling to auth/login.go' },
  { type: 'thought', text: 'Reading auth/login.go...' },
  { type: 'tool', text: 'read_file', args: 'auth/login.go', ok: true },
  { type: 'thought', text: 'Found 3 unhandled error paths. Patching...' },
  { type: 'tool', text: 'edit_file', args: 'auth/login.go', ok: true },
  { type: 'blank' },
  { type: 'result', text: 'Done. Added nil-check + proper error returns to 3 call sites.' },
]

onMounted(() => {
  let i = 0
  const tick = setInterval(() => {
    if (i < fullSequence.length) {
      lines.value.push(fullSequence[i++])
    } else {
      clearInterval(tick)
    }
  }, 650)
})
</script>

<template>
  <div class="term">
    <div class="term-bar">
      <div class="term-dots">
        <span class="dot red"></span>
        <span class="dot yellow"></span>
        <span class="dot green"></span>
      </div>
      <span class="term-title">curecode — zsh</span>
      <span></span>
    </div>
    <div class="term-body">
      <template v-for="(line, i) in lines" :key="i">
        <div v-if="line.type === 'banner'" class="l-banner">
          <div v-for="(row, j) in line.text" :key="j">{{ row }}</div>
        </div>
        <div v-else-if="line.type === 'meta'" class="l-meta">{{ line.text }}</div>
        <div v-else-if="line.type === 'blank'" class="l-blank"></div>
        <div v-else-if="line.type === 'input'" class="l-input">{{ line.text }}</div>
        <div v-else-if="line.type === 'thought'" class="l-thought">· {{ line.text }}</div>
        <div v-else-if="line.type === 'tool'" class="l-tool">
          <span class="tool-ok">{{ line.ok ? '✓' : '✗' }}</span>
          <span class="tool-name">{{ line.text }}</span>
          <span class="tool-args">{{ line.args }}</span>
        </div>
        <div v-else-if="line.type === 'result'" class="l-result">{{ line.text }}</div>
      </template>
      <span class="cursor">█</span>
    </div>
  </div>
</template>

<style scoped>
.term {
  background: #0c0c0b;
  border: 1px solid var(--line);
  border-radius: 8px;
  overflow: hidden;
  font-family: var(--mono);
}

.term-bar {
  display: grid;
  grid-template-columns: auto 1fr auto;
  align-items: center;
  padding: 10px 14px;
  background: var(--surface);
  border-bottom: 1px solid var(--line);
}

.term-dots {
  display: flex;
  gap: 6px;
}

.dot {
  width: 11px;
  height: 11px;
  border-radius: 50%;
}
.dot.red { background: #ff5f57; }
.dot.yellow { background: #febc2e; }
.dot.green { background: #28c840; }

.term-title {
  text-align: center;
  font-size: 0.72rem;
  color: var(--text-3);
}

.term-body {
  padding: 16px 20px 20px;
  min-height: 360px;
  font-size: 0.8rem;
  line-height: 1.55;
  color: #b8b8b4;
}

.l-banner {
  color: #4a9eff;
  font-size: 0.55rem;
  line-height: 1.15;
  white-space: pre;
  overflow-x: auto;
  margin-bottom: 10px;
}
.l-banner::-webkit-scrollbar { display: none; }

.l-meta { color: var(--text-3); margin-bottom: 4px; }
.l-blank { height: 0.8rem; }
.l-input { color: #eeeeec; font-weight: 500; margin-top: 4px; }
.l-thought { color: var(--text-3); padding-left: 4px; }

.l-tool {
  display: flex;
  gap: 10px;
  align-items: baseline;
}
.tool-ok { color: var(--green); font-size: 0.75rem; }
.tool-name { color: #eeeeec; }
.tool-args { color: var(--text-3); }

.l-result { color: var(--green); margin-top: 4px; }

.cursor {
  color: var(--text-2);
  animation: blink 1.1s step-end infinite;
  font-size: 0.9rem;
}

@keyframes blink {
  0%, 100% { opacity: 1; }
  50% { opacity: 0; }
}

@media (max-width: 600px) {
  .l-banner { font-size: 0.35rem; }
  .term-body { font-size: 0.72rem; padding: 12px 14px; }
}
</style>
