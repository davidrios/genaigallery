<script setup lang="ts">
import { Image } from 'lucide-vue-next'
</script>

<template>
  <div
    class="group relative overflow-hidden rounded-xl bg-white shadow-md transition-all duration-300 hover:shadow-xl dark:bg-gray-800"
  >
    <div
      class="aspect-w-1 aspect-h-1 xl:aspect-w-7 xl:aspect-h-8 w-full animate-pulse overflow-hidden bg-gray-200 dark:bg-gray-700"
    >
      <img
        class="h-full w-full object-cover object-center transition-opacity duration-300 group-hover:opacity-75"
        loading="lazy"
        src="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNkYAAAAAYAAjCB0C8AAAAASUVORK5CYII="
      />
      <div class="absolute top-20 flex items-center justify-center" style="width: 100%">
        <Image class="h-12 w-12 text-gray-500 dark:text-gray-400" :stroke-width="1.5" />
      </div>
    </div>
    <div class="flex">
      <button class="flex-grow animate-pulse overflow-hidden p-4 text-left">
        <h3 class="h-4 truncate rounded-full bg-gray-500 text-sm dark:bg-gray-400"></h3>
        <p class="mt-1 h-3 rounded-full bg-gray-400 text-xs dark:bg-gray-500"></p>
      </button>
    </div>
  </div>
</template>
