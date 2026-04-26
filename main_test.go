package main

import (
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	originalStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create stdout pipe: %v", err)
	}

	os.Stdout = w

	fn()

	if err := w.Close(); err != nil {
		t.Fatalf("failed to close writer: %v", err)
	}
	os.Stdout = originalStdout

	output, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("failed to read captured stdout: %v", err)
	}

	return string(output)
}

func writeTempCSV(t *testing.T, name string, content string) string {
	t.Helper()

	filePath := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(filePath, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write temp csv file: %v", err)
	}

	return filePath
}

func TestReadCsvFile_ParsesBasicRows(t *testing.T) {
	filePath := writeTempCSV(t, "basic.csv", "name,age\nAlice,30\nBob,40\n")

	records, err := readCsvFile(filePath)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	want := [][]string{{"name", "age"}, {"Alice", "30"}, {"Bob", "40"}}
	if !reflect.DeepEqual(records, want) {
		t.Fatalf("records mismatch\nwant: %#v\ngot:  %#v", want, records)
	}
}

func TestReadCsvFile_ParsesQuotedComma(t *testing.T) {
	filePath := writeTempCSV(t, "quoted-comma.csv", "name,city\n\"Doe, John\",NY\n")

	records, err := readCsvFile(filePath)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	want := [][]string{{"name", "city"}, {"Doe, John", "NY"}}
	if !reflect.DeepEqual(records, want) {
		t.Fatalf("records mismatch\nwant: %#v\ngot:  %#v", want, records)
	}
}

func TestReadCsvFile_ParsesQuotedNewline(t *testing.T) {
	filePath := writeTempCSV(t, "quoted-newline.csv", "id,notes\n1,\"line 1\nline 2\"\n")

	records, err := readCsvFile(filePath)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	want := [][]string{{"id", "notes"}, {"1", "line 1\nline 2"}}
	if !reflect.DeepEqual(records, want) {
		t.Fatalf("records mismatch\nwant: %#v\ngot:  %#v", want, records)
	}
}

func TestReadCsvFile_ErrorsOnMalformedQuotes(t *testing.T) {
	filePath := writeTempCSV(t, "bad-quotes.csv", "name,age\n\"Alice,30\n")

	_, err := readCsvFile(filePath)
	if err == nil {
		t.Fatal("expected an error for malformed quoted field, got nil")
	}
}

func TestReadCsvFile_ErrorsOnInconsistentFieldCount(t *testing.T) {
	filePath := writeTempCSV(t, "inconsistent-fields.csv", "a,b\n1,2,3\n")

	_, err := readCsvFile(filePath)
	if err == nil {
		t.Fatal("expected an error for inconsistent field count, got nil")
	}
}

func TestReadCsvFile_EmptyFileReturnsNoRecords(t *testing.T) {
	filePath := writeTempCSV(t, "empty.csv", "")

	records, err := readCsvFile(filePath)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(records) != 0 {
		t.Fatalf("expected no records, got: %d", len(records))
	}
}

func TestReadCsvFile_ErrorsForMissingFile(t *testing.T) {
	_, err := readCsvFile(filepath.Join(t.TempDir(), "missing.csv"))
	if err == nil {
		t.Fatal("expected an error for missing file, got nil")
	}
}

func TestProcessRecords_PrintsMarkdownTable(t *testing.T) {
	records := [][]string{
		{"name", "age"},
		{"Alice", "30"},
		{"Bob", "40"},
	}

	output := captureStdout(t, func() {
		err := processRecords(records)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
	})

	want := "| name | age |\n| --- | --- |\n| Alice | 30 |\n| Bob | 40 |\n"
	if output != want {
		t.Fatalf("markdown output mismatch\nwant:\n%s\ngot:\n%s", want, output)
	}
}

func TestProcessRecords_EscapesPipesAndNewlines(t *testing.T) {
	records := [][]string{
		{"title", "description"},
		{"a|b", "line1\nline2"},
	}

	output := captureStdout(t, func() {
		err := processRecords(records)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
	})

	want := "| title | description |\n| --- | --- |\n| a\\|b | line1<br>line2 |\n"
	if output != want {
		t.Fatalf("markdown output mismatch\nwant:\n%s\ngot:\n%s", want, output)
	}
}

func TestProcessRecords_FillsMissingCellsInShortRows(t *testing.T) {
	records := [][]string{
		{"name", "age", "city"},
		{"Alice", "30"},
	}

	output := captureStdout(t, func() {
		err := processRecords(records)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
	})

	want := "| name | age | city |\n| --- | --- | --- |\n| Alice | 30 |  |\n"
	if output != want {
		t.Fatalf("markdown output mismatch\nwant:\n%s\ngot:\n%s", want, output)
	}
}

func TestProcessRecords_ErrorsOnEmptyRecords(t *testing.T) {
	err := processRecords(nil)
	if err == nil {
		t.Fatal("expected an error for empty records, got nil")
	}
}

func TestProcessRecords_ErrorsOnEmptyHeaderRow(t *testing.T) {
	err := processRecords([][]string{{}})
	if err == nil {
		t.Fatal("expected an error for empty headers, got nil")
	}
}
