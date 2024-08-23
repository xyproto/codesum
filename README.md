# codesum

Summarize Go, Python, C, C++, Rust, Java, Kotlin, Haskell, JavaScript or TypeScript  projects as a Markdown or JSON document.

This makes it quick and easy to ie. copy several source files to the clipboard and then paste them into an AI / LLM frontend.

## Installation

    go install github.com/xyproto/codesum@latest

## Usage

Run `codesum` in the root directory of a project.

### Linux

    codesum -j | xclip -selection clipboard

### macOS

    codesum -j | pbcopy

## General info

* Version: 1.1.0
* License: BSD-3
