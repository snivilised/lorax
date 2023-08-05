# 🌐 i18n Assistance

This document aims to augment that which appears in the [goi18n project](https://github.com/nicksnyder/go-i18n/) and provide information that can help the client use the i18n functionality defined there and its integration into this template project. The translations process is quite laborious, so this project tries alleviate this process by providing helper tasks which will be documented here also.

## 📁 Directory Structure

The local directory structure is as follows:

- ___default___: contains the translation file created by the __newt__ (new translation task). Actually, this task creates an __active__ file (`active.en-GB.json`) in the `i18n/out` folder, the result  of which needs to be manually copied into the __active__ file in the `default` folder.

- ___deploy___: contains all the translations files that are intended to be deployed with the application. There will be one per supported language and by default this template project includes a translation file for __en-US__ (`astrolib.active.en-US.json`)

## ⚙️ Translation Workflow

### ✨ New Translations

__goi18n__ instructs the user to manually create an empty translation message file that they want to add (eg `translate.en-US.json`). This is taken care of by the __newt__ task. Then the requirement is to run the goi18n merge \<active\> command (goi18n merge `astrolib.active.en-US.json` `astrolib.translate.en-US.json`). This has been wrapped up into the __merge__ task and the result is that the translation file `astrolib.translation.en-US.json` is populated with the messages to be translated. So the sequence goes:

- run __newt__ task: (generates default language file `./src/i18n/out/active.en-GB.json` and empty `./src/i18n/out/us-US/astrolib.translation.en-US.json` file). This task can be run from the root folder, __goi18n__ will recursively search the directory tree for files with translate-able content, ie files with template definitions (___i18n.Message___)
- run __merge__ task: derives a translation file for the requested language __en-US__ using 2 files as inputs: source active file (`./src/i18n/out/active.en-GB.json`) and the empty __en-US__ translate file (`./src/i18n/out/us-US/astrolib.translation.en-US.json`), both of which were generated in the previous step.
- hand the translate file to your translator for them to translate
- rename the translate file to the active equivalent (`astrolib.translation.en-US.json`). Save this into the __deploy__ folder. This file will be deployed with the application.

### 🧩 Update Existing Translations (⚠️ not finalised)

__goi18n__ instructs the user to run the following steps:

1. Run `goi18n extract` to update `active.en.toml` with the new messages.
2. Run `goi18n merge active.*.toml` to generate updated `translate.*.toml` files.
3. Translate all the messages in the `translate.*.toml` files.
4. Run `goi18n merge active.*.toml translate.*.toml` to merge the translated messages into the active message files.

___The above description is way too vague and ambiguous. It is for this reason that this process has not been finalised. The intention will be to upgrade the instructions as time goes by and experience is gained.___

However, in this template, the user can execute the following steps:

- run task __update__: this will re-extract messages into `active.en-GB.json` and then runs merge to create an updated translate file.
- hand the translate file to your translator for them to translate
- as before, this translated file should be used to update the active file `astrolib.active.en-US.json` inside the __deploy__ folder.

## 🎓 Task Reference

❗This is a work in progress ...

### 💤 extract

Scans the code base for messages and extracts them into __out/active.en-GB.json__ (The name of the file can't be controlled, it is derived from the language specified). No need to call this task directly.

### 💠 newt

Invokes the `extract` task to extract messages from code. After running this task, the __translate.en-US.json__ file is empty ready to be used as one of the inputs to the merge task (_why do we need an empty input file for the merge? Perhaps the input file instructs the language tag to goi18n_).

Inputs:

- source code

Outputs:

- ./src/i18n/out/active.en-GB.json (messages extracted from code, without hashes)
- ./src/i18n/out/en-US/translate.en-US.json (empty)

### 💠 merge

Inputs:

- ./src/i18n/out/active.en-GB.json
- ./src/i18n/out/en-US/astrolib.translate-en-US.json

Outputs:

- ./src/i18n/out/active.en-US.json
- ./src/i18n/out/translate.en-US.json
