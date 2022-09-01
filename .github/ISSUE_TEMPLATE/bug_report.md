---
name: Bug report
about: Create a report to help us improve the Helm Operator
title: ''
labels: blocked needs validation, bug
assignees: ''

---

<!--

# ---- NOTICE -----

Helm Operator v1 is in maintenance mode and Flux v2 is getting
closer to GA. If you want to learn about migrating to Flux v2,
please review <https://github.com/fluxcd/flux2/discussions/413>.

As it will take longer until we get around to issues and PRs in
Helm Operator v1, we strongly recommend that you start
familiarising yourself with Flux v2: <https://fluxcd.io/>

This means that new features will only be added after very careful
consideration, if at all. Refer to the links above for more detail.

# ---- END NOTICE -----

-->

**Describe the bug**

A clear and concise description of what the bug is.

**To Reproduce**

Steps to reproduce the behaviour:
1. Provide the Helm Operator install instructions
2. Provide a HelmRelease example
3. Post the HelmRelease status, you can get this by running `kubectl describe helmrelease <name>`

**Expected behavior**

A clear and concise description of what you expected to happen.

**Logs**

If applicable, please provide logs. In a standard stand-alone installation, you'd get this by running `kubectl logs deploy/flux-helm-operator -n fluxcd`.

**Additional context**

- Helm Operator version: 
- Kubernetes version:
- Git provider:
- Helm repository provider:
