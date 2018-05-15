package tigerblood

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

type errnoTestTable struct {
	errno  Errno
	expect string
	a      []interface{}
}

var ett = []errnoTestTable{
	{BodyReadError, "Error reading the request body: error", []interface{}{"error"}},
	{JSONMarshalError, "Error marshaling test to JSON: error", []interface{}{"test", "error"}},
	{JSONUnmarshalError, "Error unmarshaling request body from JSON: error", []interface{}{"error"}},
	{InvalidIPError, "Invalid IP: test", []interface{}{"test"}},
	{InvalidReputationError, "Invalid reputation: test", []interface{}{"test"}},
	{InvalidViolationTypeError, "Invalid violation type: test", []interface{}{"test"}},
	{TooManyIPViolationEntriesError, "Too many IP, violation objects in request body", []interface{}{}},
	{MissingIPError, "Error finding IP parameter", []interface{}{}},
	{MissingReputationError, "Error finding reputation parameter in test: reputation",
		[]interface{}{"test", "reputation"}},
	{MissingViolationTypeError, "Error finding violation type: test", []interface{}{"test"}},
	{MissingIPViolationEntryError, "Error finding an IP and violation type object in request body",
		[]interface{}{}},
	{MissingDB, "Could not find database", []interface{}{}},
	{MissingDB, "Could not find database", []interface{}{}},
	{MissingViolations, "Could not find violation penalties", []interface{}{}},
	{MissingStatsdClient, "Could not find statsdClient", []interface{}{}},
	{CWDNotFound, "Error getting CWD: test", []interface{}{"test"}},
	{FileNotFound, "Error finding file path: test", []interface{}{"path", "test"}},
	{UnknownError, "Error: test", []interface{}{"test"}},
}

func TestDescribeErrno(t *testing.T) {
	for _, v := range ett {
		assert.Equal(t, v.expect, fmt.Sprintf(DescribeErrno(v.errno), v.a...))
	}
}
