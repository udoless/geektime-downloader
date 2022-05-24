package utils

import (
	"os"
	"testing"
)

func TestPrintToPdf(t *testing.T) {
	filename := "file.pdf"
	err := ColumnPrintToPDF(158248, filename, nil)

	if err != nil {
		t.Fatal("PrintToPDF test is failure", err)
	} else {
		os.Remove(filename)
	}
}
