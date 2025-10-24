package main

import "github.com/octoberswimmer/aer/cmd"

// version is overridden at link time via -ldflags.
var version = "dev"

func init() {
	if version != "" {
		cmd.RootCmd.Version = version
		cmd.RootCmd.SetVersionTemplate("aer version {{.Version}}\n")
	}
}

func main() {
	cmd.Execute()
}
