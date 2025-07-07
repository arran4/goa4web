package runtimeconfig

func resolveSMTPStartTLS(cli, file, env string) bool {
	if b, ok := parseBool(cli); ok {
		return b
	}
	if b, ok := parseBool(file); ok {
		return b
	}
	if b, ok := parseBool(env); ok {
		return b
	}
	return true
}
