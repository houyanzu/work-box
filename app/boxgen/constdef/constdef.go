package constdef

const PackageMainStr = "package main\n\n"

const ControllerImportStr = `
import (
	"github.com/gin-gonic/gin"
)

`

const ControllerFuncStr = `
// %s
func (co %s) %s(c *gin.Context) {
	//TODO: edit
}
`

const ControllerLoginFuncStr = `
// %s
func (co %s) %s(c *gin.Context, userID uint) {
	//TODO: edit
}
`
const RouterImportStr = `import (
	"fmt"
	"reflect"
	"strings"
	"unicode"
	"github.com/gin-gonic/gin"
	"github.com/houyanzu/work-box/tool/middleware"
`

const RouterHeaderStr = `// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.` + "\n\n"

const RouterRegistControllerStr = "func RegisterController(controller interface{}) {\n\tcontrollers = append(controllers, controller)\n}\n\n"

const RouterOtherStr = `
func Register(router *gin.Engine) {
	for _, v := range controllers {
		AutoRegisterRoutes(router, v)
	}
}

// 自动注册路由
func AutoRegisterRoutes(router *gin.Engine, controller interface{}) {
	controllerType := reflect.TypeOf(controller)
	controllerValue := reflect.ValueOf(controller)

	// 获取基础路由前缀
	baseRoute, pkgPath := buildBaseRoute(controllerType)

	// 遍历控制器的所有方法并注册路由
	for i := 0; i < controllerType.NumMethod(); i++ {
		method := controllerType.Method(i)
		methodName := getControllerRouterName(method.Name)
		methodType := method.Type
		numIn := methodType.NumIn()
		if numIn < 2 {
			continue
		}

		firstParamType := getParamTypeName(methodType.In(1))
		if firstParamType != "Context" {
			continue
		}

		key := pkgPath+"."+controllerType.Name()+"."+method.Name
		httpMethod := MethodTags[key]
		// 注册方法为 Gin 的 Post 路由
		route := fmt.Sprintf("api/%s/%s", baseRoute, methodName)
		switch numIn {
		case 2:
			switch httpMethod {
			case "POST":
				router.POST(route, func(ctx *gin.Context) {
					method.Func.Call([]reflect.Value{controllerValue, reflect.ValueOf(ctx)})
				})
			case "GET":
				router.GET(route, func(ctx *gin.Context) {
					method.Func.Call([]reflect.Value{controllerValue, reflect.ValueOf(ctx)})
				})
			default:
				router.POST(route, func(ctx *gin.Context) {
					method.Func.Call([]reflect.Value{controllerValue, reflect.ValueOf(ctx)})
				})
			}

		case 3:
			secondParamType := getParamTypeName(methodType.In(2))
			if secondParamType == "uint" {
				switch httpMethod {
				case "POST":
					router.POST(route, middleware.Login(), func(ctx *gin.Context) {
						userID := middleware.GetUserId(ctx)
						method.Func.Call([]reflect.Value{controllerValue, reflect.ValueOf(ctx), reflect.ValueOf(userID)})
					})
				case "GET":
					router.GET(route, middleware.Login(), func(ctx *gin.Context) {
						userID := middleware.GetUserId(ctx)
						method.Func.Call([]reflect.Value{controllerValue, reflect.ValueOf(ctx), reflect.ValueOf(userID)})
					})
				default:
					router.POST(route, middleware.Login(), func(ctx *gin.Context) {
						userID := middleware.GetUserId(ctx)
						method.Func.Call([]reflect.Value{controllerValue, reflect.ValueOf(ctx), reflect.ValueOf(userID)})
					})
				}

			}
		}

	}
}


// 构建控制器的基础路由前缀
func buildBaseRoute(controllerType reflect.Type) (string, string) {
	// 解析包路径
	pkgPath := controllerType.PkgPath()
	pkgParts := strings.Split(pkgPath, "/")

	// 构建路由路径
	var routeBuilder strings.Builder
	flag := false

	for _, part := range pkgParts {
		if part == "controller" {
			break
		}
		if flag {
			routeBuilder.WriteString(part + "/")
		}
		if part == "home" {
			flag = true
		}
	}

	// 添加控制器名
	controllerName := getControllerRouterName(controllerType.Name())
	routeBuilder.WriteString(controllerName)

	return routeBuilder.String(), pkgPath
}

// 去掉字符串末尾的 "Controller" 并将首字母变为小写
func getControllerRouterName(input string) string {
	input = strings.TrimSuffix(input, "Controller")
	if input == "" {
		return input
	}

	runes := []rune(input)
	runes[0] = unicode.ToLower(runes[0])

	return string(runes)
}

func getParamTypeName(paramType reflect.Type) string {
	if paramType.Kind() == reflect.Ptr {
		return paramType.Elem().Name()
	} else {
		return paramType.Name()
	}
}

`
