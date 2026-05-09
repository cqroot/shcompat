package linter

import (
	"fmt"
	"slices"

	"mvdan.cc/sh/v3/syntax"
)

type Checker struct {
	CheckFunc func(*Linter, *syntax.CallExpr) []CheckRule
}

var checkers = map[string]Checker{
	"curl":     {CheckFunc: (*Linter).CheckCurlCall},
	"grep":     {CheckFunc: (*Linter).CheckGrepCall},
	"realpath": {CheckFunc: (*Linter).CheckRealpathCall},
	"sed":      {CheckFunc: (*Linter).CheckSedCall},
}

var (
	ErrCurlMissingGloboff     = fmt.Errorf("curl command is missing -g or --globoff argument")
	ErrGrepStrayBackslash     = fmt.Errorf("grep command has a pattern with a stray backslash, which may cause issues in some versions of grep")
	ErrRealpathNotSupported   = fmt.Errorf("realpath command is not supported in older coreutils versions")
	ErrSedSandboxNotSupported = fmt.Errorf("sed --sandbox is not supported in older sed versions")
)

type CheckRule struct {
	Id    string
	Error error
}

var CheckRules = map[string]CheckRule{
	"SCPT0300": {Id: "SCPT0300", Error: ErrCurlMissingGloboff},
	"SCPT0700": {Id: "SCPT0700", Error: ErrGrepStrayBackslash},
	"SCPT1800": {Id: "SCPT1800", Error: ErrRealpathNotSupported},
	"SCPT1900": {Id: "SCPT1900", Error: ErrSedSandboxNotSupported},
}

func (l *Linter) CheckCurlCall(call *syntax.CallExpr) []CheckRule {
	var rules []CheckRule
	if len(l.includeRules) > 0 && !slices.Contains(l.includeRules, "SCPT0300") {
		return rules
	}
	if len(l.excludeRules) > 0 && slices.Contains(l.excludeRules, "SCPT0300") {
		return rules
	}

	// Check if curl command has -g or --globoff
	for _, arg := range call.Args[1:] {
		if arg.Parts == nil {
			continue
		}
		for _, part := range arg.Parts {
			if lit, ok := part.(*syntax.Lit); ok {
				if lit.Value == "-g" || lit.Value == "--globoff" {
					return rules
				}
			}
		}
	}

	rules = append(rules, CheckRules["SCPT0300"])
	return rules
}

func containsStrayBackslash(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' {
			if i+1 < len(s) && s[i+1] == '\\' {
				i++ // skip the escaped backslash
			} else {
				return true // stray backslash
			}
		}
	}
	return false
}

func containsStrayBackslashInWordPart(part syntax.WordPart) bool {
	switch p := part.(type) {
	case *syntax.Lit:
		return containsStrayBackslash(p.Value)
	case *syntax.SglQuoted:
		return containsStrayBackslash(p.Value)
	case *syntax.DblQuoted:
		for _, nested := range p.Parts {
			if containsStrayBackslashInWordPart(nested) {
				return true
			}
		}
	}
	return false
}

func (l *Linter) CheckGrepCall(call *syntax.CallExpr) []CheckRule {
	var rules []CheckRule
	if len(l.includeRules) > 0 && !slices.Contains(l.includeRules, "SCPT0700") {
		return rules
	}
	if len(l.excludeRules) > 0 && slices.Contains(l.excludeRules, "SCPT0700") {
		return rules
	}

	// Check if grep command has a pattern with a stray backslash
	for _, arg := range call.Args {
		if arg.Parts == nil {
			continue
		}
		for _, part := range arg.Parts {
			if containsStrayBackslashInWordPart(part) {
				rules = append(rules, CheckRules["SCPT0700"])
				return rules
			}
		}
	}

	return rules
}

func (l *Linter) CheckRealpathCall(call *syntax.CallExpr) []CheckRule {
	var rules []CheckRule
	if len(l.includeRules) > 0 && !slices.Contains(l.includeRules, "SCPT1800") {
		return rules
	}
	if len(l.excludeRules) > 0 && slices.Contains(l.excludeRules, "SCPT1800") {
		return rules
	}

	// Check if realpath command is used
	rules = append(rules, CheckRules["SCPT1800"])
	return rules
}

func (l *Linter) CheckSedCall(call *syntax.CallExpr) []CheckRule {
	var rules []CheckRule
	if len(l.includeRules) > 0 && !slices.Contains(l.includeRules, "SCPT1900") {
		return rules
	}
	if len(l.excludeRules) > 0 && slices.Contains(l.excludeRules, "SCPT1900") {
		return rules
	}

	// Check if sed command has --sandbox
	for _, arg := range call.Args[1:] {
		if arg.Parts == nil {
			continue
		}
		for _, part := range arg.Parts {
			if lit, ok := part.(*syntax.Lit); ok {
				if lit.Value == "--sandbox" {
					rules = append(rules, CheckRules["SCPT1900"])
					return rules
				}
			}
		}
	}

	return rules
}
