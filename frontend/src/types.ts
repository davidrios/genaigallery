export interface Image {
  id: string
  base_path: string
  path: string
  name: string
  created_at: string
  metadata_items?: { key: string; value: string }[]
  video_preview?: string
}

export interface Directory {
  path: string
  name: string
}

export interface BrowseResponse {
  directories: Directory[]
  images: Image[]
  total: number
  page: number
  pages: number
}
