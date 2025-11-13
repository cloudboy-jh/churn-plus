package engine

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Scanner scans a project directory and returns structured file information
type Scanner struct {
	rootPath       string
	ignorePatterns []string
}

// NewScanner creates a new project scanner
func NewScanner(rootPath string, ignorePatterns []string) *Scanner {
	return &Scanner{
		rootPath:       rootPath,
		ignorePatterns: ignorePatterns,
	}
}

// Scan traverses the project and returns all relevant files
func (s *Scanner) Scan() ([]*FileInfo, error) {
	var files []*FileInfo

	err := filepath.Walk(s.rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			// Check if directory should be ignored
			if s.shouldIgnore(path) {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip ignored files
		if s.shouldIgnore(path) {
			return nil
		}

		// Only include code files
		if !s.isCodeFile(path) {
			return nil
		}

		fileInfo, err := s.getFileInfo(path)
		if err != nil {
			// Skip files we can't read
			return nil
		}

		files = append(files, fileInfo)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scan directory: %w", err)
	}

	return files, nil
}

// getFileInfo extracts metadata about a file
func (s *Scanner) getFileInfo(path string) (*FileInfo, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	lines, err := s.countLines(path)
	if err != nil {
		lines = 0 // If we can't count lines, default to 0
	}

	language := detectLanguage(path)

	return &FileInfo{
		Path:     path,
		Language: language,
		Size:     stat.Size(),
		Lines:    lines,
	}, nil
}

// countLines counts the number of lines in a file
func (s *Scanner) countLines(path string) (int, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	count := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		count++
	}

	return count, scanner.Err()
}

// shouldIgnore checks if a path matches any ignore patterns
func (s *Scanner) shouldIgnore(path string) bool {
	relPath, err := filepath.Rel(s.rootPath, path)
	if err != nil {
		relPath = path
	}

	for _, pattern := range s.ignorePatterns {
		// Simple pattern matching (can be enhanced with glob later)
		if strings.Contains(relPath, pattern) {
			return true
		}
		// Check if basename matches
		if strings.Contains(filepath.Base(path), pattern) {
			return true
		}
	}

	return false
}

// isCodeFile determines if a file is a code file worth analyzing
func (s *Scanner) isCodeFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))

	codeExtensions := map[string]bool{
		// JavaScript/TypeScript
		".js": true, ".jsx": true, ".ts": true, ".tsx": true, ".mjs": true, ".cjs": true,
		// Python
		".py": true, ".pyw": true, ".pyx": true,
		// Go
		".go": true,
		// Rust
		".rs": true,
		// C/C++
		".c": true, ".cpp": true, ".cc": true, ".cxx": true, ".h": true, ".hpp": true,
		// Java/Kotlin
		".java": true, ".kt": true, ".kts": true,
		// C#
		".cs": true,
		// Ruby
		".rb": true,
		// PHP
		".php": true,
		// Swift
		".swift": true,
		// Shell
		".sh": true, ".bash": true, ".zsh": true,
		// Web
		".html": true, ".css": true, ".scss": true, ".sass": true, ".less": true,
		".vue": true, ".svelte": true,
		// Config (selective)
		".json": true, ".yaml": true, ".yml": true, ".toml": true,
		// Other
		".sql": true, ".graphql": true, ".proto": true,
	}

	return codeExtensions[ext]
}

// detectLanguage determines the programming language from file extension
func detectLanguage(path string) string {
	ext := strings.ToLower(filepath.Ext(path))

	languageMap := map[string]string{
		".js": "javascript", ".jsx": "javascript", ".mjs": "javascript", ".cjs": "javascript",
		".ts": "typescript", ".tsx": "typescript",
		".py": "python", ".pyw": "python", ".pyx": "python",
		".go": "go",
		".rs": "rust",
		".c":  "c", ".h": "c",
		".cpp": "cpp", ".cc": "cpp", ".cxx": "cpp", ".hpp": "cpp",
		".java": "java",
		".kt":   "kotlin", ".kts": "kotlin",
		".cs":    "csharp",
		".rb":    "ruby",
		".php":   "php",
		".swift": "swift",
		".sh":    "bash", ".bash": "bash", ".zsh": "zsh",
		".html": "html",
		".css":  "css", ".scss": "scss", ".sass": "sass", ".less": "less",
		".vue":    "vue",
		".svelte": "svelte",
		".json":   "json",
		".yaml":   "yaml", ".yml": "yaml",
		".toml":    "toml",
		".sql":     "sql",
		".graphql": "graphql",
		".proto":   "protobuf",
	}

	if lang, ok := languageMap[ext]; ok {
		return lang
	}

	return "unknown"
}

// BuildFileTree creates a hierarchical tree structure from files
func BuildFileTree(files []*FileInfo, rootPath string) *FileNode {
	root := &FileNode{
		Name:     filepath.Base(rootPath),
		Path:     rootPath,
		IsDir:    true,
		Children: make([]*FileNode, 0),
	}

	for _, file := range files {
		relPath, err := filepath.Rel(rootPath, file.Path)
		if err != nil {
			continue
		}

		parts := strings.Split(filepath.ToSlash(relPath), "/")
		current := root

		// Create directory nodes
		for i := 0; i < len(parts)-1; i++ {
			dirName := parts[i]
			dirNode := current.findChild(dirName)
			if dirNode == nil {
				dirNode = &FileNode{
					Name:     dirName,
					Path:     filepath.Join(current.Path, dirName),
					IsDir:    true,
					Children: make([]*FileNode, 0),
				}
				current.Children = append(current.Children, dirNode)
			}
			current = dirNode
		}

		// Add file node
		fileName := parts[len(parts)-1]
		fileNode := &FileNode{
			Name:     fileName,
			Path:     file.Path,
			IsDir:    false,
			FileInfo: file,
		}
		current.Children = append(current.Children, fileNode)
	}

	return root
}

// FileNode represents a node in the file tree
type FileNode struct {
	Name     string
	Path     string
	IsDir    bool
	Children []*FileNode
	FileInfo *FileInfo
	Expanded bool // For UI state
}

// findChild finds a child node by name
func (n *FileNode) findChild(name string) *FileNode {
	for _, child := range n.Children {
		if child.Name == name {
			return child
		}
	}
	return nil
}

// GetAllFiles returns a flattened list of all files in the tree
func (n *FileNode) GetAllFiles() []*FileInfo {
	var files []*FileInfo

	if !n.IsDir && n.FileInfo != nil {
		files = append(files, n.FileInfo)
	}

	for _, child := range n.Children {
		files = append(files, child.GetAllFiles()...)
	}

	return files
}
