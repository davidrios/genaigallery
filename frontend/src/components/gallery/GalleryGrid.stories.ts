import type { Meta, StoryObj } from '@storybook/vue3-vite'
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
      name: 'some image 1',
      created_at: new Date().toISOString(),
    },
    {
      id: '2',
      path: `https://picsum.photos/seed/${seed}1/768/1024`,
      name: 'some image 2',
      created_at: new Date().toISOString(),
    },
    {
      id: '7',
      path: `https://github.com/davidrios/genaigallery/raw/refs/heads/main/backend/testdata/fixtures/video/subfolder/ComfyUI_00001_.mp4`,
      name: 'ComfyUI_00001_.mp4',
      created_at: new Date().toISOString(),
    },
    {
      id: '3',
      path: `https://picsum.photos/seed/${seed}2/1024/1024`,
      name: 'some-image-with-a-really-long-name-to-see-what-happens-in-the-interface.png',
      created_at: new Date().toISOString(),
    },
    {
      id: '4',
      path: `https://picsum.photos/seed/${seed}3/800/600`,
      name: 'some image 4',
      created_at: new Date().toISOString(),
    },
    {
      id: '5',
      path: `https://picsum.photos/seed/${seed}4/768/768`,
      name: 'some image 5',
      created_at: new Date().toISOString(),
    },
    {
      id: '6',
      path: `https://picsum.photos/seed/${seed}6/2048/2048`,
      name: 'some image 6',
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
