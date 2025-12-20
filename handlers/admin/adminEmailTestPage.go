package admin

import (
	"fmt"
	"net/http"
	"net/mail"
	"sort"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
)

func (h *Handlers) AdminEmailTestPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Email Tester"

	emailKeys := map[string]bool{
		config.EnvEmailProvider:      true,
		config.EnvSMTPHost:           true,
		config.EnvSMTPPort:           true,
		config.EnvSMTPUser:           true,
		config.EnvSMTPPass:           true,
		config.EnvSMTPAuth:           true,
		config.EnvEmailFrom:          true,
		config.EnvJMAPEndpoint:       true,
		config.EnvJMAPAccount:        true,
		config.EnvJMAPIdentity:       true,
		config.EnvJMAPUser:           true,
		config.EnvJMAPPass:           true,
		config.EnvSendGridKey:        true,
		config.EnvAWSRegion:          true,
		config.EnvSMTPStartTLS:       true,
		config.EnvJMAPInsecure:       true,
		config.EnvEmailSubjectPrefix: true,
		config.EnvEmailSignOff:       true,
	}

	type Setting struct {
		Name  string
		Env   string
		Value string
		Usage string
	}

	var settings []Setting

	// Map current config values
	values := config.ValuesMap(*cd.Config)

	// Process String Options
	for _, opt := range config.StringOptions {
		if emailKeys[opt.Env] {
			val := values[opt.Env]
			if r.Method == "POST" {
				val = r.FormValue(opt.Env)
			}
			settings = append(settings, Setting{
				Name:  opt.Name,
				Env:   opt.Env,
				Value: val,
				Usage: opt.Usage,
			})
		}
	}

	// Process Bool Options
	for _, opt := range config.BoolOptions {
		if emailKeys[opt.Env] {
			val := "false"
			// If POST, read from form
			if r.Method == "POST" {
				val = r.FormValue(opt.Env)
			} else if *opt.Target(cd.Config) {
				val = "true"
			}

			settings = append(settings, Setting{
				Name:  opt.Name,
				Env:   opt.Env,
				Value: val,
				Usage: opt.Usage,
			})
		}
	}

	// Sort settings by Name for better UX
	sort.Slice(settings, func(i, j int) bool {
		return settings[i].Name < settings[j].Name
	})

	data := struct {
		Settings  []Setting
		ToAddress string
		Result    string
		Error     string
	}{
		Settings: settings,
	}

	if r.Method == "POST" {
		data.ToAddress = r.FormValue("to_address")

		// Reconstruct Config
		tempConfig := config.NewRuntimeConfig()
		// Copy current config
		*tempConfig = *cd.Config

		// Override
		for _, opt := range config.StringOptions {
			if emailKeys[opt.Env] {
				val := r.FormValue(opt.Env)
				*opt.Target(tempConfig) = val
			}
		}
		for _, opt := range config.BoolOptions {
			if emailKeys[opt.Env] {
				val := r.FormValue(opt.Env)
				*opt.Target(tempConfig) = (val == "true")
			}
		}

		// Now test
		if data.ToAddress == "" {
			data.Error = "To Address is required."
		} else {
			p, err := cd.EmailRegistry().ProviderFromConfig(tempConfig)
			if err != nil {
				data.Error = fmt.Sprintf("Error creating provider: %v", err)
			} else {
				err = p.TestConfig(r.Context())
				if err != nil {
					data.Error = fmt.Sprintf("Config Test Failed: %v", err)
				} else {
					err = p.Send(r.Context(), mail.Address{Address: data.ToAddress}, []byte("Subject: Test Email\r\n\r\nThis is a test email from Goa4Web admin dashboard."))
					if err != nil {
						data.Error = fmt.Sprintf("Send Failed: %v", err)
					} else {
						data.Result = "Email sent successfully!"
					}
				}
			}
		}
	}

	handlers.TemplateHandler(w, r, "admin/emailTestPage.gohtml", data)
}
