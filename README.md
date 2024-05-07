# codesum

Summarize Go, Python, C, C++ or Rust project as a Markdown document that can be copied and pasted into an LLM.

## Installation

    go install github.com/xyproto/codesum@latest

## Usage

Run `codesum` in the root directory of a project.

If you're on macOS, you might want to run `codesum | pbcopy` to copy all relevant code into the clipboard.

## Note

This utility is a bit experimental and needs more testing.

## Example output

````
# github.com/xyproto/codesum

### main.go

```go
package main

import (
    "bufio"
    "fmt"
    "os"
    "path/filepath"
    "strings"
)

func recognizedExtension(path string) bool {
    switch strings.ToLower(filepath.Ext(path)) {
    case ".go", ".cpp", ".hpp", ".cc", ".h", ".rs", ".c", ".py":
        return true
    }
    return false
}

func getProjectInfo() (string, string, error) {
    currentDir, err := os.Getwd()
    if err != nil {
        return "", "", err
    }
    projectName := filepath.Base(currentDir)

    if name, err := readProjectName("go.mod"); err == nil {
        return name, "go", nil
    }
    if name, err := readGitConfig(".git/config"); err == nil {
        return name, detectProjectType(), nil
    }
    if _, err := os.Stat("CMakeLists.txt"); err == nil {
        return "C++ Project", "cpp", nil
    }
    if _, err := os.Stat("Cargo.toml"); err == nil {
        return "Rust Project", "rust", nil
    }
    if _, err := os.Stat("setup.py"); err == nil {
        return "Python Project", "python", nil
    }
    if _, err := os.Stat("Makefile"); err == nil {
        return "C Project", "c", nil
    }
    return projectName, detectProjectType(), nil
}

func readProjectName(modFilePath string) (string, error) {
    file, err := os.Open(modFilePath)
    if err != nil {
        return "", err
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        if line := scanner.Text(); strings.HasPrefix(line, "module ") {
            parts := strings.Fields(line)
            if len(parts) >= 2 {
                return parts[1], nil
            }
            break
        }
    }
    return "", nil
}

func readGitConfig(configFilePath string) (string, error) {
    file, err := os.Open(configFilePath)
    if err != nil {
        return "", err
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        if line := scanner.Text(); strings.Contains(line, "url =") {
            parts := strings.Split(line, "=")
            if len(parts) >= 2 {
                url := strings.TrimSpace(parts[1])
                urlParts := strings.Split(url, "/")
                repoName := strings.TrimSuffix(urlParts[len(urlParts)-1], ".git")
                return repoName, nil
            }
        }
    }
    return "", nil
}

func detectProjectType() string {
    cppFiles := false // Track presence of C++ files
    cFiles := false   // Track presence of C files

    files, err := os.ReadDir(".")
    if err != nil {
        return "unknown"
    }

    for _, file := range files {
        switch ext := strings.ToLower(filepath.Ext(file.Name())); ext {
        case ".go":
            return "go"
        case ".cpp", ".hpp", ".cc":
            cppFiles = true
        case ".c":
            cFiles = true
        case ".h":
            // Note that .h could be either C or C++, will determine after the loop
            continue
        case ".rs":
            return "rust"
        case ".py":
            return "python"
        }
    }

    if cppFiles {
        return "cpp" // Prioritize C++ if any .cpp, .hpp, or .cc files found
    }
    if cFiles || (cppFiles && filesContainH(files)) {
        return "c" // Return C only if .c files found or both C++ and .h files without .cpp/.hpp/.cc
    }

    return "unknown"
}

// Additional helper function to check if directory contains any .h files
func filesContainH(files []os.DirEntry) bool {
    for _, file := range files {
        if strings.ToLower(filepath.Ext(file.Name())) == ".h" {
            return true
        }
    }
    return false
}

func main() {
    projectName, projectType, err := getProjectInfo()
    if err != nil {
        fmt.Printf("Failed to determine project type or name: %v\n", err)
        return
    }

    var sb strings.Builder
    sb.WriteString("# " + projectName + "\n\n")

    err = filepath.WalkDir(".", func(path string, d os.DirEntry, err error) error {
        if err != nil {
            return err
        }
        if d.IsDir() || !recognizedExtension(path) {
            return nil
        }
        content, err := os.ReadFile(path)
        if err != nil {
            return err
        }
        sb.WriteString(fmt.Sprintf("### %s\n\n```%s\n%s\n```\n\n", path, projectType, strings.TrimSpace(string(content))))
        return nil
    })

    if err != nil {
        fmt.Println("Error walking through files:", err)
    } else {
        fmt.Println(sb.String())
    }
}
```
````

## General info

* Version: 1.0.0
* License: BSD-3
