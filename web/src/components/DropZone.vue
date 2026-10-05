<script setup lang="ts">
import { computed, ref } from 'vue'
import { PhCloudArrowUp, PhX } from '@phosphor-icons/vue'

const props = defineProps<{ files: File[] }>()
const emit = defineEmits<{ (e: 'update', files: File[]): void }>()

// Depth counter, not a boolean: dragenter/dragleave pairs fire per child
// element, so a boolean flickers when crossing the icon and text nodes.
const dragDepth = ref(0)
const input = ref<HTMLInputElement>()

const dragging = computed(() => dragDepth.value > 0)

function add(list: FileList | File[]) {
  const seen = new Set(props.files.map(f => `${f.name}:${f.size}:${f.lastModified}`))
  const fresh = Array.from(list).filter(f => {
    const k = `${f.name}:${f.size}:${f.lastModified}`
    return !seen.has(k) && seen.add(k)
  })
  if (fresh.length) emit('update', [...props.files, ...fresh])
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
      class="grid w-full place-items-center rounded-md border border-dashed p-6 text-center transition-colors duration-150"
      :class="dragging ? 'border-accent bg-accent/5 text-fg' : 'border-line text-muted hover:border-muted'"
      @click="input?.click()"
      @dragenter.prevent="dragDepth++"
      @dragover.prevent
      @dragleave.prevent="dragDepth = Math.max(0, dragDepth - 1)"
      @drop.prevent="dragDepth = 0; add($event.dataTransfer?.files ?? [])"
    >
      <PhCloudArrowUp :size="26" :class="dragging ? 'text-accent' : 'text-muted'" />
      <span class="mt-2 text-sm">
        <strong class="font-medium text-fg">Click to choose</strong> or drop files
      </span>
      <span class="mt-0.5 text-xs">images — multiple files process concurrently</span>
    </button>
    <input ref="input" type="file" hidden multiple accept="image/*" @change="add(($event.target as HTMLInputElement).files ?? [])">

    <div v-if="files.length" class="mt-3">
      <div class="mb-1 flex items-center justify-between">
        <span class="text-xs text-muted">{{ files.length }} queued</span>
        <button class="text-xs text-muted hover:text-err" @click="emit('update', [])">clear all</button>
      </div>
      <ul class="max-h-44 overflow-auto rounded-md border border-line">
        <li
          v-for="(f, i) in files" :key="f.name + i"
          class="flex items-center justify-between gap-2 border-b border-line px-3 py-1.5 text-sm last:border-0 hover:bg-well"
        >
          <span class="truncate font-mono text-xs">{{ f.name }}</span>
          <span class="flex shrink-0 items-center gap-2 text-xs text-muted">
            {{ fmt(f.size) }}
            <button class="text-muted hover:text-err" :aria-label="`remove ${f.name}`" @click="remove(i)">
              <PhX :size="13" />
            </button>
          </span>
        </li>
      </ul>
    </div>
  </div>
</template>
