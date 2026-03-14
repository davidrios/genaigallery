import type { Meta, StoryObj } from '@storybook/vue3'
import GalleryGrid from './GalleryGrid.vue'
const meta = {
  title: 'Gallery/GalleryGrid',
  component: GalleryGrid,
  tags: ['autodocs'],
} satisfies Meta<typeof GalleryGrid>

export default meta
type Story = StoryObj<typeof meta>

function getSampleImages(seed: string) {
  const res = [
    {
      id: '1',
      path: `https://picsum.photos/seed/${seed}/1024/768`,
      created_at: new Date().toISOString(),
    },
    {
      id: '2',
      path: `https://picsum.photos/seed/${seed}1/768/1024`,
      created_at: new Date().toISOString(),
    },
    {
      id: '7',
      path: `https://github.com/davidrios/genaigallery/raw/refs/heads/main/backend/testdata/fixtures/video/ComfyUI_00001_.mp4`,
      created_at: new Date().toISOString(),
    },
    {
      id: '3',
      path: `https://picsum.photos/seed/${seed}2/1024/1024`,
      created_at: new Date().toISOString(),
    },
    {
      id: '4',
      path: `https://picsum.photos/seed/${seed}3/800/600`,
      created_at: new Date().toISOString(),
    },
    {
      id: '5',
      path: `https://picsum.photos/seed/${seed}4/768/768`,
      created_at: new Date().toISOString(),
    },
    {
      id: '6',
      path: `https://picsum.photos/seed/${seed}6/2048/2048`,
      created_at: new Date().toISOString(),
    },
  ]

  return res
}

export const Default: Story = {
  args: {
    directories: [
      { name: 'Cyberpunk', path: '/cyberpunk' },
      { name: 'Portraits', path: '/portraits' },
    ],
    images: getSampleImages('default'),
    error: null,
  },
}

export const NoImages: Story = {
  args: {
    directories: [
      { name: 'Cyberpunk', path: '/cyberpunk' },
      { name: 'Portraits', path: '/portraits' },
    ],
    images: [],
    error: null,
  },
}

export const NoDirectories: Story = {
  args: {
    directories: [],
    images: getSampleImages('no-dirs'),
    error: null,
  },
}

export const AllEmpty: Story = {
  args: {
    directories: [],
    images: [],
    error: null,
  },
}

export const ErrorState: Story = {
  args: {
    ...Default.args,
    error: 'Failed to load gallery data. Please try again later.',
  },
}
