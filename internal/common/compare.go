package common

func CompareMapString(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		val, ok := b[k]
		if !ok || v != val {
			return false
		}
	}
	return true
}

func CompareSliceString(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
