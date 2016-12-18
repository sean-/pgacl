package pgacl

import (
	"bytes"
	"fmt"
	"strings"
)

// Schema models the privileges of a schema aclitem
type Schema struct {
	Role        string
	GrantedBy   string
	Create      bool
	CreateGrant bool
	Usage       bool
	UsageGrant  bool
}

const numSchemaOpts = 4

// NewSchema parses a PostgreSQL ACL string for a schema and returns a Schema
// object
func NewSchema(aclStr string) (Schema, error) {
	acl := Schema{}
	idx := strings.IndexByte(aclStr, '=')
	if idx == -1 {
		return Schema{}, fmt.Errorf("invalid aclStr format: %+q", aclStr)
	}

	acl.Role = aclStr[:idx]

	aclLen := len(aclStr)
	var i int
	withGrant := func() bool {
		if i+1 >= aclLen {
			return false
		}

		if aclStr[i+1] == '*' {
			i++
			return true
		}

		return false
	}

SCAN:
	for i = idx + 1; i < aclLen; i++ {
		switch aclStr[i] {
		case 'C':
			acl.Create = true
			if withGrant() {
				acl.CreateGrant = true
			}
		case 'U':
			acl.Usage = true
			if withGrant() {
				acl.UsageGrant = true
			}
		case '/':
			if i+1 <= aclLen {
				acl.GrantedBy = aclStr[i+1:]
			}
			break SCAN
		default:
			return Schema{}, fmt.Errorf("invalid byte %c in schema ACL at %d: %+q", aclStr[i], i, aclStr)
		}
	}

	return acl, nil
}

// Merge merges in into the receiver
func (s Schema) Merge(in Schema) {
	if in.Role != "" {
		s.Role = in.Role
	}

	if in.GrantedBy != "" {
		s.GrantedBy = in.GrantedBy
	}

	if in.Create {
		s.Create = in.Create
	}

	if in.CreateGrant {
		s.CreateGrant = in.CreateGrant
	}

	if in.Usage {
		s.Usage = in.Usage
	}

	if in.UsageGrant {
		s.UsageGrant = in.UsageGrant
	}
}

// String creates a PostgreSQL native output for the ACLs that apply to a
// schema.
func (s Schema) String() string {
	b := new(bytes.Buffer)
	b.Grow(len(s.Role) + numSchemaOpts + 1)

	fmt.Fprint(b, s.Role, "=")

	if s.Usage {
		fmt.Fprint(b, "U")
		if s.UsageGrant {
			fmt.Fprint(b, "*")
		}
	}

	if s.Create {
		fmt.Fprint(b, "C")
		if s.CreateGrant {
			fmt.Fprint(b, "*")
		}
	}

	if s.GrantedBy != "" {
		fmt.Fprint(b, "/", s.GrantedBy)
	}

	return b.String()
}
