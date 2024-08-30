---
# https://vitepress.dev/reference/default-theme-home-page
layout: home
title: Home

hero:
  name: stencil
  tagline: |
    A modern living-template engine for evolving repositories
  image:
    src: /logo.png
    alt: stencil
  actions:
    - theme: brand
      text: About
      link: /about/problem-statement.html
    - theme: alt
      text: Guide
      link: /guide/installation.html
    - theme: alt
      text: GitHub
      link: https://github.com/rgst-io/stencil

features:
  - icon: 📝
    title: <a href="/about">Development Lifecycle Management</a>
    details: Treat your generated files as APIs by persisting changes in customizable "blocks" to allow rendering more than once
  - icon: 🧱
    title: <a href="/reference/modules.html">Modular</a>
    details: Templates can be composed through importable modules allowing easy customization
  - icon: 🛠️
    title: <a href="/reference/native-extensions.html">Native Extensions</a>
    details: Need to interface with an API or implement custom parsing/merging logic? Stencil supports native extensions in <i>any</i> language to implement that logic
---
