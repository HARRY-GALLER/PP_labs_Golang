package stringutils


func Reversed(str string) string {
	var result string
	for i := len(str) - 1; i >= 0; i-- {
		result += string(str[i])
	}
	return result
}
