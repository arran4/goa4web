package email

import (
	"log"
	"net/mail"
)

// ParseAddress returns the mail.Address parsed from s. If parsing fails,
// the returned Address has only the Address field set.
func ParseAddress(s string) mail.Address {
	a, err := mail.ParseAddress(s)
	if err != nil {
		log.Printf("parse mail address %q: %v", s, err)
		return mail.Address{Address: s}
	}
	return *a
}
