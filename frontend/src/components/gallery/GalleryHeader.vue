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
</script>

<template>
  <div class="mb-6 flex flex-col items-center justify-between gap-4 sm:flex-row">
    <div class="flex w-full items-center gap-2 overflow-x-auto sm:w-auto">
      <button
        @click="emit('navigate', '')"
        class="rounded-full p-2 transition-colors hover:bg-gray-100 dark:hover:bg-gray-800"
        title="Home"
      >
        <svg
          xmlns="http://www.w3.org/2000/svg"
          width="20"
          height="20"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"
          class="text-gray-600 dark:text-gray-300"
        >
          <path d="m3 9 9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"></path>
          <polyline points="9 22 9 12 15 12 15 22"></polyline>
        </svg>
      </button>
      <template v-for="crumb in breadcrumbs" :key="crumb.path">
        <span class="text-gray-400">/</span>
        <button
          @click="emit('navigate', crumb.path)"
          class="font-medium whitespace-nowrap text-gray-700 hover:text-indigo-600 dark:text-gray-300 dark:hover:text-indigo-400"
        >
          {{ crumb.name }}
        </button>
      </template>
    </div>

    <div class="mx-4 max-w-lg flex-1">
      <div class="group relative">
        <div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3">
          <svg
            class="h-5 w-5 text-gray-400 transition-colors group-focus-within:text-indigo-500"
            xmlns="http://www.w3.org/2000/svg"
            viewBox="0 0 20 20"
            fill="currentColor"
          >
            <path
              fill-rule="evenodd"
              d="M8 4a4 4 0 100 8 4 4 0 000-8zM2 8a6 6 0 1110.89 3.476l4.817 4.817a1 1 0 01-1.414 1.414l-4.816-4.816A6 6 0 012 8z"
              clip-rule="evenodd"
            />
          </svg>
        </div>
        <input
          :value="searchQuery"
          @input="
            (e) => {
              emit('update:searchQuery', (e.target as HTMLInputElement).value)
              emit('search', (e.target as HTMLInputElement).value)
            }
          "
          type="text"
          class="block w-full rounded-lg border border-gray-300 bg-white py-2 pr-3 pl-10 leading-5 text-gray-900 placeholder-gray-500 shadow-sm transition-all focus:border-indigo-500 focus:placeholder-gray-400 focus:ring-1 focus:ring-indigo-500 focus:outline-none sm:text-sm dark:border-gray-700 dark:bg-gray-800 dark:text-gray-100"
          placeholder="Search metadata (e.g. seed:123 or 'cyberpunk')"
        />
      </div>
    </div>

    <button
      @click="emit('toggleSort')"
      class="flex items-center gap-2 rounded-lg border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 shadow-sm transition-colors hover:bg-gray-50 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-200 dark:hover:bg-gray-700"
    >
      <span>Sort by Date</span>
      <span class="rounded bg-gray-100 px-2 py-0.5 text-xs uppercase dark:bg-gray-900">{{
        sortOrder
      }}</span>
    </button>
  </div>
</template>
