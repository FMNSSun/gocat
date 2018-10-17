package gocat

func isident(rn rune) bool {
	return isletter(rn) || rn == '.' || rn == ':'
}

func isdigit(rn rune) bool {
	return rn >= '0' && rn <= '9'
}

func isletter(rn rune) bool {
	if rn >= 'a' && rn <= 'z' {
		return true
	}

	if rn >= 'A' && rn <= 'Z' {
		return true
	}

	return false
}

func iswhitespace(rn rune) bool {
	return rn == '\t' || rn == ' ' || rn == '\r' || rn == '\n'
}
