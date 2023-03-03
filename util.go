package simplesysinfo

import (
	"fmt"
	"strconv"
	"strings"
)

func getIncludeAll() []includedItem {
	var includes []includedItem
	for i := 0; i < int(last_iota); i++ {
		includes = append(includes, includedItem(i))
	}
	return includes
}

func writeToBufIfVerbose(buf *strings.Builder, name string, value interface{}, nullValue ...string) {
	if VERBOSE {
		writeToBuf(buf, name, value, nullValue...)
	}
}

func writeToBuf(buf *strings.Builder, name string, value interface{}, nullValue ...string) {
	var stringVal string
	switch v := value.(type) {
	case string:
		stringVal = v
	case int:
		stringVal = strconv.Itoa(v)
	case int8:
		stringVal = strconv.Itoa(int(v))
	case int16:
		stringVal = strconv.Itoa(int(v))
	case int32:
		stringVal = strconv.Itoa(int(v))
	case int64:
		stringVal = strconv.Itoa(int(v))
	case float32:
		stringVal = strconv.FormatFloat(float64(v), 'f', 2, 64)
	case float64:
		stringVal = strconv.FormatFloat(v, 'f', 2, 64)
	case bool:
		stringVal = strconv.FormatBool(v)
	case fmt.Stringer:
		stringVal = v.String()
	default:
		stringVal = fmt.Sprintf("%v", v)
	}
	buf.Grow(len(name) + len(stringVal) + len("\t: \n"))
	buf.WriteString("\t")
	buf.WriteString(name)
	buf.WriteString(": ")
	stringVal = strings.TrimSpace(stringVal)
	if stringVal == "" || stringVal == "0" || stringVal == "[]" || stringVal == "{}" || stringVal == "()" {
		if len(nullValue) > 0 {
			buf.WriteString(nullValue[0])
		} else {
			buf.WriteString("No ")
			buf.WriteString(name)
			buf.WriteString(" detected")
		}
	} else {
		buf.WriteString(stringVal)
	}
	buf.WriteString("\n")
}

func Contains[T comparable](slice []T, item T) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func ByteToGB(b uint64) string {
	return strconv.FormatFloat(float64(b)/1024/1024/1024, 'f', 2, 64) + " GB"
}
