---
order: 10
---

# Problem Statement

Stencil is a bit of a new paradigm in repository lifecycle management -- it's not just a templating engine, its intention is to use _living templates_. Existing templating tools are great at the conception of a project, allowing you to safely start from a known-good template to start your codebase and build out from there. However, as time moves on, two things tend to happen:

1. There are updates in the code from whoever built out the template in the first place (often a dedicated tooling team at larger companies).
2. The project using the template grows up and wants to add more (possibly-templated) functionality, but they didn't select the options at the time they built their project, so they usually end up making a new project with the right options and trying to shoehorn the new code into their existing project.

Both of these problems are solved by stencil, which evolve the thinking towards living templates that help all your projects evolve over time together. At larger companies, where you might be building dozens or even hundreds of microservices, having a single resilient and reliable base to build all of your services off is hugely helpful to save your developers time. It's even more helpful when that resilient base evolves regularly with patches, functionality, observability, and more. As template owners grow their templates and support more features, your users can immediately pick up those features, often times doing so in a completely automated fashion -- apply the latest stencil templates, have the PR run unit tests, and auto-merge if everything passes.
