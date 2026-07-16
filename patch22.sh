sed -i 's/if parsed.Host != cd.Config.HTTPHostname/if cd.Config != nil \&\& parsed.Host != cd.Config.HTTPHostname/g' core/common/coredata.go
