import { ref, computed, onMounted, onUnmounted, type Ref, watch } from 'vue'
import type { Image } from '@/types'
import { useFetch, useUrlSearchParams } from '@vueuse/core'
import type { PagesFetchedEvent } from 'wc-infinite-scroller'

export function useMediaOverlay(pagesFetched: Ref<PagesFetchedEvent<Image>['detail'] | null>) {
  function getPageImages(page: string | number) {
    return indexedPages.value[page]?.items ?? []
  }

  const indexedPages = computed(() => {
    return Object.fromEntries(
      (pagesFetched.value?.pages ?? []).map((page) => [page.pageNum, page.pageResult]),
    )
  })
  const images = computed(() => getPageImages(pagesFetched.value?.mainPage ?? ''))
  const currentPage = computed(() => pagesFetched.value?.mainPage ?? 0)
  const wantPage = ref(currentPage.value)

  const selectedImage = ref<Image | null>(null)
  const searchParams = useUrlSearchParams('hash', {
    removeFalsyValues: true,
    writeMode: 'replace',
  })

  const url = computed(() => {
    if (!searchParams.view) {
      return ''
    }
    return `/api/image/${searchParams.view}`
  })

  const { isFetching, data, execute } = useFetch(url, { immediate: false }).json<Image>()

  watch(data, (newData) => {
    selectedImage.value = newData
  })

  watch(
    url,
    (newUrl) => {
      if (newUrl) {
        execute()
      } else {
        selectedImage.value = null
      }
    },
    { immediate: true },
  )

  const openImage = async (image: Image) => {
    selectedImage.value = image
    searchParams.view = image.id
    if (currentImageIndex.value === -1) {
      for (const page of [currentPage.value + 1, currentPage.value - 1]) {
        if (getPageImages(page).findIndex(findImageIndex) !== -1) {
          wantPage.value = page
          break
        }
      }
    }
  }

  const closeOverlay = () => {
    selectedImage.value = null
    searchParams.view = ''
  }

  function findImageIndex(img: Image) {
    return img.id === selectedImage.value?.id || img.path === selectedImage.value?.path
  }

  const currentImageIndex = computed(() => {
    if (!selectedImage.value) return -1
    return images.value.findIndex(findImageIndex)
  })

  const navigateImage = (dir: 'next' | 'prev') => {
    if (currentImageIndex.value === -1) return
    const newIdx = dir === 'next' ? currentImageIndex.value + 1 : currentImageIndex.value - 1
    let nextImg

    if (newIdx < 0 && currentPage.value > 1) {
      wantPage.value = currentPage.value - 1
      const prevImgs = getPageImages(wantPage.value)
      nextImg = prevImgs[prevImgs.length - 1]
    } else if (newIdx >= images.value.length) {
      const nextPage = getPageImages(currentPage.value + 1)
      if (nextPage.length > 0) {
        wantPage.value = currentPage.value + 1
        nextImg = nextPage[0]
      }
    } else if (newIdx >= 0 && newIdx < images.value.length) {
      nextImg = images.value[newIdx]
    }

    if (nextImg) openImage(nextImg)
  }

  const handleKeydown = (e: KeyboardEvent) => {
    if (!selectedImage.value) return
    if (e.key === 'ArrowLeft') navigateImage('prev')
    else if (e.key === 'ArrowRight') navigateImage('next')
    else if (e.key === 'Escape') closeOverlay()
  }

  onMounted(() => {
    window.addEventListener('keydown', handleKeydown)
  })

  onUnmounted(() => {
    window.removeEventListener('keydown', handleKeydown)
  })

  return {
    selectedImage,
    isLoadingDetails: isFetching,
    openImage,
    closeOverlay,
    navigateImage,
    hasPrevious: computed(
      () => currentImageIndex.value > 0 || (pagesFetched.value?.mainPage ?? 0) > 1,
    ),
    hasNext: computed(
      () =>
        (currentImageIndex.value >= 0 && currentImageIndex.value < images.value.length - 1) ||
        getPageImages(currentPage.value + 1).length > 0,
    ),
    wantPage,
  }
}
