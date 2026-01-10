package admin

import (
	"fmt"
	"net/http"
	"net/mail"
	"sort"
	"strings"

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

	// Generate config text
	var configBuilder strings.Builder

	// Helper to append setting
	appendSetting := func(name, env, val, usage string) {
		configBuilder.WriteString(fmt.Sprintf("# %s: %s\n", name, usage))
		configBuilder.WriteString(fmt.Sprintf("%s=%s\n\n", env, val))
	}

	type OptionInfo struct {
		Name  string
		Env   string
		Value string
		Usage string
	}
	var options []OptionInfo

	values := config.ValuesMap(*cd.Config)

	for _, opt := range config.StringOptions {
		if emailKeys[opt.Env] {
			options = append(options, OptionInfo{opt.Name, opt.Env, values[opt.Env], opt.Usage})
		}
	}
	for _, opt := range config.BoolOptions {
		if emailKeys[opt.Env] {
			val := "false"
			if *opt.Target(cd.Config) {
				val = "true"
			}
			options = append(options, OptionInfo{opt.Name, opt.Env, val, opt.Usage})
		}
	}

	sort.Slice(options, func(i, j int) bool {
		return options[i].Name < options[j].Name
	})

	for _, opt := range options {
		appendSetting(opt.Name, opt.Env, opt.Value, opt.Usage)
	}

	initialConfigText := configBuilder.String()

	data := struct {
		ConfigText string
		ToAddress  string
		Result     string
		Error      string
	}{
		ConfigText: initialConfigText,
	}

	if r.Method == "POST" {
		data.ToAddress = r.FormValue("to_address")
		data.ConfigText = r.FormValue("config_text")

		// Reconstruct Config
		tempConfig := config.NewRuntimeConfig()
		*tempConfig = *cd.Config

		// Parse config text
		parsed := config.ParseEnvBytes([]byte(data.ConfigText))

		// Apply overrides
		for _, opt := range config.StringOptions {
			if val, ok := parsed[opt.Env]; ok {
				*opt.Target(tempConfig) = val
			}
		}
		for _, opt := range config.BoolOptions {
			if val, ok := parsed[opt.Env]; ok {
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

	AdminEmailTestPageTmpl.Handle(w, r, data)
}

const AdminEmailTestPageTmpl handlers.Page = "admin/emailTestPage.gohtml"
