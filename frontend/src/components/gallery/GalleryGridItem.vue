<script setup lang="ts">
import { isVideo } from '@/lib/utils'
import type { Image } from '@/types'
import { onUnmounted } from 'vue'

const props = defineProps<{
  image: Image
  onUnmounted: () => void
}>()

const emit = defineEmits<{
  (e: 'selectImage', image: Image): void
}>()

onUnmounted(() => {
  props.onUnmounted()
})
</script>

<template>
  <div
    class="group relative overflow-hidden rounded-xl bg-white shadow-md transition-all duration-300 hover:shadow-xl dark:bg-gray-800"
  >
    <div
      class="aspect-w-1 aspect-h-1 xl:aspect-w-7 xl:aspect-h-8 w-full overflow-hidden bg-gray-200 dark:bg-gray-700"
    >
      <video
        v-if="isVideo(image.path)"
        :src="image.path"
        controls
        preload="metadata"
        class="h-full w-full bg-black object-cover object-center"
      ></video>
      <img
        v-else
        :src="image.path"
        :alt="image.path"
        class="h-full w-full object-cover object-center transition-opacity duration-300 group-hover:opacity-75"
        loading="lazy"
      />
    </div>
    <div class="flex">
      <button
        class="flex-grow overflow-hidden p-4 text-left"
        @click.stop="emit('selectImage', image)"
      >
        <h3 class="truncate text-sm text-gray-500 dark:text-gray-400">
          {{ image.name }}
        </h3>
        <p class="mt-1 text-xs text-gray-400 dark:text-gray-500">
          {{ new Date(image.created_at).toLocaleDateString() }}
        </p>
      </button>
    </div>

    <button
      v-if="!isVideo(image.path)"
      class="absolute inset-0 flex items-center justify-center bg-black/60 opacity-0 transition-opacity duration-300 group-hover:opacity-100"
      @click.stop="emit('selectImage', image)"
    >
      <span
        class="rounded-full bg-white px-4 py-2 font-medium text-black transition-colors hover:bg-gray-100"
      >
        View Details
      </span>
    </button>
  </div>
</template>
