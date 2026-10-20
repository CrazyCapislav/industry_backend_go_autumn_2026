package main

func rotateRunes(s string, shift int) string {
	str := []rune(s)
	if len(str) == 0 {
		return s
	}
	shift = shift % len(str)
	if shift < 0 {
		shift += len(str)
	}
	return string(append(str[shift:], str[:shift]...))
}
