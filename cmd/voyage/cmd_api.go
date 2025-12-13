// Package main provides API documentation lookup for engine packages.
package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"sort"
	"strings"
)

// PackageAPI holds extracted API information for a package.
type PackageAPI struct {
	Name      string
	Path      string
	Types     []TypeAPI
	Functions []FuncAPI
}

// TypeAPI holds information about an exported type.
type TypeAPI struct {
	Name       string
	Doc        string
	Kind       string // "struct", "interface", "alias"
	Methods    []FuncAPI
	Fields     []FieldAPI
	Underlying string // for type aliases
}

// FuncAPI holds information about an exported function or method.
type FuncAPI struct {
	Name       string
	Doc        string
	Signature  string
	IsMethod   bool
	ReceiverTy string
}

// FieldAPI holds information about a struct field.
type FieldAPI struct {
	Name string
	Type string
	Doc  string
	Tag  string
}

// enginePackages lists all engine packages to document.
var enginePackages = map[string]string{
	"tetra":   "engine/tetra",
	"lod":     "engine/lod",
	"shader":  "engine/shader",
	"assets":  "engine/assets",
	"render":  "engine/render",
	"camera":  "engine/camera",
	"display": "engine/display",
	"input":   "engine/input",
	"view":    "engine/view",
	"effects": "engine/effects",
}

func runAPICommand(args []string) {
	fs := flag.NewFlagSet("api", flag.ExitOnError)
	searchQuery := fs.String("search", "", "Search for APIs matching query")
	showMethods := fs.Bool("methods", false, "Show method signatures")
	fs.Usage = func() {
		fmt.Println(`Usage: voyage api [options] [package[.Type]]

List and explore engine API documentation.

Examples:
  voyage api                    # List all packages
  voyage api tetra              # List types in tetra package
  voyage api tetra.Scene        # Show Scene type details
  voyage api tetra.Scene -m     # Show Scene methods
  voyage api --search camera    # Search all packages for "camera"

Options:`)
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	// Handle search mode
	if *searchQuery != "" {
		searchAPIs(*searchQuery)
		return
	}

	remaining := fs.Args()

	// No args: list all packages
	if len(remaining) == 0 {
		listPackages()
		return
	}

	target := remaining[0]

	// Check if it's package.Type format
	if strings.Contains(target, ".") {
		parts := strings.SplitN(target, ".", 2)
		pkgName := parts[0]
		typeName := parts[1]
		showTypeDetails(pkgName, typeName, *showMethods)
		return
	}

	// It's just a package name
	showPackage(target, *showMethods)
}

func listPackages() {
	fmt.Println("Engine Packages:")
	fmt.Println()

	// Sort packages for consistent output
	names := make([]string, 0, len(enginePackages))
	for name := range enginePackages {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		path := enginePackages[name]
		fmt.Printf("  %-12s %s\n", name, path)
	}

	fmt.Println()
	fmt.Println("Use 'voyage api <package>' to see types in a package")
	fmt.Println("Use 'voyage api <package>.<Type>' to see type details")
}

func showPackage(pkgName string, showMethods bool) {
	path, ok := enginePackages[pkgName]
	if !ok {
		fmt.Fprintf(os.Stderr, "Unknown package: %s\n", pkgName)
		fmt.Fprintf(os.Stderr, "Available packages: %v\n", getPackageNames())
		os.Exit(1)
	}

	api, err := extractPackageAPI(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing package: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("%s - %s\n", pkgName, path)
	fmt.Println(strings.Repeat("=", len(pkgName)+len(path)+3))
	fmt.Println()

	// Show types
	if len(api.Types) > 0 {
		fmt.Println("Types:")
		for _, t := range api.Types {
			fmt.Printf("  %s", t.Name)
			if t.Kind != "" && t.Kind != "struct" {
				fmt.Printf(" (%s)", t.Kind)
			}
			if t.Doc != "" {
				fmt.Printf(" - %s", firstLine(t.Doc))
			}
			fmt.Println()

			if showMethods && len(t.Methods) > 0 {
				for _, m := range t.Methods {
					fmt.Printf("    %s\n", m.Signature)
				}
			}
		}
		fmt.Println()
	}

	// Show standalone functions
	standaloneFuncs := []FuncAPI{}
	for _, f := range api.Functions {
		if !f.IsMethod {
			standaloneFuncs = append(standaloneFuncs, f)
		}
	}

	if len(standaloneFuncs) > 0 {
		fmt.Println("Functions:")
		for _, f := range standaloneFuncs {
			fmt.Printf("  %s\n", f.Signature)
		}
		fmt.Println()
	}
}

func showTypeDetails(pkgName, typeName string, showMethods bool) {
	path, ok := enginePackages[pkgName]
	if !ok {
		fmt.Fprintf(os.Stderr, "Unknown package: %s\n", pkgName)
		os.Exit(1)
	}

	api, err := extractPackageAPI(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing package: %v\n", err)
		os.Exit(1)
	}

	// Find the type
	var foundType *TypeAPI
	for i := range api.Types {
		if api.Types[i].Name == typeName {
			foundType = &api.Types[i]
			break
		}
	}

	if foundType == nil {
		fmt.Fprintf(os.Stderr, "Type not found: %s.%s\n", pkgName, typeName)
		fmt.Fprintf(os.Stderr, "Available types in %s:\n", pkgName)
		for _, t := range api.Types {
			fmt.Fprintf(os.Stderr, "  %s\n", t.Name)
		}
		os.Exit(1)
	}

	// Print type header
	fmt.Printf("%s.%s", pkgName, foundType.Name)
	if foundType.Kind != "" && foundType.Kind != "struct" {
		fmt.Printf(" (%s)", foundType.Kind)
	}
	fmt.Println()
	fmt.Println(strings.Repeat("-", len(pkgName)+len(foundType.Name)+1))

	if foundType.Doc != "" {
		fmt.Println()
		fmt.Println(foundType.Doc)
	}

	// Show fields for structs
	if len(foundType.Fields) > 0 {
		fmt.Println()
		fmt.Println("Fields:")
		for _, f := range foundType.Fields {
			fmt.Printf("  %-20s %s", f.Name, f.Type)
			if f.Doc != "" {
				fmt.Printf("  // %s", f.Doc)
			}
			fmt.Println()
		}
	}

	// Find constructors (functions returning this type)
	constructors := []FuncAPI{}
	for _, f := range api.Functions {
		if !f.IsMethod && strings.Contains(f.Signature, "*"+foundType.Name) {
			constructors = append(constructors, f)
		}
	}

	if len(constructors) > 0 {
		fmt.Println()
		fmt.Println("Constructors:")
		for _, c := range constructors {
			fmt.Printf("  %s\n", c.Signature)
			if c.Doc != "" {
				fmt.Printf("    %s\n", firstLine(c.Doc))
			}
		}
	}

	// Show methods
	if len(foundType.Methods) > 0 || showMethods {
		fmt.Println()
		fmt.Println("Methods:")
		if len(foundType.Methods) == 0 {
			fmt.Println("  (none)")
		}
		for _, m := range foundType.Methods {
			fmt.Printf("  %s\n", m.Signature)
			if m.Doc != "" && showMethods {
				fmt.Printf("    %s\n", firstLine(m.Doc))
			}
		}
	}

	// See also (other types in package)
	otherTypes := []string{}
	for _, t := range api.Types {
		if t.Name != foundType.Name {
			otherTypes = append(otherTypes, t.Name)
		}
	}
	if len(otherTypes) > 0 && len(otherTypes) <= 5 {
		fmt.Println()
		fmt.Printf("See also: %s\n", strings.Join(otherTypes, ", "))
	}
}

func searchAPIs(query string) {
	query = strings.ToLower(query)
	fmt.Printf("Searching for: %s\n", query)
	fmt.Println()

	found := false

	for pkgName, pkgPath := range enginePackages {
		api, err := extractPackageAPI(pkgPath)
		if err != nil {
			continue
		}

		pkgMatches := []string{}

		// Search types
		for _, t := range api.Types {
			if matchesQuery(t.Name, t.Doc, query) {
				pkgMatches = append(pkgMatches, fmt.Sprintf("  type %s.%s", pkgName, t.Name))
			}

			// Search methods
			for _, m := range t.Methods {
				if matchesQuery(m.Name, m.Doc, query) {
					pkgMatches = append(pkgMatches, fmt.Sprintf("  method %s.%s.%s", pkgName, t.Name, m.Name))
				}
			}
		}

		// Search functions
		for _, f := range api.Functions {
			if !f.IsMethod && matchesQuery(f.Name, f.Doc, query) {
				pkgMatches = append(pkgMatches, fmt.Sprintf("  func %s.%s", pkgName, f.Name))
			}
		}

		if len(pkgMatches) > 0 {
			found = true
			fmt.Printf("%s (%s):\n", pkgName, pkgPath)
			for _, m := range pkgMatches {
				fmt.Println(m)
			}
			fmt.Println()
		}
	}

	if !found {
		fmt.Println("No matches found.")
	}
}

func matchesQuery(name, doc, query string) bool {
	return strings.Contains(strings.ToLower(name), query) ||
		strings.Contains(strings.ToLower(doc), query)
}

func extractPackageAPI(pkgPath string) (*PackageAPI, error) {
	fset := token.NewFileSet()

	// Parse all Go files in the package directory
	pkgs, err := parser.ParseDir(fset, pkgPath, func(fi os.FileInfo) bool {
		// Skip test files
		return !strings.HasSuffix(fi.Name(), "_test.go")
	}, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	api := &PackageAPI{
		Path: pkgPath,
	}

	// Combine all packages (usually just one, but handle edge cases)
	for pkgName, pkg := range pkgs {
		api.Name = pkgName

		for _, file := range pkg.Files {
			extractFileAPI(file, api)
		}
	}

	// Sort for consistent output
	sort.Slice(api.Types, func(i, j int) bool {
		return api.Types[i].Name < api.Types[j].Name
	})
	sort.Slice(api.Functions, func(i, j int) bool {
		return api.Functions[i].Name < api.Functions[j].Name
	})

	return api, nil
}

func extractFileAPI(file *ast.File, api *PackageAPI) {
	// Map to collect methods by receiver type
	methodsByType := make(map[string][]FuncAPI)

	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					if !ast.IsExported(s.Name.Name) {
						continue
					}
					t := extractTypeSpec(s, d.Doc)
					api.Types = append(api.Types, t)
				}
			}
		case *ast.FuncDecl:
			if !ast.IsExported(d.Name.Name) {
				continue
			}
			f := extractFuncDecl(d)
			api.Functions = append(api.Functions, f)

			if f.IsMethod {
				methodsByType[f.ReceiverTy] = append(methodsByType[f.ReceiverTy], f)
			}
		}
	}

	// Attach methods to their types
	for i := range api.Types {
		typeName := api.Types[i].Name
		if methods, ok := methodsByType[typeName]; ok {
			api.Types[i].Methods = append(api.Types[i].Methods, methods...)
		}
	}
}

func extractTypeSpec(spec *ast.TypeSpec, doc *ast.CommentGroup) TypeAPI {
	t := TypeAPI{
		Name: spec.Name.Name,
		Doc:  extractDoc(doc),
	}

	switch ty := spec.Type.(type) {
	case *ast.StructType:
		t.Kind = "struct"
		t.Fields = extractFields(ty.Fields)
	case *ast.InterfaceType:
		t.Kind = "interface"
	case *ast.Ident:
		t.Kind = "alias"
		t.Underlying = ty.Name
	case *ast.SelectorExpr:
		t.Kind = "alias"
		t.Underlying = exprToString(ty)
	default:
		t.Kind = "type"
		t.Underlying = exprToString(ty)
	}

	return t
}

func extractFields(fields *ast.FieldList) []FieldAPI {
	if fields == nil {
		return nil
	}

	result := []FieldAPI{}
	for _, field := range fields.List {
		// Handle embedded fields (no names)
		if len(field.Names) == 0 {
			f := FieldAPI{
				Name: exprToString(field.Type),
				Type: "(embedded)",
				Doc:  extractDoc(field.Doc),
			}
			result = append(result, f)
			continue
		}

		for _, name := range field.Names {
			if !ast.IsExported(name.Name) {
				continue
			}
			f := FieldAPI{
				Name: name.Name,
				Type: exprToString(field.Type),
				Doc:  extractDoc(field.Doc),
			}
			if field.Tag != nil {
				f.Tag = field.Tag.Value
			}
			result = append(result, f)
		}
	}
	return result
}

func extractFuncDecl(decl *ast.FuncDecl) FuncAPI {
	f := FuncAPI{
		Name: decl.Name.Name,
		Doc:  extractDoc(decl.Doc),
	}

	// Check if it's a method
	if decl.Recv != nil && len(decl.Recv.List) > 0 {
		f.IsMethod = true
		recv := decl.Recv.List[0]
		f.ReceiverTy = extractReceiverType(recv.Type)
	}

	// Build signature
	f.Signature = buildFuncSignature(decl)

	return f
}

func extractReceiverType(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.StarExpr:
		return extractReceiverType(t.X)
	case *ast.Ident:
		return t.Name
	default:
		return exprToString(expr)
	}
}

func buildFuncSignature(decl *ast.FuncDecl) string {
	var b strings.Builder

	b.WriteString("func ")

	// Receiver
	if decl.Recv != nil && len(decl.Recv.List) > 0 {
		recv := decl.Recv.List[0]
		var recvName string
		if len(recv.Names) > 0 {
			recvName = recv.Names[0].Name
		}
		b.WriteString(fmt.Sprintf("(%s %s) ", recvName, exprToString(recv.Type)))
	}

	// Name
	b.WriteString(decl.Name.Name)

	// Parameters
	b.WriteString("(")
	if decl.Type.Params != nil {
		params := []string{}
		for _, p := range decl.Type.Params.List {
			paramType := exprToString(p.Type)
			if len(p.Names) == 0 {
				params = append(params, paramType)
			} else {
				for _, name := range p.Names {
					params = append(params, fmt.Sprintf("%s %s", name.Name, paramType))
				}
			}
		}
		b.WriteString(strings.Join(params, ", "))
	}
	b.WriteString(")")

	// Return type
	if decl.Type.Results != nil && len(decl.Type.Results.List) > 0 {
		results := []string{}
		for _, r := range decl.Type.Results.List {
			results = append(results, exprToString(r.Type))
		}
		if len(results) == 1 {
			b.WriteString(" " + results[0])
		} else {
			b.WriteString(" (" + strings.Join(results, ", ") + ")")
		}
	}

	return b.String()
}

func exprToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + exprToString(t.X)
	case *ast.ArrayType:
		if t.Len == nil {
			return "[]" + exprToString(t.Elt)
		}
		return fmt.Sprintf("[%s]%s", exprToString(t.Len), exprToString(t.Elt))
	case *ast.MapType:
		return fmt.Sprintf("map[%s]%s", exprToString(t.Key), exprToString(t.Value))
	case *ast.SelectorExpr:
		return exprToString(t.X) + "." + t.Sel.Name
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.FuncType:
		return "func(...)"
	case *ast.ChanType:
		return "chan " + exprToString(t.Value)
	case *ast.Ellipsis:
		return "..." + exprToString(t.Elt)
	case *ast.BasicLit:
		return t.Value
	default:
		return "?"
	}
}

func extractDoc(doc *ast.CommentGroup) string {
	if doc == nil {
		return ""
	}
	return strings.TrimSpace(doc.Text())
}

func firstLine(s string) string {
	if idx := strings.Index(s, "\n"); idx != -1 {
		return s[:idx]
	}
	return s
}

func getPackageNames() []string {
	names := make([]string, 0, len(enginePackages))
	for name := range enginePackages {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
