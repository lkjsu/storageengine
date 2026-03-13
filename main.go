package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
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
	pages   [100][]byte
}

/* Write a function that can distinguish between meta commands and SQL commands */
func isMetaCommand(input string) bool {
	return strings.HasPrefix(input, ".")
}

func processSelectCommand(input string, table *Table) {
	// This is a very basic parser for the SELECT command. It does not handle all cases and is just for demonstration.
	// It assumes the format: SELECT column1, column2 table_name WHERE condition
	// fmt.Printf("SELECT part: %s\n", selectPart)
	fmt.Printf("Table rows %d\n", table.numRows)
	for i := 0; i < table.numRows; i++ {
		slot := rowPosition(table, i)
		id := string(slot[0:8])
		username := string(slot[8:40])
		email := string(slot[40:295])
		fmt.Printf("Row %d: id=%s, username=%s, email=%s\n", i+1, id, username, email)
	}
	fmt.Errorf("Parsed SELECT command \n")
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

func saveToFile(table *Table) {
	// We need to determine which pages need to be pasted to file.
	pageSize := 4096
	rowSize := 8 + 32 + 255
	rowsPerPage := pageSize / rowSize
	pages := (table.numRows + rowsPerPage - 1) / rowsPerPage
	file, err := os.Create("file.db") // For read access.
	if err != nil {
		fmt.Print(err)
	}
	defer file.Close()

	header := new(bytes.Buffer)
	binary.Write(header, binary.BigEndian, int64(table.numRows))
	file.Write(header.Bytes())
	for i := 0; i < pages; i++ {
		if table.pages[i] != nil {
			file.Write(table.pages[i])
		}

	}
}

func loadTableFromFile(table *Table) {
	file, err := os.Open("file.db")
	if err != nil {
		fmt.Print(err)
	}

	defer file.Close()

	pageNumber := 0
	header := make([]byte, 8)
	n, err := io.ReadFull(file, header)
	if err != nil {
		fmt.Errorf("Could not read header\n")
	}
	fmt.Errorf("Read %d bytes from header\n", n)
	var rows int64
	if n > 0 {
		buf := bytes.NewReader(header)
		binary.Read(buf, binary.BigEndian, &rows)
	} else {
		fmt.Print("No database initialized!\n")
		return
	}

	// fmt.Errorf(header, table.numRows)

	table.numRows = int(rows)
	for {
		buffer := make([]byte, 4096)
		n, err := file.Read(buffer)
		if n > 0 {
			table.pages[pageNumber] = make([]byte, 4096)
			copy(table.pages[pageNumber][:], buffer[:n])
			pageNumber++
		}

		if err == io.EOF {
			break
		}
		fmt.Printf("Number Bytes read: %d\n", n)
	}

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
	return table.pages[pageNum][byteOffset : byteOffset+rowSize] // This is the position of the row in the table
}

/*
This is the main entry point for the StorageEngine.

	The goal now is to be able to properly de-markate the meta commands
	and everything else will be SQL command.
*/
func main() {
	var table Table
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("Welcome to StorageEngine\n")
	loadTableFromFile(&table)
	for {
		fmt.Print("> ")
		scanner.Scan()
		input := strings.TrimSpace(scanner.Text())
		if isMetaCommand(input) {
			if input == ".exit" {
				saveToFile(&table)
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
					processSelectCommand(input, &table)
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
