package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type ParserFunc func(sh *Shell, command, param string) error

type Shell struct {
	path string // current context, e.g. "dc3/app3"
}

func NewShell() *Shell {
	return &Shell{
		path: "root",
	}
}

func (sh *Shell) prompt() {
	fmt.Printf("%s> ", sh.path)
}

func (sh *Shell) Run(parser ParserFunc) {
	reader := bufio.NewReader(os.Stdin)

	for {
		sh.prompt()

		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Err:", err)
			return
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		cmd, params := sh.handleCommand(line)
		err = parser(sh, cmd, params)
		if err != nil {
			fmt.Println("Error:\n", err)
		}
	}
}

func (sh *Shell) handleCommand(cmd string) (string, string) {
	parts := strings.SplitN(cmd, " ", 2)
	command := parts[0]

	var args string
	if len(parts) > 1 {
		args = parts[1]
	}

	return command, args
}

func (sh *Shell) Goto(arg string) {
	if arg == "" {
		fmt.Println("Usage: goto <path>")
		return
	}
	sh.path = arg
}
