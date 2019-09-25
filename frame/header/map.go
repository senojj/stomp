package header

import (
	"fmt"
	"io"
)

type Map map[Header][]string

func (m Map) Append(key Header, value string) {
	m[key] = append(m[key], value)
}

func (m Map) Prepend(key Header, value string) {
	v := m[key]
	n := make([]string, 0, len(v)+1)
	n = append(n, value)
	n = append(n, v...)
	m[key] = n
}

func (m Map) Set(key Header, value string) {
	m[key] = []string{value}
}

func (m Map) Get(key Header) string {
	if m == nil {
		return ""
	}
	v := m[key]

	if len(v) == 0 {
		return ""
	}
	return v[0]
}

func (m Map) Del(key Header) {
	delete(m, key)
}

func (m Map) WriteTo(w io.Writer) (int64, error) {
	var written int64 = 0

	for k, v := range m {
		for _, i := range v {
			out := fmt.Sprintf("%s:%s\n", k, i)
			b, werr := w.Write([]byte(out))

			if nil != werr {
				return written, fmt.Errorf("problem writing header: %v", werr)
			}
			written = written + int64(b)
		}
	}
	return written, nil
}