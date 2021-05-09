/*
This class contaions utility methods for ticketing-service
*/
package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"strings"

	"github.com/magiconair/properties"
)

var (
	propertyFile []string
	props        *properties.Properties
	// type fieldType
)

//ReadPropertyfile method is used to read the properties from relative path, ONLY for unit testcase
func ReadPropertyfile(filepath string) {
	propertyFile = []string{filepath}
	props, _ = properties.LoadFiles(propertyFile, properties.UTF8, true)

}

func init() {
	messages := "./configs/ticketing-service.properties"
	ReadPropertyfile(messages)
}

// JSONToStruct func is used to convert JSON to struct
func JSONToStruct(data []byte, datastruct interface{}) error {
	return json.Unmarshal(data, &datastruct)
}

//Utility Method to Write Response Message
func WriteResponseMessage(w http.ResponseWriter, status int, responseText []byte) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
	if len(responseText) > 0 {
		w.Write(responseText)
	}

}

// GetProperty is used to get the property value from property file
func GetProperty(propertyName string, params ...string) string {
	msg, ok := props.Get(propertyName)
	if !ok {
		return props.MustGet("Property is not available in property file for key " + propertyName)
	}
	placehdrCnt := strings.Count(msg, "{")
	if placehdrCnt == len(params) {
		for i, val := range params {
			repalcestr := fmt.Sprintf("%s%d%s", "{", i, "}")
			msg = strings.Replace(msg, repalcestr, val, -1)
		}
	}
	return msg
}

func GetEnv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

func StringifyMap(m map[string]string) string {
	b := new(bytes.Buffer)
	for key, value := range m {
		fmt.Fprintf(b, "%s=\"%s\"\n", key, value)
	}
	return b.String()
}

func JsonifyMap(m map[string]string) string {
	jsonString, _ := json.Marshal(m)
	return string(jsonString)
}

func DeepCopyMap(source map[string]string, destination map[string]string) {
	for key, val := range source {
		destination[key] = val
	}
}

// To check whether given type is of string or not
func IsString(k interface{}) bool {
	switch k.(type) {
	case string:
		if len(strings.TrimSpace(k.(string))) <= 0 {
			return false
		}
		return true
	default:
		return false
	}
}

func GetPkVal(v interface{}) string {
	switch v.(type) {
	case string:
		return v.(string)
	case float64, float32:
		if floatVal, ok := v.(float64); ok {
			if floatVal == float64(int64(floatVal)) {
				return fmt.Sprintf("%d", int64(floatVal))
			}
			return fmt.Sprintf("%f", v)
		}
		if floatVal, ok := v.(float32); ok {
			if floatVal == float32(int32(floatVal)) {
				return fmt.Sprintf("%d", int32(floatVal))
			}
			return fmt.Sprintf("%f", v)
		}
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", v)
	default:
		return ""
	}
	return ""
}

func StringToByteArrayConversion(result map[string]interface{}) map[string]interface{} {
	res := make(map[string]interface{})
	for key, value := range result {
		if value != nil {
			rv := reflect.ValueOf(value)
			switch rv.Kind() {
			case reflect.String:
				res[key] = []byte(value.(string))
			default:
				res[key] = value
			}
		}
	}
	return res
}
