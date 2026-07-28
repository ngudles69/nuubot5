package fprof

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const importAlias = "nuubotfprof"

// Overlay contains one generated Go build overlay.
type Overlay struct {
	Replace map[string]string `json:"Replace"`
}

// InstrumentResult contains one generated overlay's identity and counts.
type InstrumentResult struct {
	OverlayPath string
	Files       int
	Functions   int
}

// Section 1 - Program Flow

// GenerateOverlay creates instrumented source copies without changing tracked source.
func GenerateOverlay(root, output string) (InstrumentResult, error) {
	// Step 1: create generated source directory
	var generated = filepath.Join(output, "_generated")
	var err = os.MkdirAll(generated, 0o755)
	if err != nil {
		return InstrumentResult{}, fmt.Errorf("create generated source directory: %w", err)
	}

	// Step 2: instrument selected Go source files
	var overlay = Overlay{Replace: make(map[string]string)}
	var result InstrumentResult
	for _, subtree := range []string{"cmd/nuubot-bt-bot", "internal"} {
		var start = filepath.Join(root, filepath.FromSlash(subtree))
		err = filepath.WalkDir(start, func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() {
				if samePath(path, filepath.Join(root, "internal", "fprof")) {
					return filepath.SkipDir
				}
				return nil
			}
			if !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
				return nil
			}
			var relative, relativeErr = filepath.Rel(root, path)
			if relativeErr != nil {
				return relativeErr
			}
			var destination = filepath.Join(generated, relative)
			var functions int
			functions, relativeErr = instrumentFile(root, path, destination)
			if relativeErr != nil {
				return relativeErr
			}
			if functions == 0 {
				return nil
			}
			var sourceAbsolute, sourceErr = filepath.Abs(path)
			if sourceErr != nil {
				return sourceErr
			}
			var destinationAbsolute, destinationErr = filepath.Abs(destination)
			if destinationErr != nil {
				return destinationErr
			}
			overlay.Replace[sourceAbsolute] = destinationAbsolute
			result.Files++
			result.Functions += functions
			return nil
		})
		if err != nil {
			return InstrumentResult{}, fmt.Errorf("instrument %s: %w", subtree, err)
		}
	}

	// Step 3: write Go build overlay
	result.OverlayPath = filepath.Join(output, "overlay.json")
	var file *os.File
	file, err = os.Create(result.OverlayPath)
	if err != nil {
		return InstrumentResult{}, fmt.Errorf("create build overlay: %w", err)
	}
	var encoder = json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	var writeErr = encoder.Encode(overlay)
	var closeErr = file.Close()
	if writeErr != nil {
		return InstrumentResult{}, fmt.Errorf("write build overlay: %w", writeErr)
	}
	if closeErr != nil {
		return InstrumentResult{}, fmt.Errorf("close build overlay: %w", closeErr)
	}
	return result, nil
}

// Section 2 - Domain Helpers

func instrumentFile(root, source, destination string) (int, error) {
	var set = token.NewFileSet()
	var parsed, err = parser.ParseFile(set, source, nil, parser.ParseComments)
	if err != nil {
		return 0, err
	}
	var relative, relativeErr = filepath.Rel(root, filepath.Dir(source))
	if relativeErr != nil {
		return 0, relativeErr
	}
	var packagePath = "nuubot"
	if relative != "." {
		packagePath += "/" + filepath.ToSlash(relative)
	}
	var functions int
	for _, declaration := range parsed.Decls {
		var function, supported = declaration.(*ast.FuncDecl)
		if !supported || function.Body == nil {
			continue
		}
		var name = functionName(packagePath, function)
		var enter = &ast.DeferStmt{Call: &ast.CallExpr{
			Fun: &ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   ast.NewIdent(importAlias),
					Sel: ast.NewIdent("Enter"),
				},
				Args: []ast.Expr{&ast.BasicLit{Kind: token.STRING, Value: strconv.Quote(name)}},
			},
		}}
		var prefix = []ast.Stmt{enter}
		if packagePath == "nuubot/cmd/nuubot-bt-bot" && function.Name.Name == "main" {
			var write = &ast.DeferStmt{Call: &ast.CallExpr{Fun: &ast.SelectorExpr{
				X:   ast.NewIdent(importAlias),
				Sel: ast.NewIdent("Write"),
			}}}
			prefix = []ast.Stmt{write, enter}
		}
		function.Body.List = append(prefix, function.Body.List...)
		functions++
	}
	if functions == 0 {
		return 0, nil
	}
	addImport(parsed)
	ast.SortImports(set, parsed)
	err = os.MkdirAll(filepath.Dir(destination), 0o755)
	if err != nil {
		return 0, err
	}
	var output *os.File
	output, err = os.Create(destination)
	if err != nil {
		return 0, err
	}
	var formatErr = format.Node(output, set, parsed)
	var closeErr = output.Close()
	if formatErr != nil {
		return 0, formatErr
	}
	if closeErr != nil {
		return 0, closeErr
	}
	return functions, nil
}

func addImport(file *ast.File) {
	var spec = &ast.ImportSpec{
		Name: ast.NewIdent(importAlias),
		Path: &ast.BasicLit{Kind: token.STRING, Value: strconv.Quote("nuubot/internal/fprof/runtime")},
	}
	for _, declaration := range file.Decls {
		var imports, supported = declaration.(*ast.GenDecl)
		if supported && imports.Tok == token.IMPORT {
			imports.Specs = append(imports.Specs, spec)
			file.Imports = append(file.Imports, spec)
			return
		}
	}
	var imports = &ast.GenDecl{Tok: token.IMPORT, Specs: []ast.Spec{spec}}
	file.Decls = append([]ast.Decl{imports}, file.Decls...)
	file.Imports = append(file.Imports, spec)
}

func functionName(packagePath string, function *ast.FuncDecl) string {
	if function.Recv == nil || len(function.Recv.List) == 0 {
		return packagePath + "." + function.Name.Name
	}
	return packagePath + "." + receiverName(function.Recv.List[0].Type) + "." + function.Name.Name
}

func receiverName(expression ast.Expr) string {
	switch current := expression.(type) {
	case *ast.Ident:
		return current.Name
	case *ast.StarExpr:
		return "*" + receiverName(current.X)
	case *ast.IndexExpr:
		return receiverName(current.X)
	case *ast.IndexListExpr:
		return receiverName(current.X)
	default:
		return "receiver"
	}
}

// Section 3 - Generic Helpers

func samePath(left, right string) bool {
	var leftAbsolute, leftErr = filepath.Abs(left)
	var rightAbsolute, rightErr = filepath.Abs(right)
	if leftErr != nil || rightErr != nil {
		return false
	}
	return strings.EqualFold(filepath.Clean(leftAbsolute), filepath.Clean(rightAbsolute))
}
