package generator

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

type Generator struct {
	tmpl *templateImpl
}

func NewGenerator() (*Generator, error) {
	tmpl, err := newTemplate()
	if err != nil {
		return nil, err
	}
	return &Generator{tmpl: tmpl}, nil
}

// processPackage обрабатывает один пакет
func (g *Generator) processPackage(pkgPath string) error {
	structs, pkgName, err := g.findResetStructs(pkgPath)
	if err != nil {
		return fmt.Errorf("failed to parse package %s: %w", pkgPath, err)
	}

	if len(structs) > 0 {
		if err := g.generatePackageFile(pkgPath, pkgName, structs); err != nil {
			return fmt.Errorf("failed to generate file for package %s: %w", pkgPath, err)
		}
	}

	return nil
}

// findResetStructs находит структуры с комментарием // generate:reset
func (g *Generator) findResetStructs(pkgPath string) ([]TemplateEnum, string, error) {
	fset := token.NewFileSet()
	var structs []TemplateEnum
	var pkgName string

	pattern := filepath.Join(pkgPath, "*.go")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, "", fmt.Errorf("failed to glob files in %s: %w", pkgPath, err)
	}

	for _, filename := range files {
		base := filepath.Base(filename)
		if strings.HasSuffix(base, "_test.go") ||
			strings.HasSuffix(base, ".gen.go") ||
			strings.Contains(base, "_generated") {
			continue
		}

		file, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
		if err != nil {
			return nil, "", fmt.Errorf("failed to parse file %s: %w", filename, err)
		}

		if pkgName == "" {
			pkgName = file.Name.Name
		}

		structs = append(structs, g.findStructsInFile(file)...)
	}

	return structs, pkgName, nil
}

// findStructsInFile находит структуры с комментарием reset в одном файле
func (g *Generator) findStructsInFile(file *ast.File) []TemplateEnum {
	var structs []TemplateEnum

	ast.Inspect(file, func(n ast.Node) bool {
		genDecl, ok := n.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			return true
		}

		if !g.hasResetComment(genDecl) {
			return true
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

			structInfo := TemplateEnum{
				StructType: typeSpec.Name.Name,
				Entries:    g.collectStructFields(structType),
			}

			structs = append(structs, structInfo)
		}

		return true
	})

	return structs
}

// hasResetComment проверяет наличие комментария generate:reset
func (g *Generator) hasResetComment(genDecl *ast.GenDecl) bool {
	if genDecl.Doc == nil {
		return false
	}

	for _, comment := range genDecl.Doc.List {
		if strings.Contains(comment.Text, "generate:reset") {
			return true
		}
	}
	return false
}

// collectStructFields собирает информацию о полях структуры
func (g *Generator) collectStructFields(structType *ast.StructType) []TemplateEntries {
	var entries []TemplateEntries

	for _, field := range structType.Fields.List {
		for _, fieldName := range field.Names {
			entry := g.analyzeFieldType(fieldName.Name, field.Type)
			entries = append(entries, entry)
		}
	}

	return entries
}

// analyzeFieldType анализирует тип поля и возвращает информацию о нем
func (g *Generator) analyzeFieldType(fieldName string, expr ast.Expr) TemplateEntries {
	entry := TemplateEntries{
		Name: fieldName,
		Type: g.exprToString(expr),
	}

	entry.IsPointer = g.isPointerType(expr)

	baseExpr := expr
	if entry.IsPointer {
		if starExpr, ok := expr.(*ast.StarExpr); ok {
			baseExpr = starExpr.X
		}
	}

	entry.IsSlice = g.isSliceType(baseExpr)
	entry.IsMap = g.isMapType(baseExpr)

	if entry.IsSlice || entry.IsMap {
		entry.IsPrimitive = false
		entry.IsStruct = false
	} else {
		baseType := g.getBaseType(expr)
		entry.IsPrimitive = g.isPrimitiveType(baseType)
		entry.IsStruct = !entry.IsPrimitive

		if entry.IsPrimitive {
			entry.ZeroValue = getPrimitiveZeroValue(baseType)
		}
	}

	return entry
}

// exprToString преобразует AST выражение в строку
func (g *Generator) exprToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + g.exprToString(t.X)
	case *ast.ArrayType:
		if t.Len == nil {
			return "[]" + g.exprToString(t.Elt)
		}
		return "[" + g.exprToString(t.Len) + "]" + g.exprToString(t.Elt)
	case *ast.MapType:
		return "map[" + g.exprToString(t.Key) + "]" + g.exprToString(t.Value)
	case *ast.SelectorExpr:
		return g.exprToString(t.X) + "." + t.Sel.Name
	case *ast.BasicLit:
		return t.Value
	default:
		return fmt.Sprintf("%T", t)
	}
}

// isPointerType проверяет, является ли тип указателем
func (g *Generator) isPointerType(expr ast.Expr) bool {
	_, ok := expr.(*ast.StarExpr)
	return ok
}

// isSliceType проверяет, является ли тип слайсом
func (g *Generator) isSliceType(expr ast.Expr) bool {
	arrType, ok := expr.(*ast.ArrayType)
	return ok && arrType.Len == nil
}

// isMapType проверяет, является ли тип мапой
func (g *Generator) isMapType(expr ast.Expr) bool {
	_, ok := expr.(*ast.MapType)
	return ok
}

// isPrimitiveType проверяет, является ли тип примитивным
func (g *Generator) isPrimitiveType(typeStr string) bool {
	primitiveTypes := map[string]bool{
		"bool":   true,
		"string": true,
		"int":    true, "int8": true, "int16": true, "int32": true, "int64": true,
		"uint": true, "uint8": true, "uint16": true, "uint32": true, "uint64": true, "uintptr": true,
		"float32": true, "float64": true,
		"complex64": true, "complex128": true,
		"byte": true, "rune": true,
	}

	return primitiveTypes[typeStr]
}

// getBaseType возвращает базовый тип
func (g *Generator) getBaseType(expr ast.Expr) string {
	for {
		if ptr, ok := expr.(*ast.StarExpr); ok {
			expr = ptr.X
		} else {
			break
		}
	}

	return g.exprToString(expr)
}

// generatePackageFile генерирует файл reset.gen.go для пакета
func (g *Generator) generatePackageFile(pkgPath, pkgName string, structs []TemplateEnum) error {
	outputFile := filepath.Join(pkgPath, "reset.gen.go")

	var buf bytes.Buffer
	buf.WriteString("// Code generated by go generate; DO NOT EDIT.\n")
	buf.WriteString("package " + pkgName + "\n\n")

	for _, structInfo := range structs {
		structInfoWithPackage := structInfo
		structInfoWithPackage.Package = pkgName

		err := g.tmpl.execute(&buf, structInfoWithPackage)
		if err != nil {
			return err
		}
		buf.WriteString("\n")
	}

	generatedCode := buf.Bytes()

	bufFmt, err := format.Source(generatedCode)
	if err != nil {
		return err
	}

	if err := os.WriteFile(outputFile, bufFmt, 0644); err != nil {
		return err
	}
	return nil
}

// WalkAndProcess рекурсивно обходит директории и обрабатывает пакеты
func (g *Generator) WalkAndProcess(rootDir string) error {
	return filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() && g.hasGoFiles(path) {
			if err := g.processPackage(path); err != nil {
				return err
			}
		}

		return nil
	})
}

// hasGoFiles проверяет, есть ли в директории Go файлы
func (g *Generator) hasGoFiles(dir string) bool {
	patterns := []string{
		filepath.Join(dir, "*.go"),
	}

	for _, pattern := range patterns {
		files, err := filepath.Glob(pattern)
		if err != nil {
			continue
		}

		for _, file := range files {
			base := filepath.Base(file)
			if !isTestFile(base) && !isGeneratedFile(base) {
				return true
			}
		}
	}

	return false
}

// isTestFile проверяет, является ли файл тестовым
func isTestFile(filename string) bool {
	return strings.HasSuffix(filename, "_test.go")
}

// isGeneratedFile проверяет, является ли файл сгенерированным
func isGeneratedFile(filename string) bool {
	return strings.HasSuffix(filename, ".gen.go") ||
		strings.Contains(filename, "_generated") ||
		strings.Contains(filename, "// Code generated")
}

func getPrimitiveZeroValue(typeStr string) string {
	switch typeStr {
	case "string":
		return `""`
	case "bool":
		return "false"
	default:
		return "0"
	}
}
