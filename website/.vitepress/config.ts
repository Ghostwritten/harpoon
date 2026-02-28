import { defineConfig } from 'vitepress'

export default defineConfig({
  title: 'Harpoon (hpn)',
  description: 'Modern, efficient container image management CLI. Pull, save, load, and push with Docker, Podman, Nerdctl, and Skopeo.',
  base: '/harpoon/',
  themeConfig: {
    logo: '/logo.png',
    nav: [
      { text: 'Guide', link: '/guide/quickstart' },
      { text: 'Download', link: '/download' },
      { text: 'Changelog', link: '/changelog' },
      { text: 'GitHub', link: 'https://github.com/Ghostwritten/harpoon' },
    ],
    sidebar: {
      '/guide/': [
        {
          text: 'Getting Started',
          items: [
            { text: 'Quick Start', link: '/guide/quickstart' },
            { text: 'Installation', link: '/guide/installation' },
          ],
        },
        {
          text: 'Usage',
          items: [
            { text: 'Examples', link: '/guide/examples' },
            { text: 'Building', link: '/guide/building' },
            { text: 'Concurrency', link: '/guide/concurrency' },
          ],
        },
        {
          text: 'Advanced',
          items: [
            { text: 'Skopeo: Images', link: '/guide/advanced/skopeo-images' },
            { text: 'Skopeo: Pull vs Save', link: '/guide/advanced/skopeo-pull-vs-save' },
          ],
        },
      ],
    },
    socialLinks: [
      { icon: 'github', link: 'https://github.com/Ghostwritten/harpoon' },
    ],
  },
})
