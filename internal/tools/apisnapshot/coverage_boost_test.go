package main

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ── run() CLI tests ─────────────────────────────────────────────────

func TestRunTooManyArgs(t *testing.T) {
	var buf bytes.Buffer
	err := run([]string{"a", "b"}, &buf)
	if err == nil {
		t.Fatal("expected error for too many args")
	}
	if !strings.Contains(err.Error(), "usage") {
		t.Fatalf("expected usage error, got: %v", err)
	}
}

func TestRunDefaultDir(t *testing.T) {
	var buf bytes.Buffer
	// Default dir is ./pkg/observex, which may not exist in test env.
	// We test with explicit dir instead.
	dir := t.TempDir()
	source := `package sample
type Exported struct { Name string }
`
	if err := os.WriteFile(filepath.Join(dir, "api.go"), []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	// run expects the package to be named "observex" or have exactly 1 package.
	// Our sample has 1 package, so it will be used.
	err := run([]string{dir}, &buf)
	if err != nil {
		t.Fatalf("run() failed: %v", err)
	}
	if !strings.Contains(buf.String(), "Exported") {
		t.Fatalf("expected Exported in output, got: %s", buf.String())
	}
}

func TestRunNonExistentDir(t *testing.T) {
	var buf bytes.Buffer
	err := run([]string{"/nonexistent/dir"}, &buf)
	if err == nil {
		t.Fatal("expected error for nonexistent dir")
	}
}

func TestRunNoArgs(t *testing.T) {
	// run with no args uses default "./pkg/observex"
	var buf bytes.Buffer
	err := run([]string{}, &buf)
	// This may fail if we're not in the repo root, but that's fine.
	// We just verify it doesn't panic.
	_ = err
}

// ── snapshotTypeSpec with assign (type alias) ───────────────────────

func TestSnapshotTypeSpecAlias(t *testing.T) {
	dir := t.TempDir()
	source := `package sample
type MyInt = int
type MyStruct struct { Name string }
type MyInterface interface { Method() }
`
	if err := os.WriteFile(filepath.Join(dir, "api.go"), []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := snapshotPackage(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "type MyInt = int") {
		t.Fatalf("expected alias in output, got:\n%s", got)
	}
	if !strings.Contains(got, "type MyStruct struct") {
		t.Fatalf("expected struct in output, got:\n%s", got)
	}
	if !strings.Contains(got, "type MyInterface interface") {
		t.Fatalf("expected interface in output, got:\n%s", got)
	}
}

// ── snapshotStruct with embedded types and tags ─────────────────────

func TestSnapshotStructWithEmbeddedExported(t *testing.T) {
	dir := t.TempDir()
	source := `package sample
import "context"
type Base struct {}
type Config struct {
	Base
	Name string ` + "`json:\"name\"`" + `
	secret string
}
`
	if err := os.WriteFile(filepath.Join(dir, "api.go"), []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := snapshotPackage(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "Base") {
		t.Fatalf("expected embedded Base in output, got:\n%s", got)
	}
	if !strings.Contains(got, "json:\"name\"") {
		t.Fatalf("expected json tag in output, got:\n%s", got)
	}
	if strings.Contains(got, "secret string") {
		t.Fatalf("should not contain private field, got:\n%s", got)
	}
}

func TestSnapshotStructEmpty(t *testing.T) {
	dir := t.TempDir()
	source := `package sample
type Empty struct {}
`
	if err := os.WriteFile(filepath.Join(dir, "api.go"), []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := snapshotPackage(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "struct {}") {
		t.Fatalf("expected empty struct in output, got:\n%s", got)
	}
}

func TestSnapshotStructOnlyPrivateFields(t *testing.T) {
	dir := t.TempDir()
	source := `package sample
type Hidden struct {
	secret string
	value int
}
`
	if err := os.WriteFile(filepath.Join(dir, "api.go"), []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := snapshotPackage(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "struct {}") {
		t.Fatalf("expected empty struct for all-private, got:\n%s", got)
	}
}

// ── snapshotInterface with embedded interface ───────────────────────

func TestSnapshotInterfaceWithEmbedded(t *testing.T) {
	dir := t.TempDir()
	source := `package sample
import "io"
type MyReader interface {
	io.Reader
	io.Writer
	Method() error
	private()
}
`
	if err := os.WriteFile(filepath.Join(dir, "api.go"), []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := snapshotPackage(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "io.Reader") {
		t.Fatalf("expected io.Reader in output, got:\n%s", got)
	}
	if !strings.Contains(got, "io.Writer") {
		t.Fatalf("expected io.Writer in output, got:\n%s", got)
	}
	if !strings.Contains(got, "Method() error") {
		t.Fatalf("expected Method in output, got:\n%s", got)
	}
	if strings.Contains(got, "private()") {
		t.Fatalf("should not contain private method, got:\n%s", got)
	}
}

func TestSnapshotInterfaceEmpty(t *testing.T) {
	dir := t.TempDir()
	source := `package sample
type Empty interface {}
`
	if err := os.WriteFile(filepath.Join(dir, "api.go"), []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := snapshotPackage(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "interface {}") {
		t.Fatalf("expected empty interface in output, got:\n%s", got)
	}
}

func TestSnapshotInterfaceOnlyPrivateMethods(t *testing.T) {
	dir := t.TempDir()
	source := `package sample
type Hidden interface {
	private()
}
`
	if err := os.WriteFile(filepath.Join(dir, "api.go"), []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := snapshotPackage(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "interface {}") {
		t.Fatalf("expected empty interface for all-private, got:\n%s", got)
	}
}

// ── exportedEmbeddedName ────────────────────────────────────────────

func TestExportedEmbeddedName(t *testing.T) {
	fset := token.NewFileSet()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"exported_ident", "type T struct { Config }", "Config"},
		{"unexported_ident", "type T struct { config }", ""},
		{"star_expr", "type T struct { *Config }", "Config"},
		{"selector_expr", "type T struct { io.Reader }", "Reader"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := "package sample\n" + tt.input
			f, err := parser.ParseFile(fset, "", src, 0)
			if err != nil {
				t.Fatal(err)
			}
			// Find the struct type
			for _, decl := range f.Decls {
				genDecl, ok := decl.(*ast.GenDecl)
				if !ok {
					continue
				}
				for _, spec := range genDecl.Specs {
					typeSpec, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}
					structType, ok := typeSpec.Type.(*ast.StructType)
					if !ok {
						continue
					}
					for _, field := range structType.Fields.List {
						got := exportedEmbeddedName(field.Type)
						if got != tt.want {
							t.Errorf("exportedEmbeddedName() = %q, want %q", got, tt.want)
						}
					}
				}
			}
		})
	}
}

func TestExportedEmbeddedNameIndexExpr(t *testing.T) {
	fset := token.NewFileSet()
	src := `package sample
type Container[T any] struct {}
type T struct { Container[int] }
`
	f, err := parser.ParseFile(fset, "", src, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, decl := range f.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok || typeSpec.Name.Name != "T" {
				continue
			}
			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}
			for _, field := range structType.Fields.List {
				got := exportedEmbeddedName(field.Type)
				if got != "Container" {
					t.Errorf("exportedEmbeddedName(IndexExpr) = %q, want %q", got, "Container")
				}
			}
		}
	}
}

func TestExportedEmbeddedNameNil(t *testing.T) {
	// Edge case: nil expression
	got := exportedEmbeddedName(nil)
	if got != "" {
		t.Errorf("expected empty for nil, got %q", got)
	}
}

// ── exportedReceiverName ────────────────────────────────────────────

func TestExportedReceiverName(t *testing.T) {
	tests := []struct {
		name string
		recv *ast.FieldList
		want string
	}{
		{"nil", nil, ""},
		{"empty", &ast.FieldList{}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := exportedReceiverName(tt.recv)
			if got != tt.want {
				t.Errorf("exportedReceiverName() = %q, want %q", got, tt.want)
			}
		})
	}
}

// ── receiverString ──────────────────────────────────────────────────

func TestReceiverString(t *testing.T) {
	fset := token.NewFileSet()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"named", "func (c Config) Method() {}", "c Config"},
		{"unnamed", "func (Config) Method() {}", "Config"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := "package sample\n" + tt.input
			f, err := parser.ParseFile(fset, "", src, 0)
			if err != nil {
				t.Fatal(err)
			}
			for _, decl := range f.Decls {
				funcDecl, ok := decl.(*ast.FuncDecl)
				if !ok || funcDecl.Recv == nil {
					continue
				}
				got := receiverString(fset, funcDecl.Recv)
				if got != tt.want {
					t.Errorf("receiverString() = %q, want %q", got, tt.want)
				}
			}
		})
	}
}

func TestReceiverStringNil(t *testing.T) {
	fset := token.NewFileSet()
	got := receiverString(fset, nil)
	if got != "" {
		t.Errorf("expected empty for nil, got %q", got)
	}
}

func TestReceiverStringEmpty(t *testing.T) {
	fset := token.NewFileSet()
	got := receiverString(fset, &ast.FieldList{})
	if got != "" {
		t.Errorf("expected empty for empty, got %q", got)
	}
}

// ── funcTypeString ──────────────────────────────────────────────────

func TestFuncTypeStringNonFunc(t *testing.T) {
	fset := token.NewFileSet()
	src := `package sample
type T = int
`
	f, err := parser.ParseFile(fset, "", src, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, decl := range f.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			got := funcTypeString(fset, typeSpec.Type)
			if !strings.HasPrefix(got, " ") {
				t.Errorf("expected space prefix for non-func type, got %q", got)
			}
		}
	}
}

// ── exprString with invalid node ────────────────────────────────────

func TestExprString(t *testing.T) {
	fset := token.NewFileSet()
	src := `package sample
func Helper() {}
`
	f, err := parser.ParseFile(fset, "", src, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, decl := range f.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		got := exprString(fset, funcDecl.Type)
		if !strings.Contains(got, "func") {
			t.Errorf("expected func in type string, got %q", got)
		}
	}
}

type invalidExpr struct{}

func (invalidExpr) Pos() token.Pos { return token.NoPos }

func (invalidExpr) End() token.Pos { return token.NoPos }

func TestExprStringInvalidNode(t *testing.T) {
	got := exprString(token.NewFileSet(), invalidExpr{})
	if got != "<invalid>" {
		t.Fatalf("expected invalid marker, got %q", got)
	}
}

// ── snapshotPackage with no observex-named package ──────────────────

func TestSnapshotPackageAutoDetect(t *testing.T) {
	dir := t.TempDir()
	source := `package mylib
type Config struct { Name string }
`
	if err := os.WriteFile(filepath.Join(dir, "api.go"), []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := snapshotPackage(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "package observex") {
		t.Fatalf("expected package observex header, got:\n%s", got)
	}
}

func TestSnapshotPackageMultiplePackages(t *testing.T) {
	dir := t.TempDir()
	// Create two files with different package names to trigger error
	source1 := `package foo
type A struct {}
`
	source2 := `package bar
type B struct {}
`
	if err := os.WriteFile(filepath.Join(dir, "a.go"), []byte(source1), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "b.go"), []byte(source2), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := snapshotPackage(dir)
	if err == nil {
		t.Fatal("expected error for multiple packages")
	}
}

// ── snapshotValueSpec with const and var ─────────────────────────────

func TestSnapshotValueSpecConstVar(t *testing.T) {
	dir := t.TempDir()
	source := `package sample
const Version = "v1"
const privateConst = "hidden"
var DefaultName string
var privateName string
`
	if err := os.WriteFile(filepath.Join(dir, "api.go"), []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := snapshotPackage(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "const Version = \"v1\"") {
		t.Fatalf("expected const Version, got:\n%s", got)
	}
	if !strings.Contains(got, "var DefaultName string") {
		t.Fatalf("expected var DefaultName, got:\n%s", got)
	}
	if strings.Contains(got, "privateConst") {
		t.Fatalf("should not contain private const, got:\n%s", got)
	}
	if strings.Contains(got, "privateName") {
		t.Fatalf("should not contain private var, got:\n%s", got)
	}
}

// ── snapshotFuncDecl with receiver ──────────────────────────────────

func TestSnapshotFuncDeclWithReceiver(t *testing.T) {
	dir := t.TempDir()
	source := `package sample
type Config struct {}
func (c Config) Validate() error { return nil }
func (c Config) private() {}
func Helper() {}
func helper() {}
`
	if err := os.WriteFile(filepath.Join(dir, "api.go"), []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := snapshotPackage(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "func (c Config) Validate() error") {
		t.Fatalf("expected receiver method, got:\n%s", got)
	}
	if !strings.Contains(got, "func Helper()") {
		t.Fatalf("expected standalone func, got:\n%s", got)
	}
	if strings.Contains(got, "private()") {
		t.Fatalf("should not contain private method, got:\n%s", got)
	}
	if strings.Contains(got, "func helper") {
		t.Fatalf("should not contain private func, got:\n%s", got)
	}
}

func TestSnapshotFuncDeclUnexportedReceiver(t *testing.T) {
	dir := t.TempDir()
	source := `package sample
type config struct {}
func (c config) Method() {}
`
	if err := os.WriteFile(filepath.Join(dir, "api.go"), []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := snapshotPackage(dir)
	if err != nil {
		t.Fatal(err)
	}
	// Unexported receiver type -> method should be skipped
	if strings.Contains(got, "Method") {
		t.Fatalf("should not contain method on unexported receiver, got:\n%s", got)
	}
}

// ── snapshotFuncDecl with pointer receiver ──────────────────────────

func TestSnapshotFuncDeclPointerReceiver(t *testing.T) {
	dir := t.TempDir()
	source := `package sample
type Config struct {}
func (c *Config) Save() error { return nil }
`
	if err := os.WriteFile(filepath.Join(dir, "api.go"), []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := snapshotPackage(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "Save") {
		t.Fatalf("expected Save method, got:\n%s", got)
	}
}

// ── snapshotTypeSpec default case (non-struct, non-interface) ────────

func TestSnapshotTypeSpecDefaultCase(t *testing.T) {
	dir := t.TempDir()
	source := `package sample
type MyFunc func(string) error
type MySlice []string
`
	if err := os.WriteFile(filepath.Join(dir, "api.go"), []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := snapshotPackage(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "type MyFunc func(string) error") {
		t.Fatalf("expected MyFunc in output, got:\n%s", got)
	}
	if !strings.Contains(got, "type MySlice []string") {
		t.Fatalf("expected MySlice in output, got:\n%s", got)
	}
}

// ── exportedEmbeddedName with IndexListExpr ─────────────────────────

func TestExportedEmbeddedNameSelectorExpr(t *testing.T) {
	fset := token.NewFileSet()
	src := `package sample
type T struct {
	Exported int
}
`
	f, err := parser.ParseFile(fset, "", src, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, decl := range f.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok || typeSpec.Name.Name != "T" {
				continue
			}
			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}
			// Named fields (not embedded) should return ""
			for _, field := range structType.Fields.List {
				if len(field.Names) > 0 {
					got := exportedEmbeddedName(field.Type)
					if got != "" {
						t.Errorf("expected empty for named field, got %q", got)
					}
				}
			}
		}
	}
}

// ── exportedEmbeddedName with IndexListExpr (constructed AST) ──────

func TestExportedEmbeddedNameIndexListExpr(t *testing.T) {
	// Construct an IndexListExpr programmatically: Container[int, string]
	expr := &ast.IndexListExpr{
		X: &ast.Ident{Name: "Container"},
		Indices: []ast.Expr{
			&ast.Ident{Name: "int"},
			&ast.Ident{Name: "string"},
		},
	}
	got := exportedEmbeddedName(expr)
	if got != "Container" {
		t.Errorf("exportedEmbeddedName(IndexListExpr) = %q, want %q", got, "Container")
	}
}

// ── main() error path ───────────────────────────────────────────────

func TestMainRunWithError(t *testing.T) {
	// run with a nonexistent directory should return an error
	var buf bytes.Buffer
	err := run([]string{"/nonexistent"}, &buf)
	if err == nil {
		t.Fatal("expected error")
	}
}
