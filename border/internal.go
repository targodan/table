package border

import (
	"strings"

	"github.com/targodan/table/util"
)

const (
	// Taken from https://en.wikipedia.org/wiki/Box-drawing_character

	FullWidthHorizontalLine = "─"
	FullHeightVerticalLine  = "│"
	FullCross               = "┼"
)

type Internal struct {
	ColumnSeparator string
	RowSeparator    string
	CrossingSymbol  string
}

func (b Internal) Horizontal(row int, numRows int, columnWidths []int, widthFunc util.WidthFunc) string {
	if len(b.RowSeparator) > 0 && row >= 0 && row < numRows {
		if b.CrossingSymbol == "" {
			b.CrossingSymbol = b.RowSeparator
		}

		sb := strings.Builder{}
		for c, colWidth := range columnWidths {
			for i := 0; i < colWidth/widthFunc(b.RowSeparator); i++ {
				sb.WriteString(b.RowSeparator)
			}
			sb.WriteString(b.RowSeparator[:colWidth%widthFunc(b.RowSeparator)])
			if c < len(columnWidths)-1 {
				for i := 0; i < widthFunc(b.ColumnSeparator)/widthFunc(b.CrossingSymbol); i++ {
					sb.WriteString(b.CrossingSymbol)
				}
				sb.WriteString(b.CrossingSymbol[:widthFunc(b.ColumnSeparator)%widthFunc(b.CrossingSymbol)])
			}
		}
		sb.WriteString("\n")

		return sb.String()
	}

	return ""
}

func (b Internal) Vertical(row, col int, numRows, numCols int, widthFunc util.WidthFunc) string {
	if col >= 0 && col < numCols-1 {
		return b.ColumnSeparator
	}
	return ""
}
