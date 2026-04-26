# csv-to-md-table

`csv-to-md-table` is a small CLI tool that reads a CSV file and prints its contents as a Markdown table.

## Build

```bash
go build -o csv-to-md main.go
```

## Run

```bash
./csv-to-md -i test.csv
```

## Notes

- The `-i` flag is required and must point to a `.csv` file.
- The first CSV row is used as the Markdown table header.
