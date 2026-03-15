<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { refDebounced } from '@vueuse/core'

import { useGalleryData } from '@/composables/useGalleryData'
import { useGalleryNavigation } from '@/composables/useGalleryNavigation'
import { useMediaOverlay } from '@/composables/useMediaOverlay'
import GalleryHeader from './gallery/GalleryHeader.vue'
import GalleryGrid from './gallery/GalleryGrid.vue'
import GalleryPaginator from './gallery/GalleryPaginator.vue'
import MediaOverlay from './gallery/MediaOverlay.vue'

const { searchParams, breadcrumbs, navigateTo, toggleSort, performSearch, navigateToPage } =
  useGalleryNavigation()

const { isFetching, error, data } = useGalleryData(searchParams)
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

const handlePageChange = (page: number | string) => {
  navigateToPage(page.toString())
  window.scrollTo({ top: 0, behavior: 'instant' })
}

const handleOverlaySearch = (query: string) => {
  performSearch(query)
  window.scrollTo({ top: 0, behavior: 'instant' })
}
</script>

<template>
  <div class="container mx-auto min-h-screen p-4">
    <GalleryHeader
      v-model:search-query="searchParams.q"
      :breadcrumbs="breadcrumbs"
      :sort-order="searchParams.sort"
      @navigate="navigateTo"
      @search="performSearch"
      @toggle-sort="toggleSort"
    />

    <div v-if="isLoading" class="flex h-20 items-center justify-center">
      <div class="h-8 w-8 animate-spin rounded-full border-b-2 border-indigo-600"></div>
    </div>
    <template v-else>
      <GalleryGrid
        :directories="directories"
        :images="images"
        :error="error"
        @navigate="navigateTo"
        @select-image="openImage"
      />

      <div v-if="data?.pages && data.pages > 1" class="mt-8">
        <GalleryPaginator
          :current-page="data.page"
          :total-pages="data.pages"
          @update:current-page="handlePageChange"
        />
      </div>
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
