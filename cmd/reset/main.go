package main

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

type undefValueType string

var undefValue undefValueType = "undef"

var basicTypes = map[string]bool{
	"bool": true, "string": true,
	"int": true, "int8": true, "int16": true, "int32": true, "int64": true,
	"uint": true, "uint8": true, "uint16": true, "uint32": true, "uint64": true,
	"float32": true, "float64": true,
	"complex64": true, "complex128": true,
	"byte": true, "rune": true,
	"error": true,
}

func getDefValueBasicTypes(typeName string) string {
	val := string(undefValue)
	switch typeName {
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"uintptr", "byte", "rune":
		val = "0"
	case "float32", "float64":
		val = "0.0"
	case "complex64", "complex128":
		val = "0+0i"
	case "bool":
		val = "false"
	case "string":
		val = `""`
	}

	return val
}

func processFile(filename string) (string, string, error) {

	fset := token.NewFileSet()

	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return "", "", err
	}

	//Имя пакета
	packageName := node.Name.Name

	//Код методов
	var src strings.Builder

	ast.Inspect(node, func(n ast.Node) bool {
		//Нужны только спейификации структур
		genDecl, ok := n.(*ast.GenDecl)
		if !ok {
			return true
		}

		//Нужны только структуры с телом (не переопределение типа)
		if genDecl.Doc != nil {
			for _, comment := range genDecl.Doc.List {
				if strings.Contains(comment.Text, "generate:reset") {
					for _, spec := range genDecl.Specs {
						//Нужны только структуры
						typeSpec, ok := spec.(*ast.TypeSpec)
						if !ok {
							return true
						}

						//Нужны только структуры с телом (не переопределение типа)
						structType, ok := typeSpec.Type.(*ast.StructType)
						if !ok {
							return true
						}

						//fmt.Println(packageName, fset.Position(n.Pos()), ">>>", typeSpec, "-----", structType)
						err := createResetMethod(&src, typeSpec.Name.Name, structType.Fields.List)
						if err != nil {
							fmt.Printf("Reset generation: error generation reset method for %s (%s): %v", fset.Position(n.Pos()), typeSpec.Name.Name, err)
						}

					}
				}
			}
		}

		return true

	})

	return packageName, src.String(), nil
}

func createResetMethod(src *strings.Builder, structName string, structFields []*ast.Field) error {

	//	   rs.i = 0
	//	   rs.str = ""
	//	   if rs.strP != nil {
	//	       *rs.strP = ""
	//	   }
	//	   rs.s = rs.s[:0]
	//	   clear(rs.m)
	//	   if resetter, ok := rs.child.(interface{ Reset() }); ok && rs.child != nil {
	//	       resetter.Reset()
	//	   }
	//	}

	fmt.Fprintf(src, "\n")
	fmt.Fprintf(src, "func (rs *%s) Reset() {", structName)
	fmt.Fprintf(src, "if rs == nil {return}")
	fmt.Fprintf(src, "\n")

	//Вот тут надо обойти все поля
	for _, field := range structFields {
		if len(field.Names) > 0 { //Пустое имя поля это встраивание
			for _, name := range field.Names {

				//fmt.Println(">>", name, field.Type)

				switch t := field.Type.(type) {
				case *ast.Ident:
					val := getDefValueBasicTypes(t.Name)

					if val == string(undefValue) {
						fmt.Fprintf(src, "//rs.%s=nil //UNKNOWN TYPE (type %s)\n", name, t.Name)
					} else {
						fmt.Fprintf(src, "rs.%s=%s\n", name, val)
					}
				case *ast.ArrayType:
					fmt.Fprintf(src, "rs.%s=rs.%s[:0]\n", name, name)
				case *ast.MapType:
					fmt.Fprintf(src, "clear(rs.%s)\n", name)
				case *ast.StarExpr:
					//Чтобы провалился тот чел который придумал это задание
					switch subt := t.X.(type) {
					case *ast.Ident:
						if basicTypes[subt.Name] {
							//Указатель на базовый тип
							val := getDefValueBasicTypes(subt.Name)
							fmt.Fprintf(src, "if rs.%s!=nil {\n", name)
							fmt.Fprintf(src, "*rs.%s=%s\n", name, val)
							fmt.Fprintf(src, "}\n")
						} else {
							//Указатель на пользовательский тип (структуры в этом же пакете)
							fmt.Fprintf(src, "if resetter, ok := checkResetMethod(rs.%s); ok && rs.%s != nil {\n", name, name)
							fmt.Fprintf(src, "resetter.Reset()")
							fmt.Fprintf(src, "}\n")
						}
					case *ast.SelectorExpr:
						//Указатель на структуры в другом пакете
						fmt.Fprintf(src, "if resetter, ok := checkResetMethod(rs.%s); ok && rs.%s != nil {\n", name, name)
						fmt.Fprintf(src, "resetter.Reset()")
						fmt.Fprintf(src, "}\n")
					default:
						//Указатель на непонятно что (указатель на указатель...)
						fmt.Fprintf(src, "if rs.%s!=nil {\n", name)
						fmt.Fprintf(src, "*rs.%s=nil\n", name)
						fmt.Fprintf(src, "}\n")

					}

				case *ast.InterfaceType:
					fmt.Fprintf(src, "rs.%s=interface{} \n", name)
				default:
					fmt.Fprintf(src, "//rs.%s=nil //UNKNOWN TYPE FIELD (type %v)\n", name, field.Type)

				}
			}
		}
	}

	fmt.Fprintf(src, "")
	fmt.Fprintf(src, "}")

	return nil

}

func main() {

	//Решение в лоб - будем просто хранить список файлов которые мы сгенериировали
	fileList := make(map[string]bool, 10)

	//Обход директории
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		//Разбираем файл
		pack, src, err := processFile(path)
		if err != nil {
			return err
		}

		//Если ответ пустой то берем следующий файл
		if src == "" {
			return nil
		}

		//Имя файла с методами
		fileNameReset := filepath.Dir(path) + "/reset.gen.go"

		//Полное содержимое файла
		fileSrc := ""

		_, ok := fileList[fileNameReset]
		if !ok {
			//Создаем файл

			fileSrc = fileSrc + "// Code generated by go generate; DO NOT EDIT.\n"
			fileSrc = fileSrc + "// This file was generated by cmd/reset/main.go\n"
			fileSrc = fileSrc + "\n"
			fileSrc = fileSrc + "package " + pack + "\n"
			fileSrc = fileSrc + "\n"
			fileSrc = fileSrc + "func checkResetMethod(s interface{}) (interface{ Reset() }, bool) {\n"
			fileSrc = fileSrc + "meth, ok := s.(interface{ Reset() })\n"
			fileSrc = fileSrc + "return meth, ok\n"
			fileSrc = fileSrc + "}\n"
			fileSrc = fileSrc + "\n"

		} else {
			data, err := os.ReadFile(fileNameReset)
			if err != nil {
				//return err
				data = []byte("//error open file")
			}

			fileSrc = string(data)
		}

		fileSrc = fileSrc + src
		generated := []byte(fileSrc)
		formatted, err := format.Source(generated)

		if err != nil {
			//fmt.Println("--------------------------")
			//fmt.Println(fileSrc)
			//fmt.Println("--------------------------")
			return err
		}

		//fmt.Println(string(formatted))

		err = os.WriteFile(fileNameReset, formatted, 0644)
		if err != nil {
			return err
		}

		fileList[fileNameReset] = true

		return nil

	})

	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
