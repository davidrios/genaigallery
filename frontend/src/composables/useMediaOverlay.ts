import { ref, computed, onMounted, onUnmounted, type Ref, watch } from 'vue'
import type { Image } from '@/types'
import { useFetch, useUrlSearchParams } from '@vueuse/core'

export function useMediaOverlay(images: Ref<Image[]>) {
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
      }
    },
    { immediate: true },
  )

  const openImage = async (image: Image) => {
    selectedImage.value = image
    searchParams.view = image.id
  }

  const closeOverlay = () => {
    selectedImage.value = null
    searchParams.view = ''
  }

  const currentImageIndex = computed(() => {
    if (!selectedImage.value) return -1
    return images.value.findIndex(
      (img) => img.id === selectedImage.value?.id || img.path === selectedImage.value?.path,
    )
  })

  const navigateImage = (dir: 'next' | 'prev') => {
    if (currentImageIndex.value === -1) return
    const newIdx = dir === 'next' ? currentImageIndex.value + 1 : currentImageIndex.value - 1
    if (newIdx >= 0 && newIdx < images.value.length) {
      const nextImg = images.value[newIdx]
      if (nextImg) openImage(nextImg)
    }
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
    hasPrevious: computed(() => currentImageIndex.value > 0),
    hasNext: computed(
      () => currentImageIndex.value !== -1 && currentImageIndex.value < images.value.length - 1,
    ),
  }
}
