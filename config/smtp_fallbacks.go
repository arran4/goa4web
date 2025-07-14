package config

import (
	"fmt"
	"log"
	"strings"
)

// ApplySMTPFallbacks ensures EmailFrom and EmailSMTPUser are set when using the
// SMTP provider. If one is blank but the other looks like an email address it
// is copied over and the action is logged. If both remain empty an error is
// returned. If both are set to different addresses a warning is logged.
func ApplySMTPFallbacks(cfg *RuntimeConfig) error {
	if strings.ToLower(cfg.EmailProvider) != "smtp" {
		return nil
	}
	if cfg.EmailFrom == "" && cfg.EmailSMTPUser == "" {
		return fmt.Errorf("smtp: %s and %s not set", EnvSMTPUser, EnvEmailFrom)
	}
	if cfg.EmailFrom == "" && strings.Contains(cfg.EmailSMTPUser, "@") {
		cfg.EmailFrom = cfg.EmailSMTPUser
		log.Printf("%s not set, using %s=%q", EnvEmailFrom, EnvSMTPUser, cfg.EmailFrom)
	} else if cfg.EmailSMTPUser == "" && strings.Contains(cfg.EmailFrom, "@") {
		cfg.EmailSMTPUser = cfg.EmailFrom
		log.Printf("%s not set, using %s=%q", EnvSMTPUser, EnvEmailFrom, cfg.EmailSMTPUser)
	} else {
		if strings.Contains(cfg.EmailSMTPUser, "@") && strings.Contains(cfg.EmailFrom, "@") && cfg.EmailSMTPUser != cfg.EmailFrom {
			log.Printf("%s=%q and %s=%q differ", EnvSMTPUser, cfg.EmailSMTPUser, EnvEmailFrom, cfg.EmailFrom)
		} else {
			log.Printf("using %s=%q and %s=%q", EnvSMTPUser, cfg.EmailSMTPUser, EnvEmailFrom, cfg.EmailFrom)
		}
	}
	return nil
}
