package table

import (
	"fmt"
	"github.com/targodan/golib/math"
	"reflect"
	"strings"

	"github.com/dustin/go-humanize"
)

func RegisterHumanizeFormatters(t *T, siDigits int) *T {
	t.WithValueFormatter(Bytes(0), FormatterBytes1000(siDigits))
	registerHumanizeFormatters(t, siDigits)
	return t
}

func registerHumanizeFormatters(t *T, siDigits int) {
	t.WithValueFormatter(Percent(0), FormatterPercent(siDigits))
	t.WithValueFormatter(int(0), FormatterSI(siDigits))
	t.WithValueFormatter(int8(0), FormatterSI(siDigits))
	t.WithValueFormatter(int16(0), FormatterSI(siDigits))
	t.WithValueFormatter(int32(0), FormatterSI(siDigits))
	t.WithValueFormatter(int64(0), FormatterSI(siDigits))
	t.WithValueFormatter(uint(0), FormatterSI(siDigits))
	t.WithValueFormatter(uint8(0), FormatterSI(siDigits))
	t.WithValueFormatter(uint16(0), FormatterSI(siDigits))
	t.WithValueFormatter(uint32(0), FormatterSI(siDigits))
	t.WithValueFormatter(uint64(0), FormatterSI(siDigits))
	t.WithValueFormatter(float32(0), FormatterSI(siDigits))
	t.WithValueFormatter(float64(0), FormatterSI(siDigits))
}

type Bytes uint64

func FormatterBytes1000(siDigits int) ValueFormatter {
	return func(v any) (string, error) {
		b, ok := v.(Bytes)
		if !ok {
			return "", fmt.Errorf("expected type '%v', got type '%v'", reflect.TypeOf(Bytes(0)), reflect.TypeOf(v))
		}
		return FormatSI(uint64(b), siDigits, "B")
	}
}

type Percent float64

func FormatterPercent(siDigits int) ValueFormatter {
	return func(v any) (string, error) {
		p, ok := v.(Percent)
		if !ok {
			return "", fmt.Errorf("expected type '%v', got type '%v'", reflect.TypeOf(Percent(0)), reflect.TypeOf(v))
		}
		return fmt.Sprintf(fmt.Sprintf("%%6.%df %%%%", siDigits), p), nil
	}
}

func FormatterSI(siDigits int) ValueFormatter {
	return func(v any) (string, error) {
		return FormatSI(v, siDigits, "")
	}
}

func FormatSI(v any, siDigits int, unit string) (string, error) {
	switch v.(type) {
	case int:
		return formatSI(v.(int), siDigits, unit)
	case int8:
		return formatSI(v.(int8), siDigits, unit)
	case int16:
		return formatSI(v.(int16), siDigits, unit)
	case int32:
		return formatSI(v.(int32), siDigits, unit)
	case int64:
		return formatSI(v.(int64), siDigits, unit)
	case uint:
		return formatSI(v.(uint), siDigits, unit)
	case uint8:
		return formatSI(v.(uint8), siDigits, unit)
	case uint16:
		return formatSI(v.(uint16), siDigits, unit)
	case uint32:
		return formatSI(v.(uint32), siDigits, unit)
	case uint64:
		return formatSI(v.(uint64), siDigits, unit)
	case float32:
		return formatSI(v.(float32), siDigits, unit)
	case float64:
		return formatSI(v.(float64), siDigits, unit)
	default:
		return "", fmt.Errorf("invalid type %v for FormatSI", reflect.TypeOf(v))
	}
}

func formatSI[T Number](v T, siDigits int, unit string) (string, error) {
	return strings.TrimSpace(humanize.SIWithDigits(float64(v), siDigits, unit)), nil
}

type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64
}

func unifyUnitLength(s string) string {
	parts := strings.Split(s, " ")
	return fmt.Sprintf("%s %-3s", parts[0], parts[1])
}

func formatPercent(percent float64) string {
	return fmt.Sprintf("%5.2f %%", percent)
}

func DefaultValueFormatter(v any) (string, error) {
	return fmt.Sprint(v), nil
}

func NoopRowFormatter(row int, cells []string, maxObservedWidths []int) []string {
	return cells
}

func AlignRight(col int, cells []string, maxObservedWidth int) []string {
	for i, v := range cells {
		cells[i] = formatConstWidth(v, maxObservedWidth)
	}
	return cells
}

func AlignLeft(col int, cells []string, maxObservedWidth int) []string {
	for i, v := range cells {
		cells[i] = formatConstWidth(v, -maxObservedWidth)
	}
	return cells
}

func AlignCenterFormatter(t *T) ColumnFormatter {
	return func(col int, cells []string, maxObservedWidth int) []string {
		return AlignCenter(t, col, cells, maxObservedWidth)
	}
}

func AlignCenter(t *T, col int, cells []string, maxObservedWidth int) []string {
	for i, v := range cells {
		width := t.WidthFunc(v)
		paddingSum := maxObservedWidth - width
		paddingRight := paddingSum / 2
		paddingLeft := paddingSum - paddingRight

		sb := &strings.Builder{}
		sb.Grow(maxObservedWidth)
		stringBuilderAddN(sb, " ", paddingLeft)
		sb.WriteString(v)
		stringBuilderAddN(sb, " ", paddingRight)

		cells[i] = sb.String()
	}
	return cells
}

func AlignAtDotFormatter(t *T, firstRowFormatter ColumnFormatter) ColumnFormatter {
	return func(col int, cells []string, maxObservedWidth int) []string {
		return AlignAtDot(t, firstRowFormatter, col, cells, maxObservedWidth)
	}
}

func AlignAtDot(t *T, firstRowFormatter ColumnFormatter, col int, cells []string, maxObservedWidth int) []string {
	var maxLeftWidth, maxRightWidth int

	parts := make([][]string, len(cells))
	for i, v := range cells {
		if i == 0 && firstRowFormatter != nil {
			continue
		}

		parts[i] = strings.Split(v, ".")
		if len(parts[i]) == 2 {
			maxLeftWidth = math.Max(maxLeftWidth, t.WidthFunc(parts[i][0]))
			maxRightWidth = math.Max(maxRightWidth, t.WidthFunc(parts[i][1]))
		}
		if len(parts[i]) == 1 {
			maxLeftWidth = math.Max(maxLeftWidth, t.WidthFunc(parts[i][0]))
		}
	}

	maxObservedWidth = math.Max(maxObservedWidth, maxLeftWidth+1+maxRightWidth)

	if firstRowFormatter != nil {
		cells[0] = firstRowFormatter(col, []string{cells[0]}, maxObservedWidth)[0]
		maxObservedWidth = math.Max(maxObservedWidth, t.WidthFunc(cells[0]))
	}

	extraPadding := math.Max(0, maxObservedWidth-maxLeftWidth-1-maxRightWidth)
	maxLeftWidth += extraPadding

	for i, p := range parts {
		if i == 0 && firstRowFormatter != nil {
			continue
		}

		if len(p) > 2 {
			cells[i] = fmt.Sprintf("<invalid number of dots: '%s'>", cells[i])
		}
		if len(p) == 2 {
			cells[i] = fmt.Sprintf(fmt.Sprintf("%%%ds.%%-%ds", maxLeftWidth, maxRightWidth), p[0], p[1])
		}
		if len(p) == 1 {
			cells[i] = fmt.Sprintf(fmt.Sprintf("%%%ds %%-%ds", maxLeftWidth, maxRightWidth), p[0], "")
		}
	}
	return cells
}

func stringBuilderAddN(sb *strings.Builder, s string, times int) {
	for i := 0; i < times; i++ {
		sb.WriteString(s)
	}
}

func formatConstWidth(s string, width int) string {
	return fmt.Sprintf(fmt.Sprintf("%%%ds", width), s)
}
