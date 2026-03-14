<script setup lang="ts">
defineProps<{
  breadcrumbs: Array<{ name: string; path: string }>
  searchQuery: string
  sortOrder: 'asc' | 'desc'
}>()

const emit = defineEmits<{
  (e: 'navigate', path: string): void
  (e: 'update:searchQuery', value: string): void
  (e: 'search', value: string): void
  (e: 'toggleSort'): void
}>()

import { Home, Search } from 'lucide-vue-next'
</script>

<template>
  <div class="mb-6 flex flex-col items-center justify-between gap-4 sm:flex-row">
    <div class="flex w-full items-center gap-2 overflow-x-auto sm:w-auto">
      <button
        class="rounded-full p-2 transition-colors hover:bg-gray-100 dark:hover:bg-gray-800"
        title="Home"
        @click="emit('navigate', '')"
      >
        <Home class="h-5 w-5 text-gray-600 dark:text-gray-300" />
      </button>
      <template v-for="crumb in breadcrumbs" :key="crumb.path">
        <span class="text-gray-400">/</span>
        <button
          class="font-medium whitespace-nowrap text-gray-700 hover:text-indigo-600 dark:text-gray-300 dark:hover:text-indigo-400"
          @click="emit('navigate', crumb.path)"
        >
          {{ crumb.name }}
        </button>
      </template>
    </div>

    <div class="mx-4 max-w-lg flex-1">
      <div class="group relative">
        <div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3">
          <Search
            class="h-5 w-5 text-gray-400 transition-colors group-focus-within:text-indigo-500"
          />
        </div>
        <input
          :value="searchQuery"
          type="text"
          class="block w-full rounded-lg border border-gray-300 bg-white py-2 pr-3 pl-10 leading-5 text-gray-900 placeholder-gray-500 shadow-sm transition-all focus:border-indigo-500 focus:placeholder-gray-400 focus:ring-1 focus:ring-indigo-500 focus:outline-none sm:text-sm dark:border-gray-700 dark:bg-gray-800 dark:text-gray-100"
          placeholder="Search metadata (e.g. seed:123 or 'cyberpunk')"
          @input="
            (e) => {
              emit('update:searchQuery', (e.target as HTMLInputElement).value)
              emit('search', (e.target as HTMLInputElement).value)
            }
          "
        />
      </div>
    </div>

    <button
      class="flex items-center gap-2 rounded-lg border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 shadow-sm transition-colors hover:bg-gray-50 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-200 dark:hover:bg-gray-700"
      @click="emit('toggleSort')"
    >
      <span>Sort by Date</span>
      <span class="rounded bg-gray-100 px-2 py-0.5 text-xs uppercase dark:bg-gray-900">{{
        sortOrder
      }}</span>
    </button>
  </div>
</template>
