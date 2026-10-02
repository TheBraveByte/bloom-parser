<script setup lang="ts">
import { computed } from 'vue'
import { parseCSV, type TableExtractResult } from '../lib/api'

const props = defineProps<{ result: TableExtractResult; singleName?: string }>()

const rows = computed(() => parseCSV(props.result.csv))
const header = computed(() => rows.value[0] ?? [])
const flagIdx = computed(() => header.value.indexOf('flag'))
const failed = computed(() => props.result.files.filter(f => f.error))

function download() {
  const a = document.createElement('a')
  a.href = URL.createObjectURL(new Blob([props.result.csv], { type: 'text/csv' }))
  const base = props.singleName?.replace(/\.[^.]*$/, '') ?? 'tables'
  a.download = `${base}.normalized.csv`
  a.click()
}
</script>

<template>
  <div>
    <div class="mb-4 flex flex-wrap gap-2">
      <span class="pill">files <b>{{ result.files.length }}</b></span>
      <span class="pill">facts <b>{{ result.rows }}</b></span>
      <span class="pill">flagged <b>{{ result.flagged }}</b></span>
      <span class="pill">failed <b :class="failed.length ? 'text-err' : 'text-ok'">{{ failed.length }}</b></span>
      <button class="pill text-accent hover:border-accent" @click="download">download CSV</button>
    </div>

    <div v-for="f in failed" :key="f.file" class="mb-2 rounded-md border border-err/40 bg-err/10 px-3 py-2 text-sm text-err">
      <b>{{ f.file }}</b> — {{ f.error }}
    </div>

    <div class="overflow-auto rounded-lg border border-line">
      <table class="w-full border-collapse text-sm">
        <thead>
          <tr>
            <th v-for="h in header" :key="h" class="border border-line bg-well px-2 py-1.5 text-left text-muted">{{ h }}</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="(r, i) in rows.slice(1)" :key="i"
            :class="flagIdx >= 0 && r[flagIdx] ? 'text-warn' : ''"
            class="hover:bg-panel"
          >
            <td v-for="j in header.length" :key="j" class="border border-line px-2 py-1 font-mono text-xs">{{ r[j - 1] }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>


