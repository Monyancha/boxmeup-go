package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

var hooks = []string{
	"imagery",
}

func removeBuildIgnore(content []byte) []byte {
	rest := strings.Split(string(content), "\n")
	return []byte(strings.Join(rest[2:], "\n"))
}

func main() {
	fmt.Fprintln(os.Stderr, "Importing proprietary hooks...")
	path := os.Args[1]
	for _, hook := range hooks {
		var content []byte
		content, err := ioutil.ReadFile(fmt.Sprintf("%s/vendor/github.com/cjsaylor/boxmeup-hooks/%s.go", path, hook))
		if err != nil {
			content, err = ioutil.ReadFile(fmt.Sprintf("%s/hooks/placeholders/%s.go", path, hook))
			content = removeBuildIgnore(content)
		}
		err = ioutil.WriteFile(fmt.Sprintf("%s/hooks/%s.go", path, hook), content, 0644)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		fmt.Fprintf(os.Stderr, "- %s\n", hook)
	}
}
