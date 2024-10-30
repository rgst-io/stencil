---
order: 5
---

# Adopt Existing Projects

Often times, your projects have already been built using a template or code generation of some sort, so there is some level of standardization in place. We have added some simple heuristics to stencil to detect what customization content should go inside blocks as you adopt these existing projects into stencil. This system is designed to help automate bringing stencil into established codebases and save the majority of the time taken to manually reconcile code that goes into blocks as you one-by-one convert your projects to use stencil.

## Usage

Once you have created a stencil.yaml file in an existing project's repository directory, you can run `stencil --adopt` to enable the heuristics to detect block content. Do not run --adopt multiple times, as it will multiply the block start/end lines. Run it once, compare the diff, and do any remaining manual reconciliation needed.

## Example

Let's say you have a template for a YAML file that looks like:

```yaml
global:
  deploymentEnvironment: prod
## <<Stencil::Block(version)>>
{{- if empty (trim (file.Block "version")) }}
  version: latest
{{- else }}
{{ file.Block "version"}}
{{- end }}
## <</Stencil::Block>>
somechart:
  somestuff: 4
```

And you have an existing YAML file that looks like:

```yaml
global:
  deploymentEnvironment: prod
  version: xyz
somechart:
  somestuff: 4
```

If you run `stencil --adopt`, then the heuristics will detect the line in between `deploymentEnvironment: prod` and `somechart:` (the lines above and below the block start/finish) and insert `version: xyz` into the contents of that block during the first render, and you'll end up with the following file:

```yaml
global:
  deploymentEnvironment: prod
  ## <<Stencil::Block(version)>>
  version: xyz
## <</Stencil::Block>>
somechart:
  somestuff: 4
```

The heuristics look for increasing numbers of lines above and below the blocks until it gets down to the point of only having one matching potential-block, and then fills that block in. If you find other example scenarios that are not handled well by this, feel free to open an issue on [GitHub](https://github.com/rgst-io/stencil/issues) with the example, and we'll see what we can do to improve it.
