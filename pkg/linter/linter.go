package linter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"mvdan.cc/sh/v3/syntax"
)

type Linter struct {
	rootDir string
	verbose bool
}

func New(rootDir string, verbose bool) (*Linter, error) {
	var err error
	rootDir, err = filepath.Abs(rootDir)
	if err != nil {
		return nil, err
	}

	l := Linter{
		rootDir: rootDir,
		verbose: verbose,
	}
	return &l, nil
}

type CheckResult struct {
	FilePath string
	Line     int
	Column   int
	Text     string
	Cmd      *syntax.CallExpr
	Error    error
}

func (l *Linter) Run() error {
	var allFailures []CheckResult
	var totalFiles int

	err := filepath.Walk(l.rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if l.verbose {
				fmt.Fprintf(os.Stderr, "walk error at %s: %v\n", path, err)
			}
			return nil
		}
		if info.IsDir() {
			name := info.Name()
			if strings.HasPrefix(name, ".") || name == "node_modules" || name == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}

		if !strings.HasSuffix(path, ".sh") && !strings.HasSuffix(path, ".bash") {
			return nil
		}

		totalFiles++
		if l.verbose {
			fmt.Printf("Check file: %s\n", path)
		}

		failures, err := l.ScanShellScript(path)
		if err != nil {
			if l.verbose {
				fmt.Fprintf(os.Stderr, "skip file %s: %v\n", path, err)
			}
			return nil
		}

		for _, f := range failures {
			fmt.Printf("%s  %s\n", color.BlueString("%s:%d:%d", f.FilePath, f.Line, f.Column), f.Error)
			fmt.Printf("    %s\n", strings.TrimSpace(f.Text))
			fmt.Println()
			allFailures = append(allFailures, f)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to walk directory: %w", err)
	}

	if len(allFailures) == 0 {
		color.HiGreen("%s\n", l.FormatSummary(len(allFailures), totalFiles))
		return nil
	} else {
		color.HiRed("%s\n", l.FormatSummary(len(allFailures), totalFiles))
	}
	return nil
}

func (l *Linter) FormatSummary(issues int, files int) string {
	sb := strings.Builder{}

	if issues == 1 {
		sb.WriteString("Found 1 issue")
	} else {
		fmt.Fprintf(&sb, "Found %d issues", issues)
	}

	if files == 1 {
		sb.WriteString(" in 1 file.")
	} else {
		fmt.Fprintf(&sb, " in %d files.", files)
	}

	return sb.String()
}

// ScanShellScript scans a single shell script file and returns found issues.
func (l *Linter) ScanShellScript(filePath string) ([]CheckResult, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	src := string(content)

	// Parse shell script
	parser := syntax.NewParser(
		syntax.Variant(syntax.LangBash),
		syntax.KeepComments(true),
	)

	file, err := parser.Parse(strings.NewReader(src), filePath)
	if err != nil {
		if l.verbose {
			fmt.Fprintf(os.Stderr, "Parse failed %s: %v\n", filePath, err)
		}
		return nil, err
	}

	var failures []CheckResult
	l.InspectSyntaxNode(file, filePath, src, &failures)

	return failures, nil
}

// Recursively traverse AST to find all errors
func (l *Linter) InspectSyntaxNode(node syntax.Node, filePath string, src string, failures *[]CheckResult) {
	syntax.Walk(node, func(n syntax.Node) bool {
		// Find CallExpr (command call)
		call, ok := n.(*syntax.CallExpr)
		if !ok || len(call.Args) == 0 {
			return true
		}

		cmdWord := call.Args[0]
		if len(cmdWord.Parts) == 0 {
			return true
		}

		lit, ok := cmdWord.Parts[0].(*syntax.Lit)
		if !ok {
			return true
		}

		cmdName := strings.TrimSuffix(lit.Value, ".exe")
		if val, exists := checkers[cmdName]; exists {
			errors := val(call)
			if len(errors) > 0 {
				results := l.constructFailures(errors, filePath, call, src)
				*failures = append(*failures, results...)
			}
		}

		return true
	})
}

func (l *Linter) constructFailures(errors []error, filePath string, call *syntax.CallExpr, src string) []CheckResult {
	results := make([]CheckResult, len(errors))

	relPath, err := filepath.Rel(l.rootDir, filePath)
	if err != nil {
		relPath = filePath
	}
	line := int(call.Pos().Line())
	column := int(call.Pos().Col())

	text := ""
	if call.Pos().IsValid() && call.End().IsValid() {
		start := int(call.Pos().Offset())
		end := int(call.End().Offset())
		if end <= len(src) && start < end {
			text = src[start:end]
		}
	}

	for i, err := range errors {
		res := CheckResult{
			FilePath: relPath,
			Line:     line,
			Column:   column,
			Text:     text,
			Cmd:      call,
			Error:    err,
		}
		results[i] = res
	}
	return results
}
