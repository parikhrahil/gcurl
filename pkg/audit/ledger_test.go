package audit_test

import (
	"io"
	"strings"
	"testing"

	"github.com/parikhrahil/gcurl/pkg/audit"
)

func TestAuditReader_Read(t *testing.T) {
	reader := strings.NewReader("test")
	var count int64
	auditReader := audit.NewAuditReader(reader, &count)

	_, err := io.Copy(io.Discard, auditReader)
	if err != nil {
		t.Fatalf("Error not expected while reading from audit reader: %s", err.Error())
	}

	expectedBytes := int64(len("test"))
	if count != expectedBytes {
		t.Fatalf("Expected bytes: %d, got: %d", expectedBytes, count)
	}
}
