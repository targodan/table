package table

import (
	"fmt"
	"testing"
)

func TestTable(t *testing.T) {
	tab := New()
	tab.Header = []string{"A", "B", "C"}
	tab.AddRow(14, 2, 3)
	tab.AddRow(4, 5, 622)

	fmt.Println(tab)
}
