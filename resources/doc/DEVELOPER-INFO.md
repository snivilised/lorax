# 🔨 lorax: Developer Info

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

## ☑️ Github changes

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

## ☑️ Code coverage

+ `coveralls.io`: add maestro project

## 🌐 l10n Translations

This module has been setup to support localisation. The default language is `en-GB` with support for `en-US`. There is a translation file for `en-US` defined as __i18n/deploy/lorax.active.en-US.json__. This is the initial translation for `en-US` that should be deployed with the app.

Make sure that the go-i18n package has been installed so that it can be invoked as cli, see [go-i18n](https://github.com/nicksnyder/go-i18n) for installation instructions.

To maintain localisation of the application, the user must take care to implement all steps to ensure translate-ability of all user facing messages. Whenever there is a need to add/change user facing messages including error messages, to maintain this state, the user must:

+ define template struct (__xxxTemplData__) in __src/i18n/messages.go__ and corresponding __Message()__ method. All messages are defined here in the same location, simplifying the message extraction process as all extractable strings occur at the same place. Please see [go-i18n](https://github.com/nicksnyder/go-i18n) for all translation/pluralisation options and other regional sensitive content.

For more detailed workflow instructions relating to i18n, please see [i18n README](./resources/doc/i18n-README.md)
