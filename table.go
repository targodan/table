package table

import (
	"fmt"
	"github.com/targodan/golib/fpExtra"
	"reflect"
	"strings"
	"unicode/utf8"

	"github.com/targodan/golib/math"

	"github.com/targodan/table/border"
	"github.com/targodan/table/util"

	"github.com/repeale/fp-go"
	"github.com/targodan/golib/fpExtra/pair"
)

var (
	DefaultWidthFunc       util.WidthFunc  = utf8.RuneCountInString
	DefaultColumnFormatter ColumnFormatter = AlignRight
	DefaultBorder                          = border.Internal{ColumnSeparator: " "}
)

// ValueFormatter should format arbitrary values to strings.
type ValueFormatter func(v any) (string, error)

// RowFormatter receives the row-index, all cells of the row as formatted by the configured ValueFormatter and the maximum
// width of each column. It should return the formatted cells. The input slice may be modified or copied. The width of each cell
// may also be changed.
type RowFormatter func(row int, cells []string, maxObservedWidths []int) []string

// ColumnFormatter receives the column-index, all cells of the column as formatted by the configured ValueFormatter and the maximum
// width of this column. It should return the formatted cells. The width of the cells may be changed, but all cells must have the same
// width.
type ColumnFormatter func(col int, cells []string, maxObservedWidth int) []string

type Border interface {
	Horizontal(row int, numRows int, columnWidths []int, widthFunc util.WidthFunc) string
	Vertical(row, col int, numRows, numCols int, widthFunc util.WidthFunc) string
}

type T struct {
	Header       []string
	Rows         [][]any
	Border       Border
	RowSeparator string
	WidthFunc    util.WidthFunc

	valueFormatters  map[string]ValueFormatter
	headerFormatter  RowFormatter
	columnFormatters []ColumnFormatter
}

func New() *T {
	return &T{
		Header:           nil,
		Rows:             nil,
		Border:           DefaultBorder,
		RowSeparator:     "\n",
		WidthFunc:        DefaultWidthFunc,
		valueFormatters:  make(map[string]ValueFormatter, 0),
		headerFormatter:  NoopRowFormatter,
		columnFormatters: make([]ColumnFormatter, 0),
	}
}

func (t *T) WithValueFormatter(typ any, f ValueFormatter) *T {
	t.valueFormatters[reflect.TypeOf(typ).String()] = f
	return t
}

func (t *T) WithHeaderFormatter(f RowFormatter) *T {
	t.headerFormatter = f
	return t
}

func (t *T) WithFirstColumnFormatter(f ColumnFormatter) *T {
	return t.WithColumnFormatter(0, f)
}

func (t *T) WithColumnFormatter(col int, f ColumnFormatter) *T {
	if col >= len(t.columnFormatters) {
		newFormatters := make([]ColumnFormatter, col+1)
		copy(newFormatters, t.columnFormatters)
		t.columnFormatters = newFormatters
	}
	t.columnFormatters[col] = f
	return t
}

func (t *T) WithColumnFormatters(f ...ColumnFormatter) *T {
	t.columnFormatters = f
	return t
}

func (t *T) WithuWidthFunction(w util.WidthFunc) *T {
	t.WidthFunc = w
	return t
}

func (t *T) AddRow(values ...any) *T {
	t.Rows = append(t.Rows, values)
	return t
}

func (t *T) Format() (string, error) {
	numCols := fp.Compose2(
		fp.Reduce(math.Max[int], len(t.Header)),
		fp.Map(fpExtra.SliceLen[any]),
	)(t.Rows)

	formatted := make([][]string, len(t.Rows)+1)
	formatted[0] = make([]string, len(t.Header))
	copy(formatted[0], t.Header)

	colWidths := make([]int, numCols)
	for i, row := range t.Rows {
		formatted[i+1] = make([]string, len(row))

		for j, val := range row {
			valS, err := t.formatValue(val)
			if err != nil {
				return "", fmt.Errorf("could not format cell %d, %d (row, col), reason: %w", i, j, err)
			}
			formatted[i+1][j] = valS
			colWidths[j] = math.Max(colWidths[j], t.WidthFunc(valS))
		}
	}

	formatted[0] = t.headerFormatter(0, formatted[0], colWidths)

	colWidths = fp.Compose3(
		// Choose maximum for each column
		fp.Map(pair.Reduce(math.Max[int])),
		// Compute new cell widths
		pair.MapRight[int, string, int](t.WidthFunc),
		// Resolve default values
		pair.Map(fpExtra.GetOrElseValue(0), fpExtra.GetOrElseValue("")),
	)(pair.ZipLongest(colWidths, formatted[0]))

	for j := 0; j < numCols; j++ {
		var formatter ColumnFormatter
		if j < len(t.columnFormatters) && t.columnFormatters[j] != nil {
			formatter = t.columnFormatters[j]
		} else {
			formatter = DefaultColumnFormatter
		}
		values := make([]string, numCols)
		for i, row := range formatted {
			if j < len(row) {
				values[i] = row[j]
			}
		}

		values = formatter(j, values, colWidths[j])
		colWidths[j] = fp.Compose2(fp.Reduce(math.Max[int], colWidths[j]), fp.Map(t.WidthFunc))(values)

		for i, row := range formatted {
			if j < len(row) {
				row[j] = values[i]
			}
		}
	}

	rowWidth := 0
	for _, w := range colWidths {
		rowWidth += w
	}

	builder := &strings.Builder{}
	builder.Grow((rowWidth + len(t.RowSeparator)) * len(t.Rows))

	builder.WriteString(t.Border.Horizontal(-1, len(t.Rows), colWidths, t.WidthFunc))
	for i, row := range formatted {
		builder.WriteString(t.Border.Vertical(i, -1, len(t.Rows), numCols, t.WidthFunc))

		for j, val := range row {
			builder.WriteString(val)
			builder.WriteString(t.Border.Vertical(i, j, len(t.Rows), len(row), t.WidthFunc))
		}

		builder.WriteString(t.RowSeparator)

		builder.WriteString(t.Border.Horizontal(i, len(t.Rows), colWidths, t.WidthFunc))
	}
	return builder.String(), nil
}

func (t *T) String() string {
	out, err := t.Format()
	if err != nil {
		return fmt.Sprintf("<error: %v>", err)
	}
	return out
}

func (t *T) formatValue(val any) (string, error) {
	typ := reflect.TypeOf(val)
	f, ok := t.valueFormatters[typ.String()]
	if !ok {
		return DefaultValueFormatter(val)
	}
	return f(val)
}
