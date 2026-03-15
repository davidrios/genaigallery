import type { Meta, StoryObj } from '@storybook/vue3-vite'
import MediaOverlay from './MediaOverlay.vue'
const meta = {
  title: 'Gallery/MediaOverlay',
  component: MediaOverlay,
  tags: ['autodocs'],
  argTypes: {
    isLoadingDetails: { control: 'boolean' },
    hasPrevious: { control: 'boolean' },
    hasNext: { control: 'boolean' },
  },
  args: {},
} satisfies Meta<typeof MediaOverlay>

export default meta
type Story = StoryObj<typeof meta>

const sampleImageWithMetadata = {
  id: '1',
  name: 'some image',
  basePath: 'comfyui',
  path: `https://picsum.photos/seed/seed1/1024/768`,
  created_at: new Date().toISOString(),
  metadata_items: [
    { key: 'prompt', value: 'a cyberpunk city street, neon lights, rainy, 8k, masterpiece' },
    { key: 'negative_prompt', value: 'ugly, blurry, low res' },
    { key: 'seed', value: '123456789' },
    { key: 'sampler', value: 'Euler a' },
    { key: 'steps', value: '20' },
  ],
}

export const Default: Story = {
  args: {
    selectedImage: sampleImageWithMetadata,
    isLoadingDetails: false,
    hasPrevious: true,
    hasNext: true,
  },
}

export const LoadingDetails: Story = {
  args: {
    ...Default.args,
    selectedImage: {
      ...sampleImageWithMetadata,
      metadata_items: undefined,
      path: `https://picsum.photos/seed/seed2/768/1024`,
    },
    isLoadingDetails: true,
    hasPrevious: false,
  },
}

export const NoMetadata: Story = {
  args: {
    ...Default.args,
    selectedImage: {
      ...sampleImageWithMetadata,
      metadata_items: [],
      path: `https://picsum.photos/seed/seed3/768/768`,
    },
    hasNext: false,
  },
}

export const WithVideo: Story = {
  args: {
    ...Default.args,
    selectedImage: {
      ...sampleImageWithMetadata,
      metadata_items: undefined,
      path: `https://github.com/davidrios/genaigallery/raw/refs/heads/main/backend/testdata/fixtures/video/subfolder/ComfyUI_00001_.mp4`,
    },
  },
}
