---
name: Bug report
about: Create a report to help us improve the Helm Operator
title: ''
labels: blocked needs validation, bug
assignees: ''

---

<!--
NOTE: the Helm Operator is in maintenance mode, and it will take a bit
      longer until we get around to issues and PRs.
      
      For more information, and details about the Helm Operator's
      future, see: https://github.com/fluxcd/helm-operator/issues/546
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
