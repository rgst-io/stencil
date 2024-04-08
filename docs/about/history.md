---
order: 20
---

# History

Stencil was created by developers at [Outreach](https://www.outreach.io/) in 2020, and originally named `bootstrap`, for obvious reasons -- if you wanted to bootstrap a new project, we wanted to automate that process for our developers. We were just entering an era of monolith decomposition -- starting to build microservices en masse to extract functionality out of the monolith -- and wanted to keep our feature developers from diverging too heavily from each other. We quickly saw that basic templating didn't get us very far -- our users immediately went off the rails to accomplish what they needed since bootstrap was no help after creation. We quickly came up with the idea of _blocks_ to provide an explicit interface for where we wanted users to extend, separating it from code that we wanted to expressly forbid them from changing.

As the project grew and grew, we took our learnings from `bootstrap` and created a new project from scratch called `stencil`, which we intended to be open sourced from day 1, and included a generic module system from the start. This allowed our different tooling teams to build multiple modules of independent templates, and consumers using stencil could independently version those modules as they wanted to update (or roll back if a bad template release happened).

Stencil is now maintained as an open source project, no longer affiliated with Outreach, by some of the original developers of the tool.
