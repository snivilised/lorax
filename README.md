# 🌟 lorax: ___Go Concurrency with Functional/Reactive Extensions___

[![A B](https://img.shields.io/badge/branching-commonflow-informational?style=flat)](https://commonflow.org)
[![A B](https://img.shields.io/badge/merge-rebase-informational?style=flat)](https://git-scm.com/book/en/v2/Git-Branching-Rebasing)
[![A B](https://img.shields.io/badge/branch%20history-linear-blue?style=flat)](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/defining-the-mergeability-of-pull-requests/managing-a-branch-protection-rule)
[![Go Reference](https://pkg.go.dev/badge/github.com/snivilised/lorax.svg)](https://pkg.go.dev/github.com/snivilised/lorax)
[![Go report](https://goreportcard.com/badge/github.com/snivilised/lorax)](https://goreportcard.com/report/github.com/snivilised/lorax)
[![Coverage Status](https://coveralls.io/repos/github/snivilised/lorax/badge.svg?branch=master)](https://coveralls.io/github/snivilised/lorax?branch=master&kill_cache=1)
[![lorax Continuous Integration](https://github.com/snivilised/lorax/actions/workflows/ci-workflow.yml/badge.svg)](https://github.com/snivilised/lorax/actions/workflows/ci-workflow.yml)
[![pre-commit](https://img.shields.io/badge/pre--commit-enabled-brightgreen?logo=pre-commit&logoColor=white)](https://github.com/pre-commit/pre-commit)
[![A B](https://img.shields.io/badge/commit-conventional-commits?style=flat)](https://www.conventionalcommits.org/)

<!-- MD013/Line Length -->
<!-- MarkDownLint-disable MD013 -->

<!-- MD014/commands-show-output: Dollar signs used before commands without showing output mark down lint -->
<!-- MarkDownLint-disable MD014 -->

<!-- MD033/no-inline-html: Inline HTML -->
<!-- MarkDownLint-disable MD033 -->

<!-- MD040/fenced-code-language: Fenced code blocks should have a language specified -->
<!-- MarkDownLint-disable MD040 -->

<!-- MD028/no-blanks-blockquote: Blank line inside blockquote -->
<!-- MarkDownLint-disable MD028 -->

<!-- MD010/no-hard-tabs: Hard tabs -->
<!-- MarkDownLint-disable MD010 -->

<p align="left">
  <a href="https://go.dev"><img src="resources/images/go-logo-light-blue.png" width="50" alt="go.dev" /></a>
</p>

## 🔰 Introduction

___Reactive extensions for go (limited set of functionality as required by snivilised)___

This project has been setup for the currency needs of __snivilised__ projects. Ordinarily, I would have used [___RxGo___](https://github.com/ReactiveX/RxGo) for reactive functionality, but it is currently not in a working state and seems to be abandoned. So ___lorax/rx___ will aim to fill the gap, implementing the functionality required by __snivilised__ projects (in particular ___extendio___, which requires the ability to navigate a directory tree concurrently).

The intention is to write a version of RxGo with identical semantics as the original, having fixed some of the outstanding issues in RxGo reported by other users and possibly introduce new features. If and when version 3 of RxGo becomes available, then ___lorax/rx___ will be considered surplice to requirements and retired.

___lorax/boost___ also provides a worker pool implementation using [🐜🐜🐜 ants](https://github.com/panjf2000/ants).

## 🎀 Features

<p align="left">
  <a href="https://onsi.github.io/ginkgo/"><img src="https://onsi.github.io/ginkgo/images/ginkgo.png" width="100" alt="ginkgo"/></a>
  <a href="https://onsi.github.io/gomega/"><img src="https://onsi.github.io/gomega/images/gomega.png" width="100" alt="gomega"/></a>
</p>

+ unit testing with [Ginkgo](https://onsi.github.io/ginkgo/)/[Gomega](https://onsi.github.io/gomega/)
+ implemented with [🐍 Cobra](https://cobra.dev/) cli framework, assisted by [🐲 Cobrass](https://github.com/snivilised/cobrass)
+ uses a heavily modified version of [🐜🐜🐜 ants](https://github.com/panjf2000/ants)
+ i18n with [go-i18n](https://github.com/nicksnyder/go-i18n)
+ linting configuration and pre-commit hooks, (see: [linting-golang](https://freshman.tech/linting-golang/)).

___RxGo___

<p align="left">
  <a href="https://rxjs.dev/guide/overview"><img src="https://avatars.githubusercontent.com/u/6407041?s=200&v=4" width="50" alt="rxjs" /></a>
</p>

To support concurrency features, ___lorax/rx___ uses the reactive model provided by [RxGo](https://github.com/ReactiveX/RxGo). However, since ___RxGo___ seems to be a dead project with its last release in April 2021 and its unit tests not currently fully running successfully, the decision has been made to re-implement this locally. One of the main reasons for the project no longer being actively maintained is the release of generics feature in Go version 1.18, and supporting generics in RxGo would require significant effort to re-write the entire library. While work on this has begun, it's unclear when this will be delivered. Despite this, the reactive model's support for concurrency is highly valued, and ___lorax/rx___ aims to make use of a minimal functionality set for parallel processing during directory traversal. The goal is to replace it with the next version of RxGo when/if it becomes available.

See:

+ [🔆 boost worker pool](./resources/doc/WORKER-POOL.md)
+ [🦑 rx](./resources/doc/RX.md)
+ [🌐 i18n](./resources/doc/i18n-README.md)
+ [🔨 dev notes](./resources/doc/DEVELOPER-INFO.md)
