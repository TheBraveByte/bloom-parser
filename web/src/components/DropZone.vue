<script setup lang="ts">
import { ref } from 'vue'

const props = defineProps<{ files: File[] }>()
const emit = defineEmits<{ (e: 'update', files: File[]): void }>()

const dragging = ref(false)
const input = ref<HTMLInputElement>()

function add(list: FileList | File[]) {
  emit('update', [...props.files, ...Array.from(list)])
}
function remove(i: number) {
  const next = props.files.slice()
  next.splice(i, 1)
  emit('update', next)
}
function fmt(n: number) {
  return n > 1e6 ? `${(n / 1e6).toFixed(1)} MB` : `${(n / 1e3).toFixed(0)} KB`
}
</script>

<template>
  <div>
    <button
      type="button"
      class="glass glass-interactive w-full rounded-lg border border-dashed p-5 text-center
             focus-visible:outline-2 focus-visible:outline-accent"
      :class="dragging ? 'border-accent text-fg' : 'text-muted'"
      @click="input?.click()"
      @dragover.prevent="dragging = true"
      @dragleave.prevent="dragging = false"
      @drop.prevent="dragging = false; add($event.dataTransfer?.files ?? [])"
    >
      <strong class="text-fg">Click to choose</strong> or drop files<br>
      <span class="text-xs">images — multiple files process concurrently</span>
    </button>
    <input ref="input" type="file" hidden multiple accept="image/*" @change="add(($event.target as HTMLInputElement).files ?? [])">

    <ul v-if="files.length" class="glass mt-2 max-h-40 overflow-auto rounded-lg">
      <li
        v-for="(f, i) in files" :key="f.name + i"
        class="flex items-center justify-between gap-2 border-b border-line px-3 py-1.5 text-sm last:border-0 hover:bg-panel"
      >
        <span class="truncate font-mono text-xs">{{ f.name }}</span>
        <span class="flex items-center gap-2 text-xs text-muted">
          {{ fmt(f.size) }}
          <button class="text-muted hover:text-err" aria-label="remove file" @click="remove(i)">&times;</button>
        </span>
      </li>
    </ul>
  </div>
</template>
