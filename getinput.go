package main

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

var commands = []string{"exit", "echo"}

func HandleCompletion(prefix string, idx int) (string, int) {
	matches := []string{}
	for _, c := range commands {
		if strings.HasPrefix(c, prefix) {
			matches = append(matches, c)
		}
	}
	if len(matches) == 0 {
		return prefix, 0
	}
	return matches[idx%len(matches)], idx + 1
}

func ReadInput() (string, error) {
	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return "", err
	}
	defer term.Restore(fd, oldState)

	buf := make([]byte, 1)
	cmd := ""
	completionIndex := 0
	prefix := ""

	for {
		_, err := os.Stdin.Read(buf)
		if err != nil {
			return "", err
		}
		if buf[0] == 9 { // Tab key
			if prefix == "" {
				prefix = cmd
			}
			cmd, completionIndex = HandleCompletion(prefix, completionIndex)
		} else if buf[0] == 127 { // Backspace key
			if len(cmd) > 0 {
				cmd = cmd[:len(cmd)-1]
				completionIndex = 0
				prefix = cmd
			}
		} else if buf[0] == 13 { // Enter key
			fmt.Println("\r")
			return cmd, nil
		} else {
			cmd += string(buf[0])
			completionIndex = 0
			prefix = cmd
		}

		// Clear and redraw line
		fmt.Print("\x1b[2K\r$ " + cmd)
	}
}
