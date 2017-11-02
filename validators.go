package tigerblood

import (
	"net"
	"regexp"
)

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

func IsValidReputation(reputation uint) bool {
	// see createReputationTableSQL db.go
	return reputation >= 0 && reputation <= 100
}

func IsValidViolationPenalty(penalty uint64) bool {
	return penalty >= 0 && penalty <= 100
}

var violationRegex = regexp.MustCompile(`[:\w]{1,255}`)

func IsValidViolationName(name string) bool {
	return violationRegex.MatchString(name)
}

func IsValidReputationEntry(entry ReputationEntry) bool {
	return IsValidReputationCIDROrIP(entry.IP) && IsValidReputation(entry.Reputation)
}
