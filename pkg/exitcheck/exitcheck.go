package exitcheck

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// ExitAnalyzer анализатор для проверки os.Exit в main
var ExitAnalyzer = &analysis.Analyzer{
	Name: "exitcheck",
	Doc:  "checking os.exit in main.main",
	Run:  run,
}

// Основная функция проверки
func run(pass *analysis.Pass) (interface{}, error) {

	pkgName := pass.Pkg.Name()

	// Проверяем, что это пакет main
	if pkgName != "main" {
		return nil, nil
	}

	for _, file := range pass.Files {

		//fmt.Println(pass.Fset.Position(file.Pos()).Filename)

		// Убираем проверку из .cache т.к. там есть сгенерированные файлы
		filePath := pass.Fset.Position(file.Pos()).Filename
		if strings.Contains(filePath, ".cache") {
			continue
		}

		// функцией ast.Inspect проходим по всем узлам AST
		ast.Inspect(file, func(node ast.Node) bool {
			if funcDecl, ok := node.(*ast.FuncDecl); ok {
				// Получаем имя функции
				funcName := funcDecl.Name.Name

				// Ищем функцию main пакета main
				if funcName == "main" && funcDecl.Recv == nil {
					checkExitInMain(pass, funcDecl)
					return false //в файле может быть только одна функция main
				}
			}
			return true
		})
	}

	return nil, nil
}

// Функция проверки содержимого main.main
func checkExitInMain(pass *analysis.Pass, mainFunc *ast.FuncDecl) {
	ast.Inspect(mainFunc, func(node ast.Node) bool {
		switch exp := node.(type) {
		case *ast.CallExpr:
			if selExpr, ok := exp.Fun.(*ast.SelectorExpr); ok {
				if ident, ok := selExpr.X.(*ast.Ident); ok {
					if ident.Name == "os" && selExpr.Sel.Name == "Exit" {
						pass.Reportf(exp.Fun.Pos(), "os.Exit() call discover in main.main")
					}
				}
			}

		}

		return true
	})
}
