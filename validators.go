package tigerblood

import (
	log "github.com/sirupsen/logrus"
	"go.mozilla.org/mozlogrus"
	"net"
	"regexp"
)

func init() {
	mozlogrus.Enable("tigerblood")
}

// IsValidReputationCIDROrIP checks that string can be parsed as a cidr or ip
func IsValidReputationCIDROrIP(s string) bool {
	ip, _, err := net.ParseCIDR(s)
	if err == nil {
		return true
	}

	ip = net.ParseIP(s)
	if ip != nil {
		return true
	}

	return false
}

// IsValidReputation checks if reputation is in [0, 100]
func IsValidReputation(reputation uint) bool {
	// see createReputationTableSQL db.go
	return reputation >= 0 && reputation <= 100
}

// IsValidViolationPenalty checks if violation penalty is in [0, 100]
func IsValidViolationPenalty(penalty uint64) bool {
	return penalty >= 0 && penalty <= 100
}

var violationRegex = regexp.MustCompile(`[:\w]{1,255}`)

// IsValidViolationName checks if a violation name matches [:\w]{1,255}
func IsValidViolationName(name string) bool {
	return violationRegex.MatchString(name)
}

// IsValidReputationEntry checks if a ReputationEntry has valid IP and reputation fields
func IsValidReputationEntry(entry ReputationEntry) bool {
	return IsValidReputationCIDROrIP(entry.IP) && IsValidReputation(entry.Reputation)
}

// ValidateIPViolationEntryAndGetPenalty validates violation type and returns violation penalty and Errno or 0 for no error
func ValidateIPViolationEntryAndGetPenalty(entry IPViolationEntry) (uint, Errno) {
	if !IsValidViolationName(entry.Violation) {
		log.WithFields(log.Fields{"errno": InvalidViolationTypeError}).Infof(DescribeErrno(InvalidViolationTypeError), entry.Violation)
		return 0, InvalidViolationTypeError
	}

	if violationPenalties == nil {
		log.WithFields(log.Fields{"errno": MissingViolations}).Warnf(DescribeErrno(MissingViolations))
		return 0, MissingViolations
	}

	// lookup violation weight in config map
	var penalty, ok = violationPenalties[entry.Violation]
	if !ok {
		log.WithFields(log.Fields{"errno": MissingViolationTypeError}).Infof("Could not find violation type: %s", entry.Violation)
		return 0, MissingViolationTypeError
	}

	return uint(penalty), 0
}
