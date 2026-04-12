package linter

import (
	"fmt"

	"mvdan.cc/sh/v3/syntax"
)

var checkers = map[string]func(*syntax.CallExpr) []CheckRule{
	"curl":     CheckCurlCall,
	"realpath": CheckRealpathCall,
	"sed":      CheckSedCall,
}

var (
	ErrSedSandboxNotSupported = fmt.Errorf("sed --sandbox is not supported in older sed versions")
	ErrCurlMissingGloboff     = fmt.Errorf("curl command is missing -g or --globoff argument")
	ErrRealpathNotSupported   = fmt.Errorf("realpath command is not supported in older coreutils versions")
)

type CheckRule struct {
	Id      string
	Enabled bool
	Error   error
}

var CheckRules = map[string]CheckRule{
	"SCPT0300": {Id: "SCPT0300", Enabled: true, Error: ErrCurlMissingGloboff},
	"SCPT1800": {Id: "SCPT1800", Enabled: true, Error: ErrRealpathNotSupported},
	"SCPT1900": {Id: "SCPT1900", Enabled: true, Error: ErrSedSandboxNotSupported},
}

func CheckCurlCall(call *syntax.CallExpr) []CheckRule {
	var rules []CheckRule
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

func CheckRealpathCall(call *syntax.CallExpr) []CheckRule {
	var rules []CheckRule
	// Check if realpath command is used
	rules = append(rules, CheckRules["SCPT1800"])
	return rules
}

func CheckSedCall(call *syntax.CallExpr) []CheckRule {
	var rules []CheckRule
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
