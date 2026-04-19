package babalu_tnsnames_ora

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

// File represents the parsed contents of a tnsnames.ora file.
type File struct {
	entries map[string]*Entry
	order   []string
}

// Entry is a top-level alias assignment.
type Entry struct {
	Name  string
	Value Value
}

// Value is either an atom or a grouped list of assignments.
type Value struct {
	Atom     string
	Children []*Node
}

// Node is a single KEY=VALUE assignment inside a group.
type Node struct {
	Key   string
	Value Value
}

// Endpoint is a best-effort extraction of an ADDRESS node.
type Endpoint struct {
	Protocol string
	Host     string
	Port     int
}

// ConnectData contains common CONNECT_DATA fields.
type ConnectData struct {
	ServiceName  string
	SID          string
	InstanceName string
	Server       string
}

// DescriptorDetails summarizes common fields from a parsed descriptor.
type DescriptorDetails struct {
	Endpoints   []Endpoint
	ConnectData ConnectData
}

// Parse reads tnsnames.ora content from bytes.
func Parse(data []byte) (*File, error) {
	return ParseString(string(data))
}

// ParseFile reads and parses a tnsnames.ora file from disk.
func ParseFile(path string) (*File, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Parse(data)
}

// ParseString parses tnsnames.ora content from a string.
func ParseString(input string) (*File, error) {
	p := &parser{src: stripComments(input)}
	file := &File{
		entries: make(map[string]*Entry),
	}

	for {
		p.skipSpace()
		if p.eof() {
			break
		}

		entry, err := p.parseEntry()
		if err != nil {
			return nil, err
		}

		key := strings.ToUpper(entry.Name)
		if _, exists := file.entries[key]; !exists {
			file.order = append(file.order, key)
		}
		file.entries[key] = entry
	}

	return file, nil
}

// Entries returns the parsed entries in file order.
func (f *File) Entries() []*Entry {
	out := make([]*Entry, 0, len(f.order))
	for _, name := range f.order {
		out = append(out, f.entries[name])
	}
	return out
}

// AliasNames returns the known aliases in file order.
func (f *File) AliasNames() []string {
	out := make([]string, len(f.order))
	copy(out, f.order)
	return out
}

// Entry returns an alias by name, case-insensitively.
func (f *File) Entry(alias string) (*Entry, bool) {
	entry, ok := f.entries[strings.ToUpper(strings.TrimSpace(alias))]
	return entry, ok
}

// MustEntry returns an alias or an error if it does not exist.
func (f *File) MustEntry(alias string) (*Entry, error) {
	entry, ok := f.Entry(alias)
	if !ok {
		return nil, fmt.Errorf("tnsnames: alias %q not found", alias)
	}
	return entry, nil
}

// Descriptor returns the canonical descriptor string for the alias.
func (f *File) Descriptor(alias string) (string, error) {
	entry, err := f.MustEntry(alias)
	if err != nil {
		return "", err
	}
	return entry.Descriptor(), nil
}

// ConnectString returns a simple host:port/service connect string for the alias.
func (f *File) ConnectString(alias string) (string, error) {
	entry, err := f.MustEntry(alias)
	if err != nil {
		return "", err
	}
	return entry.ConnectString()
}

// Descriptor renders the right-hand side of the alias assignment.
func (e *Entry) Descriptor() string {
	return e.Value.String()
}

// String renders the full alias assignment.
func (e *Entry) String() string {
	return fmt.Sprintf("%s=%s", e.Name, e.Value.String())
}

// Details extracts common connection fields from the descriptor tree.
func (e *Entry) Details() DescriptorDetails {
	var details DescriptorDetails
	walkValue(e.Value, func(node *Node) {
		switch strings.ToUpper(node.Key) {
		case "ADDRESS":
			details.Endpoints = append(details.Endpoints, extractEndpoint(node))
		case "CONNECT_DATA":
			data := extractConnectData(node)
			if data.ServiceName != "" {
				details.ConnectData.ServiceName = data.ServiceName
			}
			if data.SID != "" {
				details.ConnectData.SID = data.SID
			}
			if data.InstanceName != "" {
				details.ConnectData.InstanceName = data.InstanceName
			}
			if data.Server != "" {
				details.ConnectData.Server = data.Server
			}
		}
	})
	return details
}

// EZConnect returns a simple host:port/service connect string when possible.
func (e *Entry) EZConnect() (string, error) {
	details := e.Details()
	if len(details.Endpoints) == 0 {
		return "", fmt.Errorf("tnsnames: alias %q has no ADDRESS entries", e.Name)
	}

	endpoint := details.Endpoints[0]
	if endpoint.Host == "" {
		return "", fmt.Errorf("tnsnames: alias %q is missing ADDRESS/HOST", e.Name)
	}
	if endpoint.Port == 0 {
		return "", fmt.Errorf("tnsnames: alias %q is missing ADDRESS/PORT", e.Name)
	}

	target := details.ConnectData.ServiceName
	if target == "" {
		target = details.ConnectData.SID
	}
	if target == "" {
		return "", fmt.Errorf("tnsnames: alias %q is missing CONNECT_DATA/SERVICE_NAME or SID", e.Name)
	}

	return fmt.Sprintf("%s:%d/%s", endpoint.Host, endpoint.Port, target), nil
}

// ConnectString is an alias for EZConnect.
func (e *Entry) ConnectString() (string, error) {
	return e.EZConnect()
}

// SortedAliasNames returns aliases sorted alphabetically.
func (f *File) SortedAliasNames() []string {
	out := f.AliasNames()
	sort.Strings(out)
	return out
}

// String renders the whole file in canonical form.
func (f *File) String() string {
	var parts []string
	for _, entry := range f.Entries() {
		parts = append(parts, entry.String())
	}
	return strings.Join(parts, "\n")
}

// String renders a grouped value or atom.
func (v Value) String() string {
	if len(v.Children) == 0 {
		return v.Atom
	}

	var b strings.Builder
	for _, child := range v.Children {
		b.WriteString(child.String())
	}
	return b.String()
}

// String renders a KEY=VALUE assignment.
func (n *Node) String() string {
	return fmt.Sprintf("(%s=%s)", n.Key, n.Value.String())
}

type parser struct {
	src string
	pos int
}

func (p *parser) parseEntry() (*Entry, error) {
	name, err := p.parseKey()
	if err != nil {
		return nil, err
	}
	if err := p.expect('='); err != nil {
		return nil, err
	}
	value, err := p.parseValue()
	if err != nil {
		return nil, err
	}
	return &Entry{Name: name, Value: value}, nil
}

func (p *parser) parseNode() (*Node, error) {
	if err := p.expect('('); err != nil {
		return nil, err
	}

	key, err := p.parseKey()
	if err != nil {
		return nil, err
	}
	if err := p.expect('='); err != nil {
		return nil, err
	}

	value, err := p.parseValue()
	if err != nil {
		return nil, err
	}

	p.skipSpace()
	if err := p.expect(')'); err != nil {
		return nil, err
	}

	return &Node{Key: key, Value: value}, nil
}

func (p *parser) parseValue() (Value, error) {
	p.skipSpace()
	if p.eof() {
		return Value{}, p.errorf("unexpected end of input while parsing value")
	}

	if p.peek() != '(' {
		atom := p.parseAtom()
		if atom == "" {
			return Value{}, p.errorf("expected value")
		}
		return Value{Atom: atom}, nil
	}

	var children []*Node
	for !p.eof() {
		p.skipSpace()
		if p.eof() || p.peek() != '(' {
			break
		}
		node, err := p.parseNode()
		if err != nil {
			return Value{}, err
		}
		children = append(children, node)
	}
	if len(children) == 0 {
		return Value{}, p.errorf("expected grouped assignments")
	}
	return Value{Children: children}, nil
}

func (p *parser) parseAtom() string {
	start := p.pos
	for !p.eof() {
		ch := p.peek()
		if ch == '\n' || ch == '\r' {
			break
		}
		if ch == ')' {
			break
		}
		if ch == '(' {
			break
		}
		p.pos++
	}
	return strings.TrimSpace(p.src[start:p.pos])
}

func (p *parser) parseKey() (string, error) {
	p.skipSpace()
	start := p.pos
	for !p.eof() {
		ch := p.peek()
		if ch == '=' || ch == ')' || ch == '(' || unicode.IsSpace(rune(ch)) {
			break
		}
		p.pos++
	}

	key := strings.TrimSpace(p.src[start:p.pos])
	if key == "" {
		return "", p.errorf("expected key")
	}
	return key, nil
}

func (p *parser) expect(ch byte) error {
	p.skipSpace()
	if p.eof() || p.peek() != ch {
		return p.errorf("expected %q", ch)
	}
	p.pos++
	return nil
}

func (p *parser) skipSpace() {
	for !p.eof() && unicode.IsSpace(rune(p.peek())) {
		p.pos++
	}
}

func (p *parser) eof() bool {
	return p.pos >= len(p.src)
}

func (p *parser) peek() byte {
	return p.src[p.pos]
}

func (p *parser) errorf(format string, args ...any) error {
	return fmt.Errorf("tnsnames: %s at offset %d", fmt.Sprintf(format, args...), p.pos)
}

func stripComments(input string) string {
	var out strings.Builder
	inQuote := false

	for i := 0; i < len(input); i++ {
		ch := input[i]
		if ch == '"' {
			inQuote = !inQuote
			out.WriteByte(ch)
			continue
		}

		if !inQuote && (ch == '#' || ch == ';') {
			for i < len(input) && input[i] != '\n' {
				i++
			}
			if i < len(input) {
				out.WriteByte(input[i])
			}
			continue
		}

		out.WriteByte(ch)
	}

	return out.String()
}

func walkValue(value Value, visit func(*Node)) {
	for _, child := range value.Children {
		visit(child)
		walkValue(child.Value, visit)
	}
}

func extractEndpoint(node *Node) Endpoint {
	var endpoint Endpoint
	for _, child := range node.Value.Children {
		switch strings.ToUpper(child.Key) {
		case "PROTOCOL":
			endpoint.Protocol = child.Value.Atom
		case "HOST":
			endpoint.Host = child.Value.Atom
		case "PORT":
			port, _ := strconv.Atoi(child.Value.Atom)
			endpoint.Port = port
		}
	}
	return endpoint
}

func extractConnectData(node *Node) ConnectData {
	var data ConnectData
	for _, child := range node.Value.Children {
		switch strings.ToUpper(child.Key) {
		case "SERVICE_NAME":
			data.ServiceName = child.Value.Atom
		case "SID":
			data.SID = child.Value.Atom
		case "INSTANCE_NAME":
			data.InstanceName = child.Value.Atom
		case "SERVER":
			data.Server = child.Value.Atom
		}
	}
	return data
}
