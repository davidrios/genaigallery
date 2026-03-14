import { useDebounceFn, useUrlSearchParams } from '@vueuse/core'
import { computed } from 'vue'

export interface SearchParams {
  path: string
  q: string
  sort: 'desc' | 'asc'
  page: string
}

export function useGalleryNavigation() {
  const searchParams = useUrlSearchParams('history', {
    removeFalsyValues: true,
    writeMode: 'push',
  })

  const breadcrumbs = computed(() => {
    if (!searchParams.path) return []
    const parts = (searchParams.path as string).split('/')
    let accum = ''
    return parts.map((part) => {
      accum = accum ? `${accum}/${part}` : part
      return { name: part, path: accum }
    })
  })

  const navigateTo = (path: string) => {
    searchParams.path = path
    searchParams.q = ''
    searchParams.page = '1'
  }

  const navigateToPage = (page: string) => {
    searchParams.page = page
  }

  const toggleSort = () => {
    searchParams.sort = searchParams.sort === 'desc' ? 'asc' : 'desc'
  }

  const performSearch = useDebounceFn((queryVal: string) => {
    searchParams.q = queryVal
    searchParams.page = '1'
  }, 500)

  return {
    searchParams: computed<SearchParams>(
      () =>
        ({
          path: searchParams.path ?? '',
          q: searchParams.q ?? '',
          sort: searchParams.sort ?? 'desc',
          page: searchParams.page ?? '1',
        }) as SearchParams,
    ),
    breadcrumbs,
    navigateTo,
    toggleSort,
    performSearch,
    navigateToPage,
  }
}
