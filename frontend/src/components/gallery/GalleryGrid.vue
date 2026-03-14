<script setup lang="ts">
import { isVideo } from '@/lib/utils'
import type { Directory, Image } from '@/types'

defineProps<{
  directories: Directory[]
  images: Image[]
  error: string | null
}>()

const emit = defineEmits<{
  (e: 'navigate', path: string): void
  (e: 'selectImage', image: Image): void
}>()
</script>

<template>
  <div>
    <!-- Error State -->
    <div
      v-if="error"
      class="relative mb-4 rounded border border-red-400 bg-red-100 px-4 py-3 text-red-700 dark:bg-red-900 dark:text-red-200"
      role="alert"
    >
      <strong class="font-bold">Error!</strong>
      {{ error }}
    </div>
    <template v-else>
      <!-- Subdirectories -->
      <div
        v-if="directories.length > 0"
        class="mb-8 grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6"
      >
        <div
          v-for="dir in directories"
          :key="dir.path"
          @click="emit('navigate', dir.path)"
          class="group flex cursor-pointer flex-col items-center justify-center rounded-xl border border-gray-200 bg-gray-50 p-4 transition-all duration-200 hover:border-indigo-400 hover:bg-indigo-50 dark:border-gray-700 dark:bg-gray-800/50 dark:hover:border-indigo-500 dark:hover:bg-indigo-900/20"
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            width="48"
            height="48"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="1.5"
            stroke-linecap="round"
            stroke-linejoin="round"
            class="text-indigo-400 transition-transform group-hover:scale-110 dark:text-indigo-300"
          >
            <path
              d="M4 20h16a2 2 0 0 0 2-2V8a2 2 0 0 0-2-2h-7.93a2 2 0 0 1-1.66-.9l-.82-1.2A2 2 0 0 0 7.93 2H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2Z"
            ></path>
          </svg>
          <span
            class="mt-2 w-full truncate text-center text-sm font-medium text-gray-700 dark:text-gray-200"
            >{{ dir.name }}</span
          >
        </div>
      </div>

      <!-- Pages Content -->
      <div class="flex flex-col gap-6">
        <div class="relative">
          <div class="grid grid-cols-1 gap-6 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4">
            <div
              v-for="image in images"
              :key="image.id"
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
                    {{ image.path.split('/').pop() }}
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
          </div>
        </div>
      </div>

      <!-- Empty State -->
      <div
        v-if="directories.length === 0 && images.length == 0"
        class="py-20 text-center text-gray-500"
      >
        Empty directory.
      </div>
    </template>
  </div>
</template>
