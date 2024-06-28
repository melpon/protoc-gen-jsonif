package internal

import "strings"

func ToUpperCamel(name string) string {
	xs := strings.Split(name, "_")
	for i, x := range xs {
		if len(x) != 0 {
			xs[i] = strings.ToUpper(x[:1]) + x[1:]
		}
	}
	return strings.Join(xs, "")
}

func ToLowerCamel(name string) string {
	xs := strings.Split(name, "_")
	for i, x := range xs {
		if len(x) != 0 {
			if i == 0 {
				xs[i] = strings.ToLower(x[:1]) + x[1:]
			} else {
				xs[i] = strings.ToUpper(x[:1]) + x[1:]
			}
		}
	}
	return strings.Join(xs, "")
}

func ToSnakeCase(name string) string {
	result := ""
	for i, c := range name {
		if 'A' <= c && c <= 'Z' {
			if i != 0 {
				result += "_"
			}
			result += string(c - 'A' + 'a')
		} else {
			result += string(c)
		}
	}
	return result
}
