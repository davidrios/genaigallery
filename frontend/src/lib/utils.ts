import type { ClassValue } from 'clsx'
import { clsx } from 'clsx'
import { twMerge } from 'tailwind-merge'

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

export function isVideo(path: string) {
  const ext = path.split('.').pop()?.toLowerCase()
  return ['mp4', 'webm', 'mov'].includes(ext || '')
}
