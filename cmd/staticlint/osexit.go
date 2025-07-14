package main

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// OSExitAnalyzer - analyzer for checking os.Exit usage in main function of main package
var OSExitAnalyzer = &analysis.Analyzer{
	Name:     "osexit",
	Doc:      "check for os.Exit usage in main function of main package",
	Run:      runOSExitAnalyzer,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
}

// runOSExitAnalyzer - runs the analyzer for checking os.Exit usage in main function of main package
func runOSExitAnalyzer(pass *analysis.Pass) (interface{}, error) {
	// Check if this is the main package
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	// Filter to examine only function declarations
	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
	}

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		funcDecl := n.(*ast.FuncDecl)

		// Check if this is the main function
		if funcDecl.Name.Name != "main" {
			return
		}

		// Check for receiver (methods are not main functions)
		if funcDecl.Recv != nil {
			return
		}

		// Inspect the main function body for os.Exit calls
		ast.Inspect(funcDecl, func(node ast.Node) bool {
			callExpr, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}

			// Check if this is a selector expression (package.function)
			selExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}

			// Check if the function is "Exit"
			if selExpr.Sel.Name != "Exit" {
				return true
			}

			// Check if the package is "os"
			if ident, ok := selExpr.X.(*ast.Ident); ok {
				// Check the object type
				obj := pass.TypesInfo.ObjectOf(ident)
				if pkg, ok := obj.(*types.PkgName); ok && pkg.Imported().Path() == "os" {
					pass.Reportf(callExpr.Pos(), "direct call to os.Exit is not allowed in main function of main package")
				}
			}

			return true
		})
	})

	return nil, nil
}
