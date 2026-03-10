package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Row struct {
	id       int64
	username [32]byte
	email    [255]byte
}

type Table struct {
	numRows int
	pages    [100][]byte
}

/* Write a function that can distinguish between meta commands and SQL commands */
func isMetaCommand(input string) bool {
	return strings.HasPrefix(input, ".")
}


func processSelectCommand(input string) {	// This is a very basic parser for the SELECT command. It does not handle all cases and is just for demonstration.
	// It assumes the format: SELECT column1, column2 table_name WHERE condition
	fmt.Printf("Parsed SELECT command: \n")
}

func processInsertCommand(input string) []byte {
	// This is a very basic parser for the INSERT command. Assumes a single table
	// It assumes the format: INSERT value1, value2, ...
	parts := strings.SplitN(input, " ", 2)
	if len(parts) < 2 {
		fmt.Println("Invalid INSERT command format.")
		return []byte{}
	}
	valuesPart := parts[1]
	values := strings.Split(valuesPart, ",")
	for i, value := range values {
		values[i] = strings.TrimSpace(value)
	}

	/* Values form the Schema for Rows with
	   int, varchar(32), varchar(255) for id
	   username and email.
	*/
	row := Row{}
	if len(values) >= 1 {
		id, err := strconv.ParseInt(values[0], 10, 64)
		if err != nil {
			panic(err)
		}
		row.id = id
	}
	if len(values) >= 2 {
		copy(row.username[:], values[1])
	}
	if len(values) >= 3 {
		copy(row.email[:], values[2])
	}

	pageSize := 4096
	// fmt.Printf("Calculated rows per page: %d\n", rowsPerPage)

	buffer := make([]byte, pageSize)
	// Here we would need to serialize the row into the buffer. This is a simple example and does not handle all edge cases.
	copy(buffer[0:8], []byte(strconv.FormatInt(row.id, 10)))
	copy(buffer[8:40], row.username[:])
	copy(buffer[40:295], row.email[:])
	// Store this row into the table
	// Before that I think encoding will be necessary.

	fmt.Printf("Parsed INSERT command with values: %v\n", values)
	return buffer
}

func rowPosition(table *Table, rowNum int) []byte {
	pageSize := 4096
	rowSize := 8 + 32 + 255 // Size of id + username + email
	rowsPerPage := pageSize / rowSize
	pageNum := rowNum / rowsPerPage
	byteOffset := (rowNum % rowsPerPage) * rowSize
	if table.pages[pageNum] == nil {
		table.pages[pageNum] = make([]byte, pageSize)
	}
	return table.pages[pageNum][byteOffset:byteOffset+rowSize] // This is the position of the row in the table
}

/* This is the main entry point for the StorageEngine.
   The goal now is to be able to properly de-markate the meta commands
   and everything else will be SQL command.
*/
func main() {
	var table Table
    scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("Welcome to StorageEngine\n")
	for {
		fmt.Print("> ")
        scanner.Scan()
		input := strings.TrimSpace(scanner.Text())
        if isMetaCommand(input) {
			if input == ".exit" {
				fmt.Println("Exiting StorageEngine. Goodbye!")
				break
			} else {
				fmt.Println("Unrecognized meta command:", input)
			}
		} else {
			if input != "" {
				// Check if the input contains valid SQL commands SELECT, INSERT for now, upper or lower case.
				if strings.HasPrefix(strings.ToUpper(input), "SELECT") {
					// Start processing the SELECT command using a simple parser.
					processSelectCommand(input)
				} else if strings.HasPrefix(strings.ToUpper(input), "INSERT") {
					// Start processing the INSERT command function.
					row := processInsertCommand(input)
					if len(row) == 0 {
						fmt.Println("Failed to parse INSERT command.")
						continue
					}
					slot := rowPosition(&table, table.numRows)
					copy(slot, row)
					table.numRows++
				} else {
					fmt.Println("Unrecognized SQL command:", input)
				}
			}
		}
    }
}
