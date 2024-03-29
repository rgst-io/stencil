import { defineConfig } from 'vitepress'

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
    sidebar: [],

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
        indexName: 'stencil',
        appId: '',
        apiKey: '',
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
