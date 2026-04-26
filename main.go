package main

import (
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"strings"
)

func main() {
	var filePath string
	flag.StringVar(&filePath, "i", "", "Input CSV file")
	flag.Parse()

	if filePath == "" {
		log.Fatal("Input CSV file is required. Use -i flag to specify the file path.")
	}

	// check the file exists and is a CSV file
	if !isCSVFile(filePath) {
		log.Fatal("Input file must be a CSV file.")
	}

	// Read the CSV file and process it
	records, err := readCsvFile(filePath)

	if err != nil {
		log.Fatal(err)
	}

	// Process the records as needed
	if err := processRecords(records); err != nil {
		log.Fatal(err)
	}
}

// isCSVFile checks if the given file path exists and has a .csv extension
func isCSVFile(filePath string) bool {
	if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
		return false
	}

	// Check if the file has a .csv extension
	return path.Ext(filePath) == ".csv"
}

// readCsvFile reads the CSV file and returns the records as a slice of string slices
func readCsvFile(filePath string) ([][]string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, err
	}

	return records, nil
}

// processRecords takes the CSV records and prints them in Markdown table format
func processRecords(records [][]string) error {
	if len(records) == 0 {
		return fmt.Errorf("CSV file is empty")
	}

	headers := records[0]
	if len(headers) == 0 {
		return fmt.Errorf("CSV file has no headers")
	}

	escapeCell := func(value string) string {
		value = strings.ReplaceAll(value, "|", "\\|")
		value = strings.ReplaceAll(value, "\n", "<br>")
		return value
	}

	// Header row
	fmt.Print("|")
	for _, header := range headers {
		fmt.Printf(" %s |", escapeCell(header))
	}
	fmt.Println()

	// Markdown separator row
	fmt.Print("|")
	for range headers {
		fmt.Print(" --- |")
	}
	fmt.Println()

	// Data rows
	for _, row := range records[1:] {
		fmt.Print("|")
		for i := range headers {
			cell := ""
			if i < len(row) {
				cell = escapeCell(row[i])
			}
			fmt.Printf(" %s |", cell)
		}
		fmt.Println()
	}
	return nil
}
