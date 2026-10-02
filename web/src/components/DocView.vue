<script setup lang="ts">
import { computed } from 'vue'
import type { DocTable, DocValue, ParsedDocument } from '../lib/api'

const props = defineProps<{ doc: ParsedDocument }>()

const failed = computed(() => props.doc.pages?.filter(p => p.error).length ?? 0)

function valueText(v: DocValue | undefined): { t: string; null: boolean } {
  if (!v || Object.keys(v).length === 0) return { t: '', null: true }
  if (v.stringValue !== undefined) return { t: v.stringValue, null: false }
  if (v.intValue !== undefined) return { t: v.intValue, null: false }
  if (v.doubleValue !== undefined) return { t: String(v.doubleValue), null: false }
  if (v.boolValue !== undefined) return { t: String(v.boolValue), null: false }
  if (v.datetimeValue !== undefined) return { t: v.datetimeValue, null: false }
  return { t: '', null: true }
}

function kindTag(kind: string) {
  return (kind || '').replace('PAGE_KIND_', '').toLowerCase()
}
function colType(t: string) {
  return (t || '').replace('COLUMN_TYPE_', '').toLowerCase()
}
function rowsOf(tbl: DocTable) {
  return tbl.rows ?? []
}
</script>

<template>
  <div>
    <div class="mb-4 flex flex-wrap gap-2">
      <span class="pill">format <b>{{ doc.format?.replace('DOCUMENT_FORMAT_', '') }}</b></span>
      <span class="pill">pages <b>{{ doc.pages?.length ?? 0 }}</b></span>
      <span class="pill">failed <b :class="failed ? 'text-err' : 'text-ok'">{{ failed }}</b></span>
      <span class="pill">name <b>{{ doc.name }}</b></span>
    </div>

    <div v-for="p in doc.pages" :key="p.number" class="mb-4 rounded-lg border border-line bg-panel p-4">
      <h3 class="mb-2 flex items-center gap-2 text-base">
        Page {{ p.number }}
        <span class="rounded-full bg-well px-2 py-0.5 text-xs text-muted">{{ kindTag(p.kind) }}</span>
        <span class="text-xs text-muted">{{ p.source }}</span>
      </h3>
      <div v-if="p.error" class="mb-2 rounded-md border border-err/40 bg-err/10 px-3 py-2 text-sm text-err">
        <b>{{ p.error.code }}</b> — {{ p.error.message }}
      </div>
      <div v-for="(im, i) in p.images ?? []" :key="i" class="text-xs text-muted">
        image: {{ im.format }} · {{ im.width }}×{{ im.height }} · {{ im.colorModel }}
      </div>
      <div v-if="p.confidence != null && p.confidence >= 0" class="text-xs text-muted">
        confidence: {{ (p.confidence * 100).toFixed(1) }}%
      </div>
      <pre v-if="p.text" class="mt-2 max-h-56 overflow-auto whitespace-pre-wrap rounded-md border border-line bg-well p-2.5 text-sm">{{ p.text }}</pre>

      <div v-for="(t, ti) in p.tables ?? []" :key="ti" class="mt-2 overflow-auto">
        <table class="w-full border-collapse text-sm">
          <thead>
            <tr>
              <th v-for="c in t.columns ?? []" :key="c.name" class="border border-line bg-well px-2 py-1 text-left">
                {{ c.name }}<br>
                <span class="text-xs font-normal text-muted">{{ colType(c.type) }}</span>
              </th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(row, ri) in rowsOf(t)" :key="ri">
              <td v-for="(_, ci) in t.columns ?? []" :key="ci" class="border border-line px-2 py-1">
                <span :class="valueText(row.values?.[ci]).null ? 'italic text-muted' : ''">
                  {{ valueText(row.values?.[ci]).null ? 'null' : valueText(row.values?.[ci]).t }}
                </span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <ul v-if="p.warnings?.length" class="mt-2 list-disc pl-5 text-xs text-warn">
        <li v-for="(w, wi) in p.warnings" :key="wi">{{ w.code }}: {{ w.message }}</li>
      </ul>
    </div>

    <details class="mt-4">
      <summary class="cursor-pointer text-muted">Raw JSON</summary>
      <pre class="mt-2 max-h-96 overflow-auto rounded-lg border border-line bg-well p-3 font-mono text-xs">{{ JSON.stringify(doc, null, 2) }}</pre>
    </details>
  </div>
</template>


