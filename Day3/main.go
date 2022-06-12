package main

import (
	"log"
	"reflect"
	"runtime"
)

func GetFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

func main() {
	mainWithFunctionName(foo)()
}

func foo() {
	log.Println("yoyo")
}

func mainWithFunctionName(fn func()) func() {
	return func() {
		functionName := GetFunctionName(fn)
		log.Printf("%v start", functionName)
		defer log.Printf("%v done", functionName)
		fn()
	}
}
