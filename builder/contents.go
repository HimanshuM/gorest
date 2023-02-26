package builder

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/template"
)

type RoutesContent struct {
	Package string
	Imports string
	Lines   []string
	Level   string
}

func writeRoutesFile(path, level, pkg, pkgPath string, leaf *AST) error {
	path += level + ".go"
	hnd, err := os.Create(path)
	if err != nil {
		return err
	}
	if err = writeRoutesContent(hnd, level, pkg, pkgPath, leaf); err != nil {
		return err
	}
	if err = hnd.Close(); err != nil {
		return err
	}
	return nil
}

func writeRoutesContent(hnd *os.File, level, pkg, pkgPath string, leaf *AST) error {
	tpl, err := template.ParseFiles("tpl/base_routes.tpl")
	if err != nil {
		return err
	}
	content, err := buildRoutesContent(level, pkg, pkgPath, leaf)
	if err != nil {
		return err
	}
	return tpl.Execute(hnd, content)
}

func buildRoutesContent(level, pkg, pkgPath string, leaf *AST) (*RoutesContent, error) {
	importStrings := []string{"github.com/gin-gonic/gin"}
	if len(leaf.Tree) > 0 {
		importStrings = append(importStrings, fmt.Sprintf("%s/%s", pkgPath, level))
	}
	imports, err := imports(importStrings...)
	if err != nil {
		return nil, err
	}
	ctn := &RoutesContent{
		Package: pkg,
		Imports: imports,
		Level:   strings.Title(level),
	}
	ctn.linesFromRoute(leaf, level)
	return ctn, nil
}

func (ctn *RoutesContent) linesFromRoute(leaf *AST, level string) {
	i := 0
	ctn.Lines = make([]string, len(leaf.Tree)+len(leaf.Node.Methods))
	for _, node := range leaf.Tree {
		ctn.Lines[i] = fmt.Sprintf("%s.Setup%sRoutes(server)", level, strings.Title(cleanupRoute(node.Level)))
		i++
	}
	for method, def := range leaf.Node.Methods {
		ctn.Lines[i] = fmt.Sprintf("server.%s(\"%s\", %s)", method, leaf.Node.URL, def.Handler)
		i++
	}
}

func imports(pkgs ...string) (string, error) {
	if len(pkgs) == 0 {
		return "", nil
	}
	if len(pkgs) == 1 {
		return writeImportContent(pkgs[0])
	}
	return writeImportsContent(pkgs)
}

func writeImportContent(pkg string) (string, error) {
	tpl, err := template.ParseFiles("tpl/import.tpl")
	if err != nil {
		return "", err
	}
	var content bytes.Buffer
	if err = tpl.Execute(&content, pkg); err != nil {
		return "", err
	}
	return content.String(), nil
}

func writeImportsContent(pkgs []string) (string, error) {
	tpl, err := template.ParseFiles("tpl/imports.tpl")
	if err != nil {
		return "", err
	}
	var content bytes.Buffer
	if err = tpl.Execute(&content, pkgs); err != nil {
		return "", err
	}
	return content.String(), nil
}
