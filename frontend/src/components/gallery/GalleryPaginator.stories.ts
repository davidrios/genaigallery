import type { Meta, StoryObj } from '@storybook/vue3'
import GalleryPaginator from './GalleryPaginator.vue'
import { ref } from 'vue'

const meta = {
  title: 'Gallery/GalleryPaginator',
  component: GalleryPaginator,
  tags: ['autodocs'],
  argTypes: {
    currentPage: { control: 'number' },
    totalPages: { control: 'number' },
  },
} satisfies Meta<typeof GalleryPaginator>

export default meta
type Story = StoryObj<typeof meta>

export const Default: Story = {
  args: {
    currentPage: 1,
    totalPages: 10,
  },
  render: (args) => ({
    components: { GalleryPaginator },
    setup() {
      const currentPage = ref(args.currentPage)
      return { args, currentPage }
    },
    template: `
      <div class="p-4 bg-zinc-50 dark:bg-zinc-900 rounded-lg">
        <GalleryPaginator
          :currentPage="currentPage"
          :totalPages="args.totalPages"
          @update:currentPage="currentPage = $event"
        />
        <div class="mt-4 text-center text-sm text-zinc-500 dark:text-zinc-400">
          Showing page {{ currentPage }} of {{ args.totalPages }}
        </div>
      </div>
    `,
  }),
}

export const FewPages: Story = {
  args: {
    currentPage: 1,
    totalPages: 3,
  },
  render: Default.render,
}

export const MiddlePage: Story = {
  args: {
    currentPage: 5,
    totalPages: 10,
  },
  render: Default.render,
}

export const LastPage: Story = {
  args: {
    currentPage: 10,
    totalPages: 10,
  },
  render: Default.render,
}
