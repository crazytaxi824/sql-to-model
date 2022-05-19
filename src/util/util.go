package util

import "strings"

func StructFieldName(src string) string {
	sa := strings.Split(src, "_")
	for i := range sa {
		if len(sa[i]) > 0 {
			if strings.ToUpper(sa[i]) == "ID" {
				sa[i] = "ID"
			} else {
				sa[i] = strings.ToUpper(sa[i][0:1]) + sa[i][1:]
			}
		}
	}

	return strings.Join(sa, "")
}
