<script setup lang="ts">
import { isVideo } from '@/lib/utils'
import type { Image } from '@/types'
import { onUnmounted } from 'vue'
import { Play } from 'lucide-vue-next'

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
      class="aspect-w-1 aspect-h-1 xl:aspect-w-7 xl:aspect-h-8 relative w-full overflow-hidden bg-gray-200 dark:bg-gray-700"
    >
      <template v-if="isVideo(image.path)">
        <img
          class="h-full max-h-[50vh] w-full object-contain object-center transition-opacity duration-300 group-hover:opacity-75"
          loading="lazy"
          :src="
            image.video_preview ||
            'data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNkYAAAAAYAAjCB0C8AAAAASUVORK5CYII='
          "
        />
        <div
          class="absolute flex items-center justify-center"
          style="width: 100%; top: calc(50% - 24px)"
        >
          <Play
            class="h-12 w-12 rounded-md bg-gray-200 p-1 text-gray-500 opacity-75 dark:bg-gray-700 dark:text-gray-400"
            :stroke-width="1.5"
          />
        </div>
      </template>
      <img
        v-else
        :src="image.path"
        :alt="image.path"
        class="h-full max-h-[50vh] w-full object-contain object-center transition-opacity duration-300 group-hover:opacity-75"
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
