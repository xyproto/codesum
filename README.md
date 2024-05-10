# codesum

Summarize Go, Python, C, C++, Rust, Java, Kotlin, Haskell, JavaScript or TypeScript  projects as a Markdown or JSON document that can be copied and pasted into an LLM.

## Installation

    go install github.com/xyproto/codesum@latest

## Usage

Run `codesum` in the root directory of a project.

If you're on macOS, you might want to run `codesum | pbcopy` to copy all relevant code into the clipboard.

This alias is also a possibility:

```bash
alias c="$HOME/go/bin/codesum | /usr/bin/pbcopy"
```

Use the `-j` or `--json` flag to get a JSON summary instead.

## Note

* This utility is a bit experimental and may need more testing.
* Batching source code into a given token length would be a nice addition. (TODO)

## General info

* Version: 1.0.4
* License: BSD-3
