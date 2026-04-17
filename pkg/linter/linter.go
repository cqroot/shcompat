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
	target  string
	verbose bool
	format  Format
}

type Option func(*Linter)

func New(target string, opts ...Option) (*Linter, error) {
	l := Linter{
		verbose: false,
		format:  FormatTTY,
	}

	for _, opt := range opts {
		opt(&l)
	}

	var err error
	l.target, err = filepath.Abs(target)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(l.target)
	if err != nil {
		return nil, fmt.Errorf("failed to stat target path: %w", err)
	}

	if info.IsDir() {
		l.rootDir = l.target
	} else {
		l.rootDir = filepath.Dir(l.target)
	}

	return &l, nil
}

type CheckResult struct {
	FilePath string           `json:"file"`
	Line     int              `json:"line"`
	Column   int              `json:"column"`
	Text     string           `json:"text"`
	Cmd      *syntax.CallExpr `json:"cmd"`
	Rule     CheckRule        `json:"rule"`
}

func (l *Linter) Run() error {
	var allFailures []CheckResult
	var totalFiles int

	err := filepath.Walk(l.target, func(path string, info os.FileInfo, err error) error {
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

		failures, err := l.CheckShellFile(path)
		if err != nil {
			if l.verbose {
				fmt.Fprintf(os.Stderr, "skip file %s: %v\n", path, err)
			}
			return nil
		}

		switch l.format {
		case FormatTTY:
			fmt.Print(ToTTY(failures))
		case FormatJSON:
			jsonStr, err := ToJSON(failures)
			if err != nil {
				return err
			}
			fmt.Println(jsonStr)
		}
		allFailures = append(allFailures, failures...)

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to walk directory: %w", err)
	}

	if l.format != FormatTTY {
		return nil
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

// CheckShellFile scans a single shell script file and returns found issues.
func (l *Linter) CheckShellFile(filePath string) ([]CheckResult, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	src := string(content)

	// Parse shell script
	failures, err := l.CheckShellContent(src, filePath)
	if err != nil && l.verbose {
		fmt.Fprintf(os.Stderr, "failed to parse file %s: %v\n", filePath, err)
	}
	return failures, err
}

// CheckShellContent parses shell script content and returns found issues.
func (l *Linter) CheckShellContent(src string, filePath string) ([]CheckResult, error) {
	parser := syntax.NewParser(
		syntax.Variant(syntax.LangBash),
		syntax.KeepComments(true),
	)

	file, err := parser.Parse(strings.NewReader(src), filePath)
	if err != nil {
		if l.verbose {
			fmt.Fprintf(os.Stderr, "failed to parse content in %s: %v\n", filePath, err)
		}
		return nil, err
	}

	var failures []CheckResult
	l.CheckSyntaxNode(file, filePath, src, &failures)

	return failures, nil
}

// CheckSyntaxNode recursively traverse AST to find all errors
func (l *Linter) CheckSyntaxNode(node syntax.Node, filePath string, src string, failures *[]CheckResult) {
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
			rules := val(call)
			if len(rules) > 0 {
				results := l.constructFailures(rules, filePath, call, src)
				*failures = append(*failures, results...)
			}
		}

		return true
	})
}

func (l *Linter) constructFailures(rules []CheckRule, filePath string, call *syntax.CallExpr, src string) []CheckResult {
	results := make([]CheckResult, len(rules))

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

	for i, rule := range rules {
		res := CheckResult{
			FilePath: relPath,
			Line:     line,
			Column:   column,
			Text:     text,
			Cmd:      call,
			Rule:     rule,
		}
		results[i] = res
	}
	return results
}
