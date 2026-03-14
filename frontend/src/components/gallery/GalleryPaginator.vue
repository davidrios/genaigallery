<script setup lang="ts">
import { computed } from 'vue'
import { ChevronLeft, ChevronRight, MoreHorizontal } from 'lucide-vue-next'

const props = defineProps<{
  currentPage: number
  totalPages: number
}>()

const emit = defineEmits<{
  (e: 'update:currentPage', page: number): void
}>()

const visiblePages = computed(() => {
  const pages: (number | string)[] = []

  if (props.totalPages <= 7) {
    for (let i = 1; i <= props.totalPages; i++) {
      pages.push(i)
    }
    return pages
  }

  pages.push(1)

  if (props.currentPage > 3) {
    pages.push('...')
  }

  const start = Math.max(2, props.currentPage - 1)
  const end = Math.min(props.totalPages - 1, props.currentPage + 1)

  for (let i = start; i <= end; i++) {
    if (pages[pages.length - 1] !== i) {
      pages.push(i)
    }
  }

  if (props.currentPage < props.totalPages - 2) {
    pages.push('...')
  }

  if (pages[pages.length - 1] !== props.totalPages) {
    pages.push(props.totalPages)
  }

  return pages
})

const changePage = (page: number | string) => {
  if (
    typeof page === 'number' &&
    page >= 1 &&
    page <= props.totalPages &&
    page !== props.currentPage
  ) {
    emit('update:currentPage', page)
  }
}
</script>

<template>
  <nav
    v-if="totalPages > 1"
    class="flex items-center justify-center space-x-2 py-4"
    aria-label="Pagination"
  >
    <button
      :disabled="currentPage === 1"
      class="inline-flex h-9 w-9 items-center justify-center rounded-md border border-zinc-200 bg-white text-sm font-medium text-zinc-900 transition-colors hover:bg-zinc-100 hover:text-zinc-900 disabled:pointer-events-none disabled:opacity-50 dark:border-zinc-800 dark:bg-zinc-950 dark:text-zinc-50 dark:hover:bg-zinc-800 dark:hover:text-zinc-50"
      aria-label="Go to previous page"
      @click="changePage(currentPage - 1)"
    >
      <ChevronLeft class="h-4 w-4" />
    </button>

    <div class="flex items-center space-x-1">
      <template v-for="(page, index) in visiblePages" :key="index">
        <button
          v-if="typeof page === 'number'"
          :class="[
            'inline-flex h-9 min-w-[2.25rem] items-center justify-center rounded-md border text-sm font-medium transition-colors focus-visible:ring-1 focus-visible:ring-indigo-500 focus-visible:outline-none',
            page === currentPage
              ? 'border-indigo-600 bg-indigo-600 text-white shadow-sm hover:bg-indigo-600/90'
              : 'border-zinc-200 bg-white text-zinc-900 shadow-sm hover:bg-zinc-100 hover:text-zinc-900 dark:border-zinc-800 dark:bg-zinc-950 dark:text-zinc-50 dark:hover:bg-zinc-800 dark:hover:text-zinc-50',
          ]"
          :aria-current="page === currentPage ? 'page' : undefined"
          @click="changePage(page)"
        >
          {{ page }}
        </button>
        <div v-else class="flex h-9 w-9 items-center justify-center">
          <MoreHorizontal class="h-4 w-4 text-zinc-500" />
        </div>
      </template>
    </div>

    <button
      :disabled="currentPage === totalPages"
      class="inline-flex h-9 w-9 items-center justify-center rounded-md border border-zinc-200 bg-white text-sm font-medium text-zinc-900 transition-colors hover:bg-zinc-100 hover:text-zinc-900 disabled:pointer-events-none disabled:opacity-50 dark:border-zinc-800 dark:bg-zinc-950 dark:text-zinc-50 dark:hover:bg-zinc-800 dark:hover:text-zinc-50"
      aria-label="Go to next page"
      @click="changePage(currentPage + 1)"
    >
      <ChevronRight class="h-4 w-4" />
    </button>
  </nav>
</template>
