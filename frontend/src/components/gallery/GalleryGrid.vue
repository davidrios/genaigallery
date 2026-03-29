<script setup lang="ts">
import { Folder } from 'lucide-vue-next'
import {
  InfiniteScroller,
  register,
  type PageChangedEvent,
  type PageResult,
  type PagesFetchedEvent,
} from 'wc-infinite-scroller'

import type { Directory, Image } from '@/types'
import GalleryGridItem from './GalleryGridItem.vue'
import { computed, getCurrentInstance, h, onMounted, ref, render, watch } from 'vue'
import GalleryGridItemSkel from './GalleryGridItemSkel.vue'

register()

const props = defineProps<{
  directories: Directory[]
  images: Image[]
  error: string | null
  fetchPage?: (page: number) => Promise<PageResult<Image>>
  currentPage: string
}>()

const emit = defineEmits<{
  (e: 'navigate', path: string): void
  (e: 'selectImage', image: Image): void
  (e: 'navigateToPage', path: string): void
  (e: 'update:pagesFetched', data: PagesFetchedEvent<Image>['detail']): void
}>()

const error = computed((err) => {
  if (err == null || err === 'The operation was aborted.') {
    return
  }
  return props.error
})

const infiniteScroller = ref<InfiniteScroller<Image> | null>(null)
onMounted(() => {
  if (infiniteScroller.value == null) {
    console.error('unexpected state')
    return
  }

  infiniteScroller.value.fetchPage = props.fetchPage!

  infiniteScroller.value.createPageElement = () => {
    const li = document.createElement('li')
    li.classList.add(
      ...'grid grid-cols-1 gap-6 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4'.split(' '),
    )
    return li
  }

  infiniteScroller.value.createPlaceholderElements = () => {
    const el = document.createElement('div')
    const vnode = h(GalleryGridItemSkel)
    render(vnode, el)
    const skeletons = []
    for (let i = 0; i < 50; i++) {
      const skel = el.firstElementChild?.cloneNode() as HTMLElement
      skel.innerHTML = el.firstElementChild?.innerHTML ?? ''
      skeletons.push(skel)
    }
    render(null, el)
    return skeletons
  }

  infiniteScroller.value.renderItem = (item: Image) => {
    const el = document.createElement('div')
    const vnode = h(GalleryGridItem, {
      image: item,
      onSelectImage(image) {
        emit('selectImage', image)
      },
      onUnmounted() {
        console.log('unmounted!', item.id)
      },
    })
    vnode.appContext = getCurrentInstance()?.appContext ?? null
    render(vnode, el)
    return el
  }

  infiniteScroller.value.addEventListener('page-changed', (e: PageChangedEvent) => {
    if (e.detail == null) {
      return
    }
    emit('navigateToPage', e.detail.page.toString())
  })

  infiniteScroller.value.addEventListener('pages-fetched', (e: PagesFetchedEvent<Image>) => {
    if (e.detail == null) {
      return
    }
    emit('update:pagesFetched', e.detail)
  })

  infiniteScroller.value.currentPage = parseInt(props.currentPage)
  infiniteScroller.value.loadInitialPage()

  watch(
    () => props.currentPage,
    (newPage) => {
      infiniteScroller.value!.currentPage = parseInt(newPage)
    },
  )
})
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
          class="group flex cursor-pointer flex-col items-center justify-center rounded-xl border border-gray-200 bg-gray-50 p-4 transition-all duration-200 hover:border-indigo-400 hover:bg-indigo-50 dark:border-gray-700 dark:bg-gray-800/50 dark:hover:border-indigo-500 dark:hover:bg-indigo-900/20"
          @click="emit('navigate', dir.path)"
        >
          <Folder
            class="h-12 w-12 text-indigo-400 transition-transform group-hover:scale-110 dark:text-indigo-300"
            :stroke-width="1.5"
          />
          <span
            class="mt-2 w-full truncate text-center text-sm font-medium text-gray-700 dark:text-gray-200"
            >{{ dir.name }}</span
          >
        </div>
      </div>

      <!-- Pages Content -->
      <div class="flex flex-col gap-6">
        <div class="relative">
          <infinite-scroller ref="infiniteScroller">
            <ul class="dark grid gap-6"></ul>
            <div class="grid grid-cols-1 gap-6 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4"></div>
          </infinite-scroller>
        </div>
      </div>

      <!-- Empty State -->
      <div
        v-if="directories.length === 0 && images.length == 0"
        class="py-20 text-center text-gray-500"
      >
        No results found.
      </div>
    </template>
  </div>
</template>
