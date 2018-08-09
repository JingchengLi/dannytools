package mystr

import (
	"fmt"
	"strconv"
	"strings"
)

func SliceToStringUint32(t []uint32, sep string) string {
	arr := make([]string, len(t))

	for i, v := range t {
		arr[i] = fmt.Sprintf("%d", v)
	}

	return strings.Join(arr, sep)
}
func SliceToStringUint64(t []uint64, sep string) string {
	arr := make([]string, len(t))

	for i, v := range t {
		arr[i] = fmt.Sprintf("%d", v)
	}

	return strings.Join(arr, sep)
}

func Uin32SliceToStringSlice(t []uint32) []string {
	arr := make([]string, len(t))

	for i, v := range t {
		arr[i] = fmt.Sprintf("%d", v)
	}

	return arr
}

func Uint8SliceToStringSlice(t []uint8) []string {
	arr := make([]string, len(t))

	for i, v := range t {
		arr[i] = fmt.Sprintf("%d", v)
	}

	return arr
}

func UintSliceToStringSlice(t []uint) []string {
	arr := make([]string, len(t))

	for i, v := range t {
		arr[i] = fmt.Sprintf("%d", v)
	}

	return arr
}

func IntSliceToStringSlice(intArr []interface{}) []string {
	arrStr := make([]string, len(intArr))
	for j, v := range intArr {
		switch i := v.(type) {
		case int64, int32, int16, int8, int, uint64, uint32, uint16, uint8, uint:
			arrStr[j] = fmt.Sprintf("%d", i)
		default:
			arrStr[j] = fmt.Sprintf("%v", i)
		}
	}
	return arrStr
}

func ParseStringInterfaceToInt64(k string, m map[string]interface{}) int64 {
	var (
		tmpStr string
		tmpInt int64
		ok     bool
		err    error
	)
	_, ok = m[k]
	if !ok {
		return -1
	}
	tmpStr, ok = m[k].(string)
	if !ok {
		return -1
	}
	tmpInt, err = strconv.ParseInt(tmpStr, 10, 64)
	if err != nil {
		return -1
	}
	return tmpInt
}

func ParseStringInterfaceToFloat64(k string, m map[string]interface{}) float64 {
	var (
		tmpStr   string
		tmpFloat float64
		ok       bool
		err      error
	)
	_, ok = m[k]
	if !ok {
		return -1
	}
	tmpStr, ok = m[k].(string)
	if !ok {
		return -1
	}
	tmpFloat, err = strconv.ParseFloat(tmpStr, 64)
	if err != nil {
		return -1
	}
	return tmpFloat
}
