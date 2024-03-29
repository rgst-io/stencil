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
      text: Getting Started
      link: /getting-started
    - theme: alt
      text: About
      link: /about
    - theme: alt
      text: GitHub
      link: https://github.com/rgst-io/stencil

features:
  - title: <a href="/getting-started">Full Lifecycle Management</a>
    details: stencil goes further than other templating tools by enabling you to manage the full lifecycle of your repositories, including re-running,
  - title: <a href="/native-extensions.html">Native Extensions</a>
    details: Need to interface with an API or implement custom parsing/merging logic? Stencil supports native extensions in _any_ language to implement that logic.
  - title: <a href="/tasks/">Easy Templating</a> <Badge type="warning" text="experimental" />
    details: stencil uses Go's templating engine to provide a powerful and easy-to-use templating language for your repositories. Chances are, you already use it in your stack!
---
