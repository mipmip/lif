package main

import "github.com/xRiErOS/lif/internal/commands"

func main() {
	root := commands.NewRootCmd()
	commands.RegisterCoreCommands(root)
	commands.Execute(root)
}
