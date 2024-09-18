package routergen

import (
	"fmt"
	"github.com/houyanzu/work-box/app/boxgen/constdef"
	"github.com/houyanzu/work-box/app/boxgen/toolfunc"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

var imports []string
var inits []string

func Routergen(root string) {
	var err error

	// 打开或创建 register.go 文件
	file, err := os.Create(root + "register.go")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	_, err = file.WriteString(constdef.RouterHeaderStr)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}

	// 写入 package 声明
	_, err = file.WriteString(constdef.PackageMainStr)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}

	imports = make([]string, 0)
	inits = make([]string, 0)

	err = scanDirectories(root)
	if err != nil {
		fmt.Println("Error scanning directories:", err)
		return
	}

	// 写入导入语句
	importStr2 := constdef.RouterImportStr
	for _, v := range imports {
		importStr2 += "\t" + v + "\n"
	}
	importStr2 += ")\n\n"
	_, err = file.WriteString(importStr2)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}

	// 写入函数定义
	_, err = file.WriteString("var controllers []interface{}\n\n")
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}

	_, err = file.WriteString(constdef.RouterRegistControllerStr)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}

	ss := ""
	for _, v := range inits {
		// 写入函数定义
		ss += v
	}
	ss = "func init() {" + ss + "\n}\n"
	_, err = file.WriteString(ss)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}

	otherStr := constdef.RouterOtherStr
	_, err = file.WriteString(otherStr)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}

	fmt.Println("Register file generated successfully.")
}

// 递归遍历目录，处理 controller 目录中的 Go 文件
func scanDirectories(root string) error {
	return filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() && d.Name() == "controller" {
			// 遇到 controller 目录，处理其中的 Go 文件
			return filepath.WalkDir(path, func(path string, d os.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if !d.IsDir() && strings.HasSuffix(d.Name(), ".go") {
					return processGoFile(path)
				}
				return nil
			})
		}
		return nil
	})
}

// 处理 Go 文件中的控制器
func processGoFile(filePath string) error {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.AllErrors)
	if err != nil {
		return err
	}

	module, err := toolfunc.GetModuleName()
	if err != nil {
		return err
	}

	pak := toolfunc.GetImportPkg(module, filePath)
	alias := fmt.Sprintf("controller%d", len(imports))
	have := false
	for _, decl := range node.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			if toolfunc.IsControllerType(typeSpec.Name.Name) {
				controllerName := typeSpec.Name.Name
				controllerName = alias + "." + controllerName
				// Write out code to register the controller
				inits = append(inits, fmt.Sprintf("\n\tRegisterController(%s{})", controllerName))

				have = true
			}
		}
	}
	if have {
		imports = append(imports, alias+" \""+pak+"\"")
	}

	return nil
}
