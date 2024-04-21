# Sheets-orm for Golang

[![Build Status](https://drone.k8s.marcsello.com/api/badges/pproj/sheetsorm/status.svg)](https://drone.k8s.marcsello.com/pproj/sheetsorm)

This is a very-simple orm-like library for Google Sheets, that we use in our internal administration tasks.

This project is currently under heavy development.

## Example

```go
package main

import (
	"context"
	"fmt"

	"github.com/pproj/sheetsorm"
)

type Record struct {
	Name  string `sheet:"A,uid"` // The "name" field of the record is considered the UID here, record lookups will be based on this column.
	Age   int    `sheet:"B"`
	Happy *bool  `sheet:"C,True=yes,False=no"`
}

func main() {

	// Create new Sheet instance

	cfg := sheetsorm.StructureConfig{
		DocID:    "", // Google sheets ID
		Sheet:    "", // Empty string means the default sheet here
		SkipRows: 1,  // The first row is a header, so we should skip it
	}

	srv, err := sheets.NewService(context.Background(), option.WithCredentialsFile("path/To/Credentials"))
	if err != nil {
		panic(err)
	}

	sheet := sheetsorm.NewSheet(srv, cfg)

	// Load a record

	var recordToLoad Record
	recordToLoad.Name = "Bob"

	sheet.GetRecord(context.Background(), &recordToLoad)

	fmt.Println(recordToLoad)

	// Update a record

	var recordToUpdate Record
	recordToUpdate.Name = "Alice"
	recordToUpdate.Age = 22

	sheet.UpdateRecords(context.Background(), &recordToUpdate) // Happy is not updated, because it was nil

	fmt.Println(recordToUpdate) // The updated record is read back entirely... so the Happy field will be filled here
}
```