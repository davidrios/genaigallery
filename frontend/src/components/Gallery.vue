<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { refDebounced } from '@vueuse/core'
import type { PageResult } from 'wc-infinite-scroller'

import { fetchBrowse, useGalleryData } from '@/composables/useGalleryData'
import { useGalleryNavigation, type SearchParams } from '@/composables/useGalleryNavigation'
import { useMediaOverlay } from '@/composables/useMediaOverlay'
import GalleryHeader from './gallery/GalleryHeader.vue'
import GalleryGrid from './gallery/GalleryGrid.vue'
import MediaOverlay from './gallery/MediaOverlay.vue'
import type { Image } from '@/types'

const {
  searchParams,
  breadcrumbs,
  navigateTo,
  toggleSort,
  changeInPath,
  performSearch,
  navigateToPage,
} = useGalleryNavigation()

const galleryData = useGalleryData(searchParams)
const { isFetching, error, data } = galleryData
const firstLoad = ref(true)
watch(isFetching, (isFetching) => {
  if (isFetching) {
    firstLoad.value = false
  }
})

const directories = computed(() => data.value?.directories || [])
const images = computed(() => data.value?.images || [])
const isLoading = refDebounced(
  computed(() => firstLoad.value || isFetching.value),
  300,
)

const {
  selectedImage,
  isLoadingDetails,
  hasPrevious,
  hasNext,
  openImage,
  closeOverlay,
  navigateImage,
} = useMediaOverlay(images)

const handleOverlaySearch = (query: string) => {
  performSearch(query)
  window.scrollTo({ top: 0, behavior: 'instant' })
}

async function fetchPage(page: number): Promise<PageResult<Image>> {
  const res = await fetchBrowse({ ...searchParams.value, page: page.toString() })
  const pageResult = {
    currentPage: res.page,
    items: res.images,
    totalPages: res.pages,
  }
  return pageResult
}

const id = ref(1)
const oldParams = ref<SearchParams>({ ...searchParams.value })
watch(searchParams, (newParams) => {
  if (
    newParams.q !== oldParams.value.q ||
    newParams.path !== oldParams.value.path ||
    newParams.sort !== oldParams.value.sort ||
    newParams.inPath !== oldParams.value.inPath
  ) {
    id.value += 1
  }
  oldParams.value = { ...newParams }
})
</script>

<template>
  <div class="container mx-auto min-h-screen p-4">
    <GalleryHeader
      v-model:search-query="searchParams.q"
      v-model:in-path="searchParams.inPath"
      :breadcrumbs="breadcrumbs"
      :sort-order="searchParams.sort"
      @navigate="navigateTo"
      @search="performSearch"
      @toggle-sort="toggleSort"
      @change-in-path="changeInPath"
    />

    <div v-if="isLoading" class="flex h-20 items-center justify-center">
      <div class="h-8 w-8 animate-spin rounded-full border-b-2 border-indigo-600"></div>
    </div>
    <template v-else>
      <GalleryGrid
        :key="id"
        :directories="directories"
        :images="images"
        :error="error"
        :fetch-page="fetchPage"
        :current-page="searchParams.page"
        @navigate="navigateTo"
        @select-image="openImage"
        @navigate-to-page="navigateToPage"
      />
    </template>

    <MediaOverlay
      :selected-image="selectedImage"
      :is-loading-details="isLoadingDetails"
      :has-previous="hasPrevious"
      :has-next="hasNext"
      @close="closeOverlay"
      @navigate="navigateImage"
      @navigatePath="navigateTo"
      @search="handleOverlaySearch"
    />
  </div>
</template>
