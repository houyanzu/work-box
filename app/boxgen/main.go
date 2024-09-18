package main

import (
	"flag"
	"fmt"
	"github.com/houyanzu/work-box/app/boxgen/controllergen"
	"github.com/houyanzu/work-box/app/boxgen/routergen"
)

func main() {
	var action string
	var root string
	var methods string
	var route string
	flag.StringVar(&root, "root", "app/api/home/", "root dir")
	flag.StringVar(&route, "route", "", "route")
	flag.StringVar(&action, "action", "", "action")
	flag.StringVar(&methods, "methods", "", "methods")
	flag.Parse()

	switch action {
	case "addController":
		controllergen.OperateController(root, route, action, methods)
	case "addMethods":
		controllergen.OperateController(root, route, action, methods)
	case "routergen":
		routergen.Routergen(root)
	default:
		fmt.Println("action not found")
	}
}
