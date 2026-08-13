# AWS SES email provider

SES is an optional email backend registered into an `email.Registry`. Build with
`-tags ses` for the AWS implementation; without that tag registration installs a
stub that reports that SES was not compiled.

```go
registry := email.NewRegistry()
ses.Register(registry)
provider, err := registry.ProviderFromConfig(cfg) // cfg.EmailProvider == "ses"
if err != nil { return err }
err = provider.Send(ctx, mail.Address{Address: recipient}, rawRFC822Message)
```

Configuration supplies the AWS region and sender address; the AWS SDK resolves
credentials using its normal credential chain. `Send` expects a complete raw
RFC 822 message and does not build headers or MIME bodies.
