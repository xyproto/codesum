package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"
)

const versionString = "codesum 1.0.3"

var (
	jsonOutput  bool
	versionFlag bool
)

func init() {
	flag.BoolVar(&jsonOutput, "j", false, "Output in JSON format")
	flag.BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	flag.BoolVar(&versionFlag, "v", false, "Prints the version of the program")
	flag.BoolVar(&versionFlag, "version", false, "Prints the version of the program")
	flag.Parse()

	if versionFlag {
		fmt.Println(versionString)
		os.Exit(0)
	}
}

type FileInfo struct {
	Path         string `json:"path"`
	Language     string `json:"language"`
	LineCount    int    `json:"line_count,omitempty"`
	LastModified string `json:"last_modified,omitempty"`
	Contents     string `json:"contents,omitempty"`
}

type ProjectInfo struct {
	Name       string     `json:"name"`
	Repository string     `json:"repository"`
	Files      []FileInfo `json:"files"`
	Type       string     `json:"type"`
}

func recognizedExtension(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".go", ".cpp", ".hpp", ".cc", ".h", ".rs", ".c", ".py":
		return true
	}
	return false
}

func languageFromExtension(ext string) string {
	switch ext {
	case ".go":
		return "Go"
	case ".cpp", ".cc":
		return "C++"
	case ".hpp", ".h":
		return "C/C++ Header"
	case ".rs":
		return "Rust"
	case ".c":
		return "C"
	case ".py":
		return "Python"
	default:
		return "Unknown"
	}
}

func loadIgnorePatterns(filenames ...string) (map[string]struct{}, error) {
	ignores := make(map[string]struct{})
	for _, filename := range filenames {
		data, err := os.ReadFile(filename)
		if err != nil {
			continue // Ignore files that cannot be read or don't exist
		}
		scanner := bufio.NewScanner(strings.NewReader(string(data)))
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" && !strings.HasPrefix(line, "#") {
				ignores[line] = struct{}{}
			}
		}
	}
	// Add common ignores
	commonIgnores := []string{"vendor", "test", "tmp", "backup", "node_modules"}
	for _, dir := range commonIgnores {
		ignores[dir] = struct{}{}
	}
	return ignores, nil
}

func shouldSkip(path string, ignores map[string]struct{}) bool {
	for ignore := range ignores {
		if matched, _ := filepath.Match(ignore, filepath.Base(path)); matched {
			return true
		}
		if strings.HasPrefix(path, ignore+"/") {
			return true
		}
	}
	return false
}

func detectProjectType(files []FileInfo) string {
	languageCount := make(map[string]int)
	for _, file := range files {
		languageCount[file.Language]++
	}

	maxCount := 0
	projectType := "Unknown"
	for lang, count := range languageCount {
		if count > maxCount {
			maxCount = count
			projectType = lang
		}
	}
	return projectType
}

func countLines(path string) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	lineCount := 0
	for scanner.Scan() {
		lineCount++
	}
	return lineCount, scanner.Err()
}

func readProjectName(modFilePath string) (string, error) {
	file, err := os.Open(modFilePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if strings.HasPrefix(scanner.Text(), "module ") {
			parts := strings.Fields(scanner.Text())
			if len(parts) > 1 {
				return parts[1], nil // Return the module name
			}
		}
	}
	return "", fmt.Errorf("no module declaration found in %s", modFilePath)
}

func readGitConfig(configFilePath string) (string, error) {
	file, err := os.Open(configFilePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	inRemoteSection := false
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "[remote \"origin\"]") {
			inRemoteSection = true
		} else if inRemoteSection && strings.Contains(line, "url =") {
			return strings.TrimSpace(strings.Split(line, "=")[1]), nil
		} else if inRemoteSection && line == "" {
			break // Exit if we reach the end of the section
		}
	}
	return "", fmt.Errorf("no URL found in %s", configFilePath)
}

func walkDirectoryAndCollectFiles(ignores map[string]struct{}) ([]FileInfo, error) {
	var files []FileInfo
	g := new(errgroup.Group)
	var mu sync.Mutex // To protect concurrent writes to the files slice

	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && shouldSkip(path, ignores) {
			return fs.SkipDir
		}
		if !d.IsDir() && recognizedExtension(path) {
			ext := filepath.Ext(path)
			language := languageFromExtension(ext)
			if language != "Unknown" {
				g.Go(func() error {
					fileInfo, err := os.Stat(path)
					if err != nil {
						return err
					}
					content, err := os.ReadFile(path)
					if err != nil {
						return err
					}
					lineCount, _ := countLines(string(content))
					mu.Lock()
					files = append(files, FileInfo{
						Path:         path,
						Language:     language,
						LineCount:    lineCount,
						LastModified: fileInfo.ModTime().Format("2006-01-02 15:04:05"),
						Contents:     string(content),
					})
					mu.Unlock()
					return nil
				})
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return files, nil
}

func outputProjectInfo(project ProjectInfo) {
	if jsonOutput {
		data, err := json.MarshalIndent(project, "", "  ")
		if err != nil {
			fmt.Println("Error marshaling JSON:", err)
			return
		}
		fmt.Println(string(data))
	} else {
		// Start Markdown output
		fmt.Printf("# %s\n\n", project.Name)
		fmt.Printf("* Main language: %s\n", project.Type)
		fmt.Printf("* Package name: %s\n\n", project.Repository)

		fmt.Print("## Source code\n\n")
		for _, file := range project.Files {
			fmt.Printf("### %s\n\n", file.Path)
			fmt.Printf("```%s\n", file.Language)
			fmt.Println(file.Contents)
			fmt.Println("```\n")
		}
	}
}

func main() {
	ignores, err := loadIgnorePatterns(".ignore", ".gitignore")
	if err != nil {
		fmt.Printf("Error loading ignore patterns: %v\n", err)
		return
	}

	files, err := walkDirectoryAndCollectFiles(ignores)
	if err != nil {
		fmt.Printf("Error walking directory and collecting files: %v\n", err)
		return
	}

	// Fetch project name from go.mod, if available
	projectName, err := readProjectName("go.mod")
	if err != nil {
		fmt.Printf("Error reading project name from go.mod: %v\n", err)
		projectName = filepath.Base(filepath.Dir("."))
	}

	// Fetch repository name from .git/config, if available
	repoName, err := readGitConfig(".git/config")
	if err != nil {
		fmt.Printf("Error reading repository name from .git/config: %v\n", err)
		repoName = "Unknown"
	}

	projectType := detectProjectType(files)

	project := ProjectInfo{
		Name:       projectName,
		Repository: repoName,
		Files:      files,
		Type:       projectType,
	}

	outputProjectInfo(project)
}
