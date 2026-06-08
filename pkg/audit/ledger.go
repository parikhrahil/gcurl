package audit

import "io"

type AuditReader struct {
	underlying io.Reader
	byteCount  *int64
}

func NewAuditReader(r io.Reader, count *int64) io.Reader {
	return &AuditReader{
		underlying: r,
		byteCount:  count,
	}
}

func (r *AuditReader) Read(p []byte) (int, error) {
	n, err := r.underlying.Read(p)
	if n > 0 {
		*r.byteCount += int64(n)
	}
	return n, err
}
