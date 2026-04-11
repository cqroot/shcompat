package linter

import (
	"fmt"

	"mvdan.cc/sh/v3/syntax"
)

var checkers = map[string]func(*syntax.CallExpr) []error{
	"curl":     CheckCurlCall,
	"realpath": CheckRealpathCall,
	"sed":      CheckSedCall,
}

var (
	ErrSedSandboxNotSupported = fmt.Errorf("sed --sandbox is not supported in older sed versions")
	ErrCurlMissingGloboff     = fmt.Errorf("curl command is missing -g or --globoff argument")
	ErrRealpathNotSupported   = fmt.Errorf("realpath command is not supported in older coreutils versions")
)

func CheckCurlCall(call *syntax.CallExpr) []error {
	var errors []error
	// Check if curl command has -g or --globoff
	for _, arg := range call.Args[1:] {
		if arg.Parts == nil {
			continue
		}
		for _, part := range arg.Parts {
			if lit, ok := part.(*syntax.Lit); ok {
				if lit.Value == "-g" || lit.Value == "--globoff" {
					return errors
				}
			}
		}
	}

	errors = append(errors, ErrCurlMissingGloboff)
	return errors
}

func CheckSedCall(call *syntax.CallExpr) []error {
	var errors []error
	// Check if sed command has --sandbox
	for _, arg := range call.Args[1:] {
		if arg.Parts == nil {
			continue
		}
		for _, part := range arg.Parts {
			if lit, ok := part.(*syntax.Lit); ok {
				if lit.Value == "--sandbox" {
					errors = append(errors, ErrSedSandboxNotSupported)
					return errors
				}
			}
		}
	}

	return errors
}

func CheckRealpathCall(call *syntax.CallExpr) []error {
	var errors []error
	// Check if realpath command is used
	errors = append(errors, ErrRealpathNotSupported)
	return errors
}
