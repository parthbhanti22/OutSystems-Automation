package main

import "github.com/parth/os-agent/cmd"

// version is injected at build time via -ldflags.
var version = "dev"

func main() {
	cmd.SetVersion(version)
	cmd.Execute()
}
