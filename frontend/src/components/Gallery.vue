<script setup lang="ts">
import { computed } from 'vue'
import { refDebounced } from '@vueuse/core'

import { useGalleryData } from '@/composables/useGalleryData'
import { useGalleryNavigation } from '@/composables/useGalleryNavigation'
import { useMediaOverlay } from '@/composables/useMediaOverlay'
import GalleryHeader from './gallery/GalleryHeader.vue'
import GalleryGrid from './gallery/GalleryGrid.vue'
import MediaOverlay from './gallery/MediaOverlay.vue'

const { searchParams, breadcrumbs, navigateTo, toggleSort, performSearch, navigateToPage } =
  useGalleryNavigation()

const { isFetching, error, data } = useGalleryData(searchParams)

const directories = computed(() => data.value?.directories || [])
const images = computed(() => data.value?.images || [])
const isLoading = refDebounced(isFetching, 500)

const {
  selectedImage,
  isLoadingDetails,
  hasPrevious,
  hasNext,
  openImage,
  closeOverlay,
  navigateImage,
} = useMediaOverlay(images)
</script>

<template>
  <div class="container mx-auto min-h-screen p-4">
    <GalleryHeader
      :breadcrumbs="breadcrumbs"
      v-model:searchQuery="searchParams.q"
      :sortOrder="searchParams.sort"
      @navigate="navigateTo"
      @search="performSearch"
      @toggleSort="toggleSort"
    />

    <div class="flex h-20 items-center justify-center" v-if="isLoading">
      <div class="h-8 w-8 animate-spin rounded-full border-b-2 border-indigo-600"></div>
    </div>
    <template v-else>
      <GalleryGrid
        :directories="directories"
        :images="images"
        :error="error"
        @navigate="navigateTo"
        @selectImage="openImage"
      />
      <!-- Paginator here -->
    </template>

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
