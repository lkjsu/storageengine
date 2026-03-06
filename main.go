package main

import (
	"fmt"
    "bufio"
	"os"
	"strings"
)

func main() {
    scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("Welcome to StorageEngine\n")
	for {
		fmt.Print("> ")
        scanner.Scan()
		input := strings.TrimSpace(scanner.Text())
        if input == ".exit" {
		    fmt.Println("Exiting...")
		    break
	    }
		fmt.Println("Unrecognized command:", input)
    }
}
