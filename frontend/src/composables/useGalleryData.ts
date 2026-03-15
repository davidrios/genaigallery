import { computed, type Ref } from 'vue'
import type { BrowseResponse } from '@/types'
import type { SearchParams } from './useGalleryNavigation'
import { useFetch } from '@vueuse/core'

export function useGalleryData(searchParams: Ref<SearchParams>) {
  const url = computed(() => {
    const apiParams = new URLSearchParams({
      path: searchParams.value.path,
      q: searchParams.value.q,
      sort: searchParams.value.sort,
      page: searchParams.value.page,
      inPath: searchParams.value.inPath,
    })

    return `/api/browse?${apiParams.toString()}`
  })

  return useFetch(url, { refetch: true }).json<BrowseResponse>()
}
