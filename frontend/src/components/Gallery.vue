<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useGalleryData } from '@/composables/useGalleryData'
import { useGalleryNavigation } from '@/composables/useGalleryNavigation'
import { useInfiniteScrollLogic } from '@/composables/useInfiniteScrollLogic'
import { useMediaOverlay } from '@/composables/useMediaOverlay'

import GalleryHeader from './gallery/GalleryHeader.vue'
import GalleryGrid from './gallery/GalleryGrid.vue'
import MediaOverlay from './gallery/MediaOverlay.vue'

const {
  pages,
  directories,
  images,
  loading,
  error,
  totalPages,
  firstLoadedPage,
  lastLoadedPage,
  loadPageContent,
  resetData,
  addPageChunk,
} = useGalleryData()

const isUrlSyncing = ref(false)

const initGallery = async () => {
  loading.value = true
  resetData()

  // Get params from route via composable helper or direct access if easier,
  // but better to keep it decoupled.
  // Actually `useGalleryNavigation` has access to `route` so we can read from it inside `loadPageContent` calls
  // OR we just read values here.
  const targetPage = parseInt(route.query.page as string) || 1
  const path = currentPath.value
  const sort = sortOrder.value
  const q = searchQuery.value

  const result = await loadPageContent(path, sort, q, targetPage)

  if (result) {
    addPageChunk(
      { pageNum: targetPage, images: result.images },
      'push', // first chunk is effectively a push
      result.directories,
      result.totalPages,
    )
    firstLoadedPage.value = targetPage
    lastLoadedPage.value = targetPage

    // Handle Deep Link
    const viewPath = route.query.view as string
    if (viewPath) {
      const img = result.images.find((i) => i.path === viewPath)
      if (img) openImage(img, false)
    }
  }
  loading.value = false
}

const {
  currentPath,
  sortOrder,
  searchQuery,
  breadcrumbs,
  updateRoutePage,
  navigateTo,
  toggleSort,
  performSearch,
  route,
} = useGalleryNavigation(initGallery, isUrlSyncing)

const {
  selectedImage,
  isLoadingDetails,
  hasPrevious,
  hasNext,
  openImage,
  closeOverlay,
  navigateImage,
} = useMediaOverlay(images)

onMounted(() => {
  initGallery()
})
</script>

<template>
  <div class="container mx-auto min-h-screen p-4">
    <GalleryHeader
      :breadcrumbs="breadcrumbs"
      v-model:searchQuery="searchQuery"
      :sortOrder="sortOrder"
      @navigate="navigateTo"
      @search="performSearch"
      @toggleSort="toggleSort"
    />

    <GalleryGrid
      :directories="directories"
      :pages="pages"
      :loading="loading"
      :totalPages="totalPages"
      :error="error"
      @navigate="navigateTo"
      @selectImage="openImage"
    />

    <MediaOverlay
      :selectedImage="selectedImage"
      :isLoadingDetails="isLoadingDetails"
      :hasPrevious="hasPrevious"
      :hasNext="hasNext"
      @close="closeOverlay"
      @navigate="navigateImage"
    />
  </div>
</template>
