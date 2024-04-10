import { defineConfig } from 'vitepress'
import VitePressSidebar from 'vitepress-sidebar';

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
    outline: 'deep',
    nav: [],
    sidebar: VitePressSidebar.generateSidebar(vitepressSidebarOptions),

    socialLinks: [
      { icon: 'github', link: 'https://github.com/rgst-io/stencil' }
    ],

    editLink: {
      pattern: 'https://github.com/rgst-io/stencil/edit/main/docs/:path',
    },
    // TODO: Enable once approved.
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
      message: 'Licensed under the AGPL-3.0 License. Maintained by <a href="https://github.com/jaredallard">@jaredallard</a> and <a href="https://github.com/rgst-io/stencil/graphs/contributors">friends</a>.',
      copyright: 'Copyright © 2024 <a href="https://github.com/jaredallard">@jaredallard</a>',
    },
  },
  markdown: {},
  head: [],
})
