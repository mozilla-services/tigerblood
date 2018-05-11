package tigerblood

// errno and descriptions for APP-MOZLOG
// must be > 0 to indicate error https://wiki.mozilla.org/Firefox/Services/Logging#Application_Request_Summary_.28Type:.22request.summary.22.29

// Errno an error number for looking up error messages
type Errno uint

const (
	// auth errors
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

	// missing global errors usually result in warnings or 500 errors

	// MissingDB db not configured
	MissingDB = 20
	// MissingStatsdClient statsd client not configured
	MissingStatsdClient = iota
	// MissingViolations violation penalties not set
	MissingViolations = iota

	// encoding/decoding errors

	// BodyReadError error reading request body
	BodyReadError = 30
	// JSONMarshalError error marshalling json
	JSONMarshalError = iota
	// JSONUnmarshalError error unmarshalling json
	JSONUnmarshalError = iota

	// validation errors usually result in a 400 error

	// InvalidIPError IP/CIDR validation failure
	InvalidIPError = 40
	// InvalidReputationError reputation validation failure
	InvalidReputationError = iota
	// InvalidViolationTypeError violation type validation failed
	InvalidViolationTypeError = iota
	// TooManyIPViolationEntriesError too many IP Violation entries
	TooManyIPViolationEntriesError = iota
	// DuplicateIPError when the same IP occurs in multiple entries
	DuplicateIPError = iota

	// missing parameter errors usually result in a 400 error

	// MissingIPError no IP in request params or body
	MissingIPError = 50
	// MissingReputationError no reputation in request params or body
	MissingReputationError = iota
	// MissingViolationTypeError no violation type in request params or body
	MissingViolationTypeError = iota
	// MissingIPViolationEntryError no (for the multi violations endpoint)
	MissingIPViolationEntryError = iota

	// IO/DB errors

	// DBError generic postgres or postgres driver error
	DBError = 60
	// CWDNotFound error when get CWD fails
	CWDNotFound = iota
	// FileNotFound file not found error
	FileNotFound = iota

	// API key authentication errors

	// APIKeyNotSpecified indicates the header value was not found
	APIKeyNotSpecified = 70
	// APIKeyInvalid indicates the key was not a configured credential
	APIKeyInvalid = iota

	// Unknown errors

	// UnknownError is for generic errors
	UnknownError = 999
)

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
		return "Duplicate IP found in multiple entries"
	}

	switch errno {
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
