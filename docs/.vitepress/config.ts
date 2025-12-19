import { defineConfig } from 'vitepress'
import { generateSidebar } from 'vitepress-sidebar';

const vitepressSidebarOptions = {
  documentRootPath: '/',
  useTitleFromFileHeading: true,
  sortMenusByFrontmatterOrder: true,
  useFolderTitleFromIndexFile: true,
  collapsed: true,
};

// https://vitepress.dev/reference/site-config
export default defineConfig({
  title: "stencil",
  description: "stencil documentation",
  lang: 'en-US',
  lastUpdated: true,
  appearance: 'dark',
  sitemap: {
    hostname: 'https://stencil.rgst.io',
  },
  themeConfig: {
    // https://vitepress.dev/reference/default-theme-config
		logo: 'logo.png',
    outline: 'deep',
    nav: [],
    sidebar: generateSidebar(vitepressSidebarOptions),

    socialLinks: [
      { icon: 'github', link: 'https://github.com/rgst-io/stencil' }
    ],

    editLink: {
      pattern: 'https://github.com/rgst-io/stencil/edit/main/docs/:path',
    },
    search: {
      provider: 'algolia',
      options: {
        indexName: 'stencil-rgst',
        appId: 'AMQEFIC433',
        apiKey: '8f907831b792edbc9d1fe9e951324346',
        insights: true,
      }
    },
    footer: {
      message: 'Licensed under the Apache-2.0 License. Maintained by <a href="https://github.com/jaredallard">@jaredallard</a> and <a href="https://github.com/rgst-io/stencil/graphs/contributors">friends</a>.',
      copyright: 'Copyright © 2025 <a href="https://github.com/jaredallard">@jaredallard</a>',
    },
  },
  markdown: {},
	head: [
		['meta', { property: 'og:type', content: 'website' }],
		['meta', { property: 'og:title', content: 'Stencil' }],
		['meta', { property: 'og:image', content: 'https://stencil.rgst.io/logo.png' }],
		['meta', { property: 'og:description', content: 'A modern living-template engine for evolving code repositories' }],
		['meta', { name: 'twitter:card', content: 'summary_large_image' }],
		['link', { rel: 'apple-touch-icon', sizes: '57x57', href: '/apple-icon-57x57.png' }],
		['link', { rel: 'apple-touch-icon', sizes: '60x60', href: '/apple-icon-60x60.png' }],
		['link', { rel: 'apple-touch-icon', sizes: '72x72', href: '/apple-icon-72x72.png' }],
		['link', { rel: 'apple-touch-icon', sizes: '76x76', href: '/apple-icon-76x76.png' }],
		['link', { rel: 'apple-touch-icon', sizes: '114x114', href: '/apple-icon-114x114.png' }],
		['link', { rel: 'apple-touch-icon', sizes: '120x120', href: '/apple-icon-120x120.png' }],
		['link', { rel: 'apple-touch-icon', sizes: '144x144', href: '/apple-icon-144x144.png' }],
		['link', { rel: 'apple-touch-icon', sizes: '152x152', href: '/apple-icon-152x152.png' }],
		['link', { rel: 'apple-touch-icon', sizes: '180x180', href: '/apple-icon-180x180.png' }],
		['link', { rel: 'icon', type: 'image/png', sizes: '192x192', href: '/android-icon-192x192.png' }],
		['link', { rel: 'icon', type: 'image/png', sizes: '32x32', href: '/favicon-32x32.png' }],
		['link', { rel: 'icon', type: 'image/png', sizes: '96x96', href: '/favicon-96x96.png' }],
		['link', { rel: 'icon', type: 'image/png', sizes: '16x16', href: '/favicon-16x16.png' }],
		['link', { rel: 'manifest', href: '/manifest.json' }],
		['meta', { name: 'msapplication-TileColor', content: '#ffffff' }],
		['meta', { name: 'msapplication-TileImage', content: '/ms-icon-144x144.png' }],
		['meta', { name: 'theme-color', content: '#ffffff' }],
	],
})
