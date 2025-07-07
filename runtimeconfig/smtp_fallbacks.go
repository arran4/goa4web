package runtimeconfig

import (
	"fmt"
	"log"
	"strings"

	"github.com/arran4/goa4web/config"
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
		return fmt.Errorf("smtp: %s and %s not set", config.EnvSMTPUser, config.EnvEmailFrom)
	}
	if cfg.EmailFrom == "" && strings.Contains(cfg.EmailSMTPUser, "@") {
		cfg.EmailFrom = cfg.EmailSMTPUser
		log.Printf("%s not set, using %s=%q", config.EnvEmailFrom, config.EnvSMTPUser, cfg.EmailFrom)
	} else if cfg.EmailSMTPUser == "" && strings.Contains(cfg.EmailFrom, "@") {
		cfg.EmailSMTPUser = cfg.EmailFrom
		log.Printf("%s not set, using %s=%q", config.EnvSMTPUser, config.EnvEmailFrom, cfg.EmailSMTPUser)
	} else {
		if strings.Contains(cfg.EmailSMTPUser, "@") && strings.Contains(cfg.EmailFrom, "@") && cfg.EmailSMTPUser != cfg.EmailFrom {
			log.Printf("%s=%q and %s=%q differ", config.EnvSMTPUser, cfg.EmailSMTPUser, config.EnvEmailFrom, cfg.EmailFrom)
		} else {
			log.Printf("using %s=%q and %s=%q", config.EnvSMTPUser, cfg.EmailSMTPUser, config.EnvEmailFrom, cfg.EmailFrom)
		}
	}
	return nil
}
