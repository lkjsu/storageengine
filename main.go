package main

import "fmt"

func main() {
    var input string
	fmt.Print("Welcome to StorageEngine\n")
	for {
		fmt.Print("> ")
        fmt.Scanf("%s", &input)
        if input == ".exit" {
		    fmt.Println("Exiting...")
		    break
	    }
    }
}
