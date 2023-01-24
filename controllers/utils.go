package controllers

func addToMap(obj, items map[string]string) map[string]string {
	for k, v := range items {
		obj[k] = v
	}
	return obj
}

func int32Ptr(i int32) *int32 {
	return &i
}

func int64Ptr(i int64) *int64 {
	return &i
}

func booPtr(b bool) *bool {
	return &b
}
