# goa4web
It's just my code! Please don't steal it!

# Caching and State Configuration

If you are using Cloudflare or a similar CDN, you must enforce the following Cache Rules so that authenticated or state-dependent HTML is correctly separated from anonymous caching:
1. Bypass Cache when `URI Path` matches `^/(login|logout|register|forgot|admin|private|account)` (or any other path responding differently per user).
2. Bypass Cache for all HTML responses when the `goa4web_session` cookie is present.
3. Respect origin Cache-Control headers natively instead of overriding with a fixed Edge TTL.

The application automatically issues `Cache-Control: no-store` and `Cloudflare-CDN-Cache-Control: no-store` for authenticated and active state-changing endpoints (e.g. login POST responses).

## Session Security

You must persist a secure `SESSION_SECRET` (the cookie signing key) across instances and re-deployments. Rotating or losing this secret will immediately invalidate all existing user sessions globally, terminating active logins and potentially breaking mid-flight redirect continuations containing encrypted parameters.
