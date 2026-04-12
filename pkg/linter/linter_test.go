package linter_test

import (
	"testing"

	"github.com/cqroot/shcompat/pkg/linter"
	"github.com/stretchr/testify/require"
)

type TestCase struct {
	name     string
	src      string
	expected []linter.CheckResult
}

var testCases = []TestCase{
	{
		name:     "simple command",
		src:      `ls -l`,
		expected: []linter.CheckResult{},
	},
	{
		name: "curl command without -g",
		src:  "curl http://example.com",
		expected: []linter.CheckResult{
			{
				Line:   1,
				Column: 1,
				Text:   "curl http://example.com",
				Rule:   linter.CheckRules["SCPT0300"],
			},
		},
	},
	{
		name:     "curl command with -g",
		src:      "curl -g http://example.com",
		expected: []linter.CheckResult{},
	},
	{
		name: "sed command with --sandbox",
		src:  `sed --sandbox 's/foo/bar/' file.txt`,
		expected: []linter.CheckResult{
			{
				Line:   1,
				Column: 1,
				Text:   "sed --sandbox 's/foo/bar/' file.txt",
				Rule:   linter.CheckRules["SCPT1900"],
			},
		},
	},
	{
		name: "sed command with --sandbox and extra spaces",
		src:  `  sed --sandbox 's/foo/bar/' file.txt  `,
		expected: []linter.CheckResult{
			{
				Line:   1,
				Column: 3,
				Text:   "sed --sandbox 's/foo/bar/' file.txt",
				Rule:   linter.CheckRules["SCPT1900"],
			},
		},
	},
	{
		name: "sed command with --sandbox and pipeline",
		src:  `sed --sandbox 's/foo/bar/' file.txt | grep foo`,
		expected: []linter.CheckResult{
			{
				Line:   1,
				Column: 1,
				Text:   "sed --sandbox 's/foo/bar/' file.txt",
				Rule:   linter.CheckRules["SCPT1900"],
			},
		},
	},
	{
		name:     "sed command without --sandbox",
		src:      `sed 's/foo/bar/' file.txt`,
		expected: []linter.CheckResult{},
	},
	{
		name: "realpath command",
		src:  "realpath /path/to/file",
		expected: []linter.CheckResult{
			{
				Line:   1,
				Column: 1,
				Text:   "realpath /path/to/file",
				Rule:   linter.CheckRules["SCPT1800"],
			},
		},
	},
}

func TestLinter(t *testing.T) {
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			l, err := linter.New("", false)
			if err != nil {
				t.Fatalf("failed to create linter: %v", err)
			}

			results, err := l.CheckShellContent(tt.src, "")
			if err != nil {
				t.Fatalf("CheckShellContent error: %v", err)
			}

			require.Equal(t, len(tt.expected), len(results))
			if len(results) != len(tt.expected) {
				t.Fatalf("expected %d results, got %d", len(tt.expected), len(results))
			}

			for i, res := range results {
				require.Equal(t, tt.expected[i].Line, res.Line)
				require.Equal(t, tt.expected[i].Column, res.Column)
				require.Equal(t, tt.expected[i].Text, res.Text)
				require.Equal(t, tt.expected[i].Rule, res.Rule)
			}
		})
	}
}
