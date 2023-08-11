# 🌟 lorax: ___Go Concurrency with Functional/Reactive Extensions___

[![A B](https://img.shields.io/badge/branching-commonflow-informational?style=flat)](https://commonflow.org)
[![A B](https://img.shields.io/badge/merge-rebase-informational?style=flat)](https://git-scm.com/book/en/v2/Git-Branching-Rebasing)
[![A B](https://img.shields.io/badge/branch%20history-linear-blue?style=flat)](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/defining-the-mergeability-of-pull-requests/managing-a-branch-protection-rule)
[![Go Reference](https://pkg.go.dev/badge/github.com/snivilised/lorax.svg)](https://pkg.go.dev/github.com/snivilised/lorax)
[![Go report](https://goreportcard.com/badge/github.com/snivilised/lorax)](https://goreportcard.com/report/github.com/snivilised/lorax)
[![Coverage Status](https://coveralls.io/repos/github/snivilised/lorax/badge.svg?branch=master)](https://coveralls.io/github/snivilised/lorax?branch=master&kill_cache=1)
[![Astrolib Continuous Integration](https://github.com/snivilised/lorax/actions/workflows/ci-workflow.yml/badge.svg)](https://github.com/snivilised/lorax/actions/workflows/ci-workflow.yml)
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

<p align="left">
  <a href="https://go.dev"><img src="resources/images/go-logo-light-blue.png" width="50" /></a>
</p>

## 🔰 Introduction

___Reactive extensions for go (limited set of functionality as required by snivilised)___

This project has been setup for the currency needs of __snivilised__ projects. However, that is not to say it is not suitable for other 3rd party projects; it's just that ___lorax___ may not suit their needs.

Ordinarily, I would have used [___RxGo___](https://github.com/ReactiveX/RxGo) for reactive functionality, but it is currently not in a working state and seems to be abandoned. So ___lorax___ will aim to fill the gap, but will only implement the functionality required by __snivilised__ projects (in particular ___extendio___, which requires the ability to navigate a directory tree concurrently).

The intention is to combine the functional implementation in ___lo___ with reactive functionality in ___RxGo___ using generics. Further analysis may reveal this ___lo___ will not be required, but this will be discovered as implementation progresses.

___lorax___ will also provide concurrent functionality via an abstraction that implements the __worker pool__ pattern.

## 📚 Usage

💤 tbd

## 🎀 Features

<p align="left">
  <a href="https://onsi.github.io/ginkgo/"><img src="https://onsi.github.io/ginkgo/images/ginkgo.png" width="100" /></a>
  <a href="https://onsi.github.io/gomega/"><img src="https://onsi.github.io/gomega/images/gomega.png" width="100" /></a>
</p>

+ unit testing with [Ginkgo](https://onsi.github.io/ginkgo/)/[Gomega](https://onsi.github.io/gomega/)
+ implemented with [🐍 Cobra](https://cobra.dev/) cli framework, assisted by [🐲 Cobrass](https://github.com/snivilised/cobrass)
+ i18n with [go-i18n](https://github.com/nicksnyder/go-i18n)
+ linting configuration and pre-commit hooks, (see: [linting-golang](https://freshman.tech/linting-golang/)).

## 🔨 Developer Info

### ☑️ Github changes

Some general project settings are indicated as follows:

___General___

Under `Pull Requests`

+ `Allow merge commits` 🔳 _DISABLE_
+ `Allow squash merging` 🔳 _DISABLE_
+ `Allow rebase merging` ✅ _ENABLE_

___Branch Protection Rules___

Under `Protect matching branches`

+ `Require a pull request before merging` ✅ _ENABLE_
+ `Require linear history` ✅ _ENABLE_
+ `Do not allow bypassing the above settings` ✅ _ENABLE_

### ☑️ Code coverage

+ `coveralls.io`: add maestro project

## 🌐 l10n Translations

This module has been setup to support localisation. The default language is `en-GB` with support for `en-US`. There is a translation file for `en-US` defined as __i18n/deploy/lorax.active.en-US.json__. This is the initial translation for `en-US` that should be deployed with the app.

Make sure that the go-i18n package has been installed so that it can be invoked as cli, see [go-i18n](https://github.com/nicksnyder/go-i18n) for installation instructions.

To maintain localisation of the application, the user must take care to implement all steps to ensure translate-ability of all user facing messages. Whenever there is a need to add/change user facing messages including error messages, to maintain this state, the user must:

+ define template struct (__xxxTemplData__) in __src/i18n/messages.go__ and corresponding __Message()__ method. All messages are defined here in the same location, simplifying the message extraction process as all extractable strings occur at the same place. Please see [go-i18n](https://github.com/nicksnyder/go-i18n) for all translation/pluralisation options and other regional sensitive content.

For more detailed workflow instructions relating to i18n, please see [i18n README](./resources/doc/i18n-README.md)
