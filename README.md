# Sheets-orm for Golang

This is a very-simple orm-like library for Google Sheets, that we use in our internal administration tasks.

This project is currently under heavy development.

## Example

```go
package main

import (
	"fmt"

	"github.com/pproj/sheetsorm"
)

type Record struct {
	Name  string `sheet:"A,uid"`                // The "name" field of the record is considered the UID here, record lookups will be based on this column.
	Age   int    `sheet:"B"`
	Happy *bool  `sheet:"C,True=yes,False=no"`
}

func main() {

	// Create new Sheet instance

	cfg := sheets.StructureConfig{
		DocID:    "", // Google sheets ID
		Sheet:    "", // Empty string means the default sheet here
		SkipRows: 1,  // The first row is a header, so we should skip it
	}
	logger, _ := zap.NewProduction()
	sheet := sheetsorm.NewSheet("path/To/Credentials", cfg, logger)

	// Load a record

	var recordToLoad Record
	recordToLoad.Name = "Bob"

	sheet.GetRecord(ctx.Background(), &recordToLoad)

	fmt.Println(recordToLoad)

	// Update a record

	var recordToUpdate Record
	recordToUpdate.Name = "Alice"
	recordToUpdate.Age = 22

	sheet.UpdateRecords(ctx.Background(), &recordToUpdate) // Happy is not updated, because it was nil

	fmt.Println(recordToUpdate) // The updated record is read back entirely... so the Happy field will be filled here
}
```