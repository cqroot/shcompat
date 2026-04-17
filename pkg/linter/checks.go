package linter

import (
	"fmt"
	"slices"

	"mvdan.cc/sh/v3/syntax"
)

var checkers = map[string]func(*Linter, *syntax.CallExpr) []CheckRule{
	"curl":     (*Linter).CheckCurlCall,
	"realpath": (*Linter).CheckRealpathCall,
	"sed":      (*Linter).CheckSedCall,
}

var (
	ErrSedSandboxNotSupported = fmt.Errorf("sed --sandbox is not supported in older sed versions")
	ErrCurlMissingGloboff     = fmt.Errorf("curl command is missing -g or --globoff argument")
	ErrRealpathNotSupported   = fmt.Errorf("realpath command is not supported in older coreutils versions")
)

type CheckRule struct {
	Id    string
	Error error
}

var CheckRules = map[string]CheckRule{
	"SCPT0300": {Id: "SCPT0300", Error: ErrCurlMissingGloboff},
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
