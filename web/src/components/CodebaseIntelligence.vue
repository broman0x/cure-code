<script setup>
import { ref, computed } from 'vue'

const props = defineProps({
  symbols: Array,
  tasks: Array,
  usage: Object,
  toolCount: Number,
  version: String
})

const activeTab = ref('plan')

const progress = computed(() => {
  if (!props.tasks?.length) return 0
  return Math.round((props.tasks.filter(t => t.status === 'completed').length / props.tasks.length) * 100)
})
</script>

<template>
  <aside class="intel">
    <div class="intel-top">
      <span class="intel-title">Agent State</span>
      <span class="intel-ver">{{ version || 'v1.0.2' }}</span>
    </div>

    <div class="intel-progress">
      <div class="prog-bar">
        <div class="prog-fill" :style="{ width: progress + '%' }"></div>
      </div>
      <span class="prog-label">{{ progress }}%</span>
    </div>

    <div class="intel-tabs">
      <button @click="activeTab = 'plan'" :class="{ on: activeTab === 'plan' }">Plan</button>
      <button @click="activeTab = 'symbols'" :class="{ on: activeTab === 'symbols' }">Symbols</button>
      <button @click="activeTab = 'stats'" :class="{ on: activeTab === 'stats' }">Stats</button>
    </div>

    <div class="intel-body">
      <div v-if="activeTab === 'plan'">
        <div v-if="tasks?.length" class="task-list">
          <div
            v-for="(task, i) in tasks" :key="i"
            :class="['task', task.status]"
          >
            <span class="task-mark">
              <span v-if="task.status === 'completed'">✓</span>
              <span v-else-if="task.status === 'in_progress'">→</span>
              <span v-else>·</span>
            </span>
            <span class="task-text">{{ task.description }}</span>
          </div>
        </div>
        <div v-else class="empty">No active tasks.</div>
      </div>

      <div v-if="activeTab === 'symbols'">
        <div v-if="symbols?.length" class="sym-list">
          <div v-for="(sym, i) in symbols" :key="i" class="sym-row">
            <span class="sym-idx">{{ String(i + 1).padStart(2, '0') }}</span>
            <span class="sym-name">{{ sym }}</span>
          </div>
        </div>
        <div v-else class="empty">No symbols tracked.</div>
      </div>

      <div v-if="activeTab === 'stats'">
        <table class="stats-table">
          <tbody>
            <tr>
              <td class="stat-k">Tokens</td>
              <td class="stat-v">{{ usage?.total_tokens?.toLocaleString() || '—' }}</td>
            </tr>
            <tr>
              <td class="stat-k">Input</td>
              <td class="stat-v">{{ usage?.total_input_tokens?.toLocaleString() || '—' }}</td>
            </tr>
            <tr>
              <td class="stat-k">Output</td>
              <td class="stat-v">{{ usage?.total_output_tokens?.toLocaleString() || '—' }}</td>
            </tr>
            <tr>
              <td class="stat-k">Tool calls</td>
              <td class="stat-v">{{ toolCount || '—' }}</td>
            </tr>
          </tbody>
        </table>
        <div class="agent-online">
          <span class="online-dot"></span> agent running
        </div>
      </div>
    </div>
  </aside>
</template>

<style scoped>
.intel {
  background: var(--surface);
  border: 1px solid var(--line);
  border-radius: 6px;
  overflow: hidden;
  font-size: 0.82rem;
}

.intel-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 14px 16px;
  border-bottom: 1px solid var(--line);
}

.intel-title {
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--text);
  letter-spacing: 0.01em;
}

.intel-ver {
  font-size: 0.65rem;
  font-family: var(--mono);
  color: var(--text-3);
}

.intel-progress {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 16px;
  border-bottom: 1px solid var(--line);
}

.prog-bar {
  flex: 1;
  height: 3px;
  background: var(--surface-2);
  border-radius: 2px;
  overflow: hidden;
}

.prog-fill {
  height: 100%;
  background: var(--text-2);
  transition: width 0.5s ease;
}

.prog-label {
  font-family: var(--mono);
  font-size: 0.65rem;
  color: var(--text-3);
  min-width: 28px;
  text-align: right;
}

.intel-tabs {
  display: flex;
  border-bottom: 1px solid var(--line);
}

.intel-tabs button {
  flex: 1;
  background: none;
  border: none;
  border-bottom: 2px solid transparent;
  color: var(--text-3);
  font-size: 0.75rem;
  font-weight: 600;
  font-family: var(--font);
  padding: 10px 0;
  cursor: pointer;
  transition: color 0.1s, border-color 0.1s;
}

.intel-tabs button.on {
  color: var(--text);
  border-bottom-color: var(--text-2);
}

.intel-body {
  padding: 14px 16px;
  min-height: 200px;
  max-height: 320px;
  overflow-y: auto;
}

.task-list { display: flex; flex-direction: column; gap: 10px; }

.task {
  display: grid;
  grid-template-columns: 18px 1fr;
  gap: 8px;
  align-items: start;
}

.task-mark {
  font-family: var(--mono);
  font-size: 0.72rem;
  padding-top: 1px;
  color: var(--text-3);
}

.task.completed .task-mark { color: var(--green); }
.task.in_progress .task-mark { color: var(--blue); }

.task-text {
  color: var(--text-2);
  line-height: 1.45;
}

.task.completed .task-text { color: var(--text-3); text-decoration: line-through; }
.task.in_progress .task-text { color: var(--text); }

.sym-list { display: flex; flex-direction: column; }

.sym-row {
  display: grid;
  grid-template-columns: 24px 1fr;
  gap: 10px;
  padding: 7px 0;
  border-bottom: 1px solid var(--line);
  align-items: baseline;
}

.sym-row:last-child { border-bottom: none; }

.sym-idx {
  font-family: var(--mono);
  font-size: 0.65rem;
  color: var(--text-3);
}

.sym-name {
  font-family: var(--mono);
  font-size: 0.78rem;
  color: var(--text);
}

.stats-table {
  width: 100%;
  border-collapse: collapse;
  margin-bottom: 16px;
}

.stats-table tr {
  border-bottom: 1px solid var(--line);
}

.stats-table tr:last-child {
  border-bottom: none;
}

.stat-k {
  padding: 8px 0;
  color: var(--text-3);
  font-size: 0.75rem;
  width: 50%;
}

.stat-v {
  padding: 8px 0;
  color: var(--text);
  font-family: var(--mono);
  font-size: 0.75rem;
  text-align: right;
}

.agent-online {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 0.7rem;
  font-family: var(--mono);
  color: var(--text-3);
}

.online-dot {
  width: 5px;
  height: 5px;
  border-radius: 50%;
  background: var(--green);
}

.empty {
  color: var(--text-3);
  font-size: 0.78rem;
  padding: 16px 0;
}
</style>
