package toolfunc

import (
	"errors"
	"reflect"
)

// CopyMatchingFields 复制两个结构体中相同字段名且类型一致的字段值
func CopyMatchingFields(dest, src interface{}) error {
	// 检查dest是否是结构体的指针
	destVal := reflect.ValueOf(dest)
	if destVal.Kind() != reflect.Ptr || destVal.Elem().Kind() != reflect.Struct {
		return errors.New("第一个参数必须是结构体的引用")
	}
	// 检查src是否是结构体
	srcVal := reflect.ValueOf(src)
	if srcVal.Kind() != reflect.Struct {
		return errors.New("第二个参数必须是结构体")
	}

	destVal = destVal.Elem()
	destType := destVal.Type()

	// 遍历目标结构体的字段
	for i := 0; i < destVal.NumField(); i++ {
		destField := destVal.Field(i)
		destFieldType := destType.Field(i)

		// 检查源结构体是否有相同名称且类型一致的字段
		srcField := srcVal.FieldByName(destFieldType.Name)
		if srcField.IsValid() && srcField.Type() == destField.Type() {
			if destField.CanSet() {
				destField.Set(srcField)
			}
		}
	}

	return nil
}
