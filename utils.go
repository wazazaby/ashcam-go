package ashcam

import "unsafe"

func concat(parts ...string) string {
	if parts == nil {
		return ""
	}

	buf := make([]byte, 0, 128)

	for _, part := range parts {
		buf = append(buf, part...)
	}

	return *(*string)(unsafe.Pointer(&buf))
}
