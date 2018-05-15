package tigerblood

// errno and descriptions for APP-MOZLOG
// must be > 0 to indicate error https://wiki.mozilla.org/Firefox/Services/Logging#Application_Request_Summary_.28Type:.22request.summary.22.29

// Errno an error number for looking up error messages
type Errno uint

// Authentication errors
const (
	_ = iota
	// HawkAuthFormatError a hawk error
	HawkAuthFormatError
	// HawkReplayError hawk replay attack detected
	HawkReplayError
	// HawkErrNoAuth missing hawk auth header
	HawkErrNoAuth
	// HawkInvalidHash invalid hawk hash
	HawkInvalidHash
	// HawkCredError missing or incorrect hawk creds
	HawkCredError
	// HawkOtherAuthError other hawk errors
	HawkOtherAuthError
	// HawkMissingContentType no or invalid content type header
	HawkMissingContentType
	// HawkValidationError hawk validation failed
	HawkValidationError
	// HawkInvalidBodyHash invalid hawk body hash
	HawkInvalidBodyHash
	// HawkReadBodyError error reading the request body
	HawkReadBodyError
)

// missing global errors usually result in warnings or 500 errors
const (
	// MissingDB db not configured
	MissingDB = 20 + iota
	// MissingStatsdClient statsd client not configured
	MissingStatsdClient
	// MissingViolations violation penalties not set
	MissingViolations
)

// encoding/decoding errors
const (
	// BodyReadError error reading request body
	BodyReadError = 30 + iota
	// JSONMarshalError error marshalling json
	JSONMarshalError
	// JSONUnmarshalError error unmarshalling json
	JSONUnmarshalError
)

// validation errors usually result in a 400 error
const (
	// InvalidIPError IP/CIDR validation failure
	InvalidIPError = 40 + iota
	// InvalidReputationError reputation validation failure
	InvalidReputationError
	// InvalidViolationTypeError violation type validation failed
	InvalidViolationTypeError
	// TooManyIPViolationEntriesError too many IP Violation entries
	TooManyIPViolationEntriesError
	// DuplicateIPError when the same IP occurs in multiple entries
	DuplicateIPError
)

// missing parameter errors usually result in a 400 error
const (
	// MissingIPError no IP in request params or body
	MissingIPError = 50 + iota
	// MissingReputationError no reputation in request params or body
	MissingReputationError
	// MissingViolationTypeError no violation type in request params or body
	MissingViolationTypeError
	// MissingIPViolationEntryError no (for the multi violations endpoint)
	MissingIPViolationEntryError
)

// IO/DB errors
const (
	// DBError generic postgres or postgres driver error
	DBError = 60 + iota
	// CWDNotFound error when get CWD fails
	CWDNotFound
	// FileNotFound file not found error
	FileNotFound
)

// API key authentication errors
const (
	// APIKeyNotSpecified indicates the header value was not found
	APIKeyNotSpecified = 70 + iota
	// APIKeyInvalid indicates the key was not a configured credential
	APIKeyInvalid = iota
)

// UnknownError is for generic errors
const UnknownError = 999

// DescribeErrno returns a format string for the errno; not implemented for all errnos
func DescribeErrno(errno Errno) string {
	switch errno {
	case BodyReadError:
		return "Error reading the request body: %s"
	case JSONMarshalError:
		return "Error marshaling %s to JSON: %s"
	case JSONUnmarshalError:
		return "Error unmarshaling request body from JSON: %s"

	case InvalidIPError:
		return "Invalid IP: %s"
	case InvalidReputationError:
		return "Invalid reputation: %s"
	case InvalidViolationTypeError:
		return "Invalid violation type: %s"
	case TooManyIPViolationEntriesError:
		return "Too many IP, violation objects in request body"
	case DuplicateIPError:
		return "Duplicate IP found in multiple entries: %s"

	case MissingIPError:
		return "Error finding IP parameter"
	case MissingReputationError:
		return "Error finding reputation parameter in %s: %s"
	case MissingViolationTypeError:
		return "Error finding violation type: %s"
	case MissingIPViolationEntryError:
		return "Error finding an IP and violation type object in request body"

	case MissingDB:
		return "Could not find database"
	case MissingViolations:
		return "Could not find violation penalties"
	case MissingStatsdClient:
		return "Could not find statsdClient"

	case CWDNotFound:
		return "Error getting CWD: %s"
	case FileNotFound:
		return "Error finding file %s: %s"

	default:
		return "Error: %s"
	}
}
