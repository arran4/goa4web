package admin

import "strconv"

func appendID(existing any, id int) string {
	if s, ok := existing.(string); ok && s != "" {
		return s + "," + strconv.Itoa(id)
	}
	return strconv.Itoa(id)
}
