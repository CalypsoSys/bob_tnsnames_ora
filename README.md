# tnsnames

`tnsnames` is a small pure-Go parser for Oracle `tnsnames.ora` files.

It parses the balanced-parentheses TNS descriptor format into a tree, supports case-insensitive alias lookup, re-renders canonical descriptor strings, and can derive a simple client connect string for straightforward descriptors that contain `HOST`, `PORT`, and `SERVICE_NAME` or `SID`.

## Scope

This package is intended for applications that need to read `tnsnames.ora` and extract connection details in pure Go.

It does not attempt to replace the Oracle client resolver. In particular, the `ConnectString` helper is best-effort and intentionally limited to simpler descriptors.

## Features

- Parse `tnsnames.ora` from a string, byte slice, or file.
- Preserve top-level alias order.
- Look up aliases case-insensitively.
- Re-render aliases into canonical descriptor strings.
- Extract common `ADDRESS` and `CONNECT_DATA` fields.
- Build a simple `host:port/service` connect string for straightforward aliases.

## Install

```bash
go get github.com/CalypsoSys/bob_tnsnames_ora
```

## Example

```go
package main

import (
	"fmt"
	"log"

	tnsnames "github.com/CalypsoSys/bob_tnsnames_ora"
)

func main() {
	file, err := tnsnames.ParseFile("tnsnames.ora")
	if err != nil {
		log.Fatal(err)
	}

	entry, err := file.MustEntry("SALES")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(entry.Descriptor())

	connectString, err := entry.ConnectString()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(connectString)
}
```

## API Summary

```go
file, err := tnsnames.ParseFile("tnsnames.ora")
entry, err := file.MustEntry("SALES")

descriptor := entry.Descriptor()
details := entry.Details()
connectString, err := entry.ConnectString()
```

## Supported Shapes

The parser handles standard nested TNS assignments such as:

```text
SALES =
  (DESCRIPTION =
    (ADDRESS = (PROTOCOL = TCP)(HOST = db.example.com)(PORT = 1521))
    (CONNECT_DATA = (SERVICE_NAME = sales.example.com)))
```

## Limitations

- `ConnectString` uses the first parsed `ADDRESS` entry.
- Complex failover, load-balancing, and advanced Oracle Net behaviors are preserved in the tree but not fully interpreted.
- The package parses descriptors; it does not implement Oracle client search rules, environment-variable resolution, or OCI behavior.
