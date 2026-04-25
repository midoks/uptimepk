package utils

// BoolToInt 将布尔值转换为整数
// true -> 1, false -> 0
func BoolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// IntToBool 将整数转换为布尔值
// 0 -> false, 非0 -> true
func IntToBool(i int) bool {
	return i != 0
}
