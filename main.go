package main

import (
	"bufio"
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

const (
	pageSize = 4096
)

func (t *Table) rowSize() int {
	return 8 + 32 + 255
}

func (t *Table) rowsPerPage() int {
	return pageSize / t.rowSize()
}

func (t *Table) rowPosition(rowNum int) []byte {
	rSize := t.rowSize()
	rowsPer := t.rowsPerPage()
	pageNum := rowNum / rowsPer
	byteOffset := (rowNum % rowsPer) * rSize
	if t.pages[pageNum] == nil {
		t.pages[pageNum] = make([]byte, pageSize)
	}
	return t.pages[pageNum][byteOffset : byteOffset+rSize]
}

func serializeRow(r *Row) []byte {
	buf := make([]byte, pageSize) // only need rowSize but reuse constant
	binary.BigEndian.PutUint64(buf[0:8], uint64(r.id))
	copy(buf[8:40], r.username[:])
	copy(buf[40:295], r.email[:])
	return buf[:rSizeOf(r)]
}

func deserializeRow(data []byte) Row {
	var r Row
	r.id = int64(binary.BigEndian.Uint64(data[0:8]))
	copy(r.username[:], data[8:40])
	copy(r.email[:], data[40:295])
	return r
}

func rSizeOf(_ *Row) int { return 8 + 32 + 255 }

func (t *Table) insert(r Row) {
	slot := t.rowPosition(t.numRows)
	copy(slot, serializeRow(&r))
	t.numRows++
}

func (t *Table) printAll() {
	fmt.Printf("Table rows %d\n", t.numRows)
	for i := 0; i < t.numRows; i++ {
		slot := t.rowPosition(i)
		r := deserializeRow(slot)
		fmt.Printf("Row %d: id=%d, username=%s, email=%s\n",
			i+1, r.id,
			strings.TrimRight(string(r.username[:]), "\x00"),
			strings.TrimRight(string(r.email[:]), "\x00"))
	}
}

func (t *Table) save(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := binary.Write(f, binary.BigEndian, int64(t.numRows)); err != nil {
		return err
	}
	pages := (t.numRows + t.rowsPerPage() - 1) / t.rowsPerPage()
	for i := 0; i < pages; i++ {
		if t.pages[i] != nil {
			if _, err := f.Write(t.pages[i]); err != nil {
				return err
			}
		}
	}
	return nil
}

func (t *Table) load(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	var rows int64
	if err := binary.Read(f, binary.BigEndian, &rows); err != nil {
		return err
	}
	t.numRows = int(rows)

	pageNum := 0
	for {
		buf := make([]byte, pageSize)
		n, err := f.Read(buf)
		if n > 0 {
			t.pages[pageNum] = make([]byte, pageSize)
			copy(t.pages[pageNum], buf[:n])
			pageNum++
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func isMetaCommand(input string) bool {
	return strings.HasPrefix(input, ".")
}

func parseInsert(input string) (Row, error) {
	parts := strings.SplitN(input, " ", 2)
	if len(parts) < 2 {
		return Row{}, fmt.Errorf("invalid INSERT command")
	}
	vals := strings.Split(parts[1], ",")
	for i := range vals {
		vals[i] = strings.TrimSpace(vals[i])
	}
	var r Row
	if len(vals) >= 1 {
		id, err := strconv.ParseInt(vals[0], 10, 64)
		if err != nil {
			return r, err
		}
		r.id = id
	}
	if len(vals) >= 2 {
		copy(r.username[:], vals[1])
	}
	if len(vals) >= 3 {
		copy(r.email[:], vals[2])
	}
	return r, nil
}

func main() {
	var table Table
	const dbFile = "file.db"

	fmt.Println("Welcome to StorageEngine")
	if err := table.load(dbFile); err != nil && !os.IsNotExist(err) {
		fmt.Printf("unable to load table: %v\n", err)
	}

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}
		if isMetaCommand(input) {
			switch input {
			case ".exit":
				if err := table.save(dbFile); err != nil {
					fmt.Printf("error saving: %v\n", err)
				}
				fmt.Println("Exiting StorageEngine. Goodbye!")
				return
			default:
				fmt.Println("Unrecognized meta command:", input)
			}
			continue
		}
		switch strings.ToUpper(strings.Split(input, " ")[0]) {
		case "INSERT":
			row, err := parseInsert(input)
			if err != nil {
				fmt.Println("Failed to parse INSERT:", err)
				continue
			}
			table.insert(row)
		case "SELECT":
			table.printAll()
		default:
			fmt.Println("Unrecognized SQL command:", input)
		}
	}
}
