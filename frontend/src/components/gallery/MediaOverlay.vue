<script setup lang="ts">
import { computed } from 'vue'
import type { Image } from '@/types'
import { isVideo } from '@/lib/utils'
import { X, ChevronLeft, ChevronRight, Search } from 'lucide-vue-next'
import { refDebounced } from '@vueuse/core'

const props = defineProps<{
  selectedImage: Image | null
  isLoadingDetails: boolean
  hasPrevious: boolean
  hasNext: boolean
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'navigate', dir: 'next' | 'prev'): void
  (e: 'navigate-path', path: string): void
  (e: 'search', query: string): void
}>()

const isLoading = refDebounced(
  computed(() => props.isLoadingDetails),
  300,
)

const hasMetadata = refDebounced(
  computed(
    () => props.selectedImage?.metadata_items && props.selectedImage.metadata_items.length > 0,
  ),
  300,
)

const sortedMetadata = computed(() => {
  if (!props.selectedImage?.metadata_items) return []
  const grouped: Record<string, string[]> = {}
  for (const item of props.selectedImage.metadata_items) {
    if (!grouped[item.key]) grouped[item.key] = []
    const group = grouped[item.key]
    if (group) group.push(item.value)
  }
  return Object.keys(grouped)
    .sort()
    .map((key) => ({ key, values: grouped[key] }))
})
</script>

<template>
  <div
    v-if="selectedImage"
    class="fixed inset-0 z-50 flex items-center justify-center bg-black/90"
    @click="emit('close')"
  >
    <!-- Close Button -->
    <button
      class="absolute top-4 right-4 z-50 p-2 text-white hover:text-gray-300"
      @click="emit('close')"
    >
      <X class="h-8 w-8" />
    </button>

    <div class="mx-auto flex h-full w-full max-w-7xl gap-6 p-4" @click.stop>
      <!-- Media Viewer -->
      <div
        class="group/media relative flex flex-1 items-center justify-center overflow-hidden rounded-lg bg-black/50"
      >
        <!-- Navigation Buttons -->
        <button
          v-if="hasPrevious"
          class="absolute top-1/2 left-4 z-10 -translate-y-1/2 rounded-full bg-black/40 p-3 text-white opacity-0 backdrop-blur-sm transition-all duration-300 group-hover/media:opacity-100 hover:bg-black/70"
          title="Previous (Left Arrow)"
          @click.stop="emit('navigate', 'prev')"
        >
          <ChevronLeft class="h-6 w-6" />
        </button>

        <button
          v-if="hasNext"
          class="absolute top-1/2 right-4 z-10 -translate-y-1/2 rounded-full bg-black/40 p-3 text-white opacity-0 backdrop-blur-sm transition-all duration-300 group-hover/media:opacity-100 hover:bg-black/70"
          title="Next (Right Arrow)"
          @click.stop="emit('navigate', 'next')"
        >
          <ChevronRight class="h-6 w-6" />
        </button>

        <video
          v-if="isVideo(selectedImage.path)"
          :src="selectedImage.path"
          controls
          autoplay
          class="max-h-full max-w-full object-contain"
        ></video>
        <img
          v-else
          :src="selectedImage.path"
          :alt="selectedImage.path"
          class="max-h-full max-w-full object-contain"
        />
      </div>

      <!-- Sidebar -->
      <div
        class="flex w-96 flex-col overflow-hidden rounded-r-lg border-l border-gray-800 bg-gray-900 shadow-xl transition-all duration-300"
        :class="{ 'w-0 opacity-0': !selectedImage }"
      >
        <div class="sticky top-0 border-b border-gray-800 bg-gray-900/95 p-4">
          <button
            class="mt-1 text-left text-sm text-gray-500"
            @click.stop="emit('navigate-path', selectedImage.base_path)"
            @click="emit('close')"
          >
            {{ selectedImage.base_path }}/
          </button>
          <h2 class="truncate text-lg font-semibold text-gray-100" :title="selectedImage.path">
            {{ selectedImage.name }}
          </h2>
          <p class="mt-1 text-sm text-gray-500">
            {{ new Date(selectedImage.created_at).toLocaleString() }}
          </p>
        </div>

        <div class="custom-scrollbar flex-1 space-y-4 overflow-y-auto p-4">
          <div v-if="isLoading" class="py-8 text-center">
            <div
              class="mx-auto h-8 w-8 animate-spin rounded-full border-b-2 border-indigo-500"
            ></div>
          </div>

          <div v-else-if="hasMetadata">
            <h3 class="mb-3 text-xs font-bold tracking-wider text-gray-500 uppercase">
              Generation Params
            </h3>
            <div class="space-y-3">
              <div v-for="item in sortedMetadata" class="group">
                <dt
                  class="mb-1 flex items-center justify-between text-xs font-medium break-all text-indigo-400"
                >
                  <span>{{ item.key }}</span>
                  <button
                    class="ml-2 rounded p-1 text-gray-500 transition-colors hover:bg-gray-800 hover:text-indigo-300"
                    title="Search for this value"
                    @click.stop="emit('search', `${item.key}:${item.values?.[0]}`)"
                    @click="emit('close')"
                  >
                    <Search class="h-3 w-3" />
                  </button>
                </dt>
                <dd
                  class="rounded border border-transparent bg-gray-800/50 p-2 font-mono text-sm break-words text-gray-300 transition-colors"
                >
                  <ul v-if="item.values && item.values.length > 1" class="list-inside list-disc">
                    <li v-for="(val, idx) in item.values" :key="idx">{{ val }}</li>
                  </ul>
                  <span v-else>{{ item.values?.[0] }}</span>
                </dd>
              </div>
            </div>
          </div>

          <div v-else class="py-10 text-center text-gray-600 italic">
            No metadata available for this image.
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
