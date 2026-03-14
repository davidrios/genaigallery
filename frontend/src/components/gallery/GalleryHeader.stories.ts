import type { Meta, StoryObj } from '@storybook/vue3-vite'
import GalleryHeader from './GalleryHeader.vue'
const meta = {
  title: 'Gallery/GalleryHeader',
  component: GalleryHeader,
  tags: ['autodocs'],
  argTypes: {
    sortOrder: {
      control: 'radio',
      options: ['asc', 'desc'],
    },
    searchQuery: { control: 'text' },
  },
  args: {},
} satisfies Meta<typeof GalleryHeader>

export default meta
type Story = StoryObj<typeof meta>

export const Default: Story = {
  args: {
    breadcrumbs: [
      { name: 'cyberpunk', path: '/cyberpunk' },
      { name: 'cityscapes', path: '/cyberpunk/cityscapes' },
    ],
    searchQuery: '',
    sortOrder: 'desc',
  },
}

export const WithSearchQuery: Story = {
  args: {
    ...Default.args,
    searchQuery: 'neon lights',
  },
}

export const AscendingSort: Story = {
  args: {
    ...Default.args,
    sortOrder: 'asc',
  },
}
