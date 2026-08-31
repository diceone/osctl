package main

import (
	"fmt"
	"runtime"
)

// buildVersion can be overridden at build time:
//
//	go build -ldflags "-X main.buildVersion=1.2.3" ...
var buildVersion = "0.3.0"

func getVersion() string {
	return fmt.Sprintf("osctl %s (%s/%s, %s)", buildVersion, runtime.GOOS, runtime.GOARCH, runtime.Version())
}
