package gin

import (
	"os"
	"reflect"
	"runtime"
)

// H is a shortcut for map[string]any
type H map[string]any

func assert1(guard bool, text string) {
	if !guard {
		panic(text)
	}
}

func nameOfFunction(f any) string {
	return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
}

func resolveAddress(addr []string) string {
	switch len(addr) {
	case 0:
		if port := os.Getenv("PORT"); port != "" {
			debugPrint("Environment variable PORT=\"%s\"", port)
			return ":" + port
		}
		debugPrint("Environment variable PORT is undefined. Using port :8080 by default")
		return ":8080"
	case 1:
		return addr[0]
	default:
		panic("too many parameters")
	}
}
