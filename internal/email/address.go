package email

import (
	"log"
	"net/mail"
)

// ParseAddress parses s into a mail.Address. On error it logs the failure
// and returns an Address with only the Address field set.
func ParseAddress(s string) mail.Address {
	a, err := mail.ParseAddress(s)
	if err != nil {
		log.Printf("parse address %q: %v", s, err)
		return mail.Address{Address: s}
	}
	return *a
}
