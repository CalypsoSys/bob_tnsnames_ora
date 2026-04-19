package babalu_tnsnames_ora

import (
	"strings"
	"testing"
)

func TestParseBasicAlias(t *testing.T) {
	input := `
# leading comment
SALES =
  (DESCRIPTION =
    (ADDRESS = (PROTOCOL = TCP)(HOST = db.example.com)(PORT = 1521))
    (CONNECT_DATA = (SERVICE_NAME = sales.example.com)))
`

	file, err := ParseString(input)
	if err != nil {
		t.Fatalf("ParseString() error = %v", err)
	}

	entry, ok := file.Entry("sales")
	if !ok {
		t.Fatal("expected SALES alias")
	}

	wantDescriptor := "(DESCRIPTION=(ADDRESS=(PROTOCOL=TCP)(HOST=db.example.com)(PORT=1521))(CONNECT_DATA=(SERVICE_NAME=sales.example.com)))"
	if got := entry.Descriptor(); got != wantDescriptor {
		t.Fatalf("Descriptor() = %q, want %q", got, wantDescriptor)
	}

	details := entry.Details()
	if len(details.Endpoints) != 1 {
		t.Fatalf("len(Endpoints) = %d, want 1", len(details.Endpoints))
	}
	if got := details.Endpoints[0].Host; got != "db.example.com" {
		t.Fatalf("Host = %q, want db.example.com", got)
	}
	if got := details.ConnectData.ServiceName; got != "sales.example.com" {
		t.Fatalf("ServiceName = %q, want sales.example.com", got)
	}

	ez, err := entry.EZConnect()
	if err != nil {
		t.Fatalf("EZConnect() error = %v", err)
	}
	if ez != "db.example.com:1521/sales.example.com" {
		t.Fatalf("EZConnect() = %q", ez)
	}
}

func TestParseMultipleAliasesPreservesOrder(t *testing.T) {
	input := `
FIRST=(DESCRIPTION=(ADDRESS=(PROTOCOL=TCP)(HOST=one)(PORT=1521))(CONNECT_DATA=(SID=ONE)))
SECOND=(DESCRIPTION=(ADDRESS=(PROTOCOL=TCP)(HOST=two)(PORT=1522))(CONNECT_DATA=(SERVICE_NAME=TWO)))
`

	file, err := ParseString(input)
	if err != nil {
		t.Fatalf("ParseString() error = %v", err)
	}

	names := file.AliasNames()
	got := strings.Join(names, ",")
	if got != "FIRST,SECOND" {
		t.Fatalf("AliasNames() = %q, want FIRST,SECOND", got)
	}
}

func TestParenthesizedAtomValue(t *testing.T) {
	input := `WALLET_LOCATION=(SOURCE=(METHOD=FILE)(METHOD_DATA=(DIRECTORY="/tmp/wallet")))`

	file, err := ParseString(input)
	if err != nil {
		t.Fatalf("ParseString() error = %v", err)
	}

	entry, ok := file.Entry("wallet_location")
	if !ok {
		t.Fatal("expected WALLET_LOCATION alias")
	}

	if _, err := entry.EZConnect(); err == nil {
		t.Fatal("expected EZConnect to fail for non-address descriptor")
	}
}

func TestParseSyntaxError(t *testing.T) {
	input := `BROKEN=(DESCRIPTION=(ADDRESS=(HOST=db.example.com)(PORT=1521))`

	if _, err := ParseString(input); err == nil {
		t.Fatal("expected parse error")
	}
}
