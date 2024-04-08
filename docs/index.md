---
# https://vitepress.dev/reference/default-theme-home-page
layout: home
title: Home

hero:
  name: stencil
  tagline: |
      Modern repository templating engine
  actions:
    - theme: brand
      text: About
      link: /about
    - theme: alt
      text: Guide
      link: /guide/installation.html
    - theme: alt
      text: GitHub
      link: https://github.com/rgst-io/stencil

features:
  - title: <a href="/guide/installation.html">Development Lifecycle Management</a>
    details: stencil goes further than other templating tools by defining extensibility "blocks" to explicitly separate what your consumers can and can't extend, encouraging a system of regularly re-running stencil to pull in living-and-progressing templates.
  - title: <a href="/native-extensions.html">Native Extensions</a>
    details: Need to interface with an API or implement custom parsing/merging logic? Stencil supports native extensions in _any_ language to implement that logic.

---
