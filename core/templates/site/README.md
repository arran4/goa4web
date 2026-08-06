# HTML Template Structure

The templates are organized into clear conceptual layers to make finding, maintaining, and reusing fragments easier as the project grows.

## Directory Layout

*   `layouts/`: Contains root page shells, base `<head>` tags, headers, and footers (e.g., `head.gohtml`, `header.gohtml`, `footer.gohtml`).
*   `pages/`: Route-oriented entry templates.
    *   `pages/auth/`: Login, register, forgot password, and other authentication-related pages.
    *   `pages/user/`: User-specific generic pages (like reset password).
    *   `pages/misc/`: Shared utility pages (like 404 Not Found, redirect loops, tasks, confirmations).
*   `partials/`: Reusable fragments.
    *   `partials/common/`: Shared generic UI fragments used across different domains (e.g., `pagination.gohtml`, `comment.gohtml`, `_share.gohtml`).
    *   `partials/forms/`: Shared form components (e.g., comboboxes).
*   `domains/`: Feature-oriented templates.
    *   `domains/{domain_name}/`: Contains pages and partials specifically used by that domain (e.g., `linker/`, `admin/`, `forum/`). Reusable fragments that only belong to a domain stay here, rather than going to `partials/common/`.

## Naming Conventions

*   **Pages:** Suffix templates that act as entry point views with `Page.gohtml` (e.g., `adminPage.gohtml`, `loginPage.gohtml`).
*   **Partials:** Try to keep names succinct and indicative of what they output (e.g., `comment.gohtml`, `pagination.gohtml`).
*   **Template definitions:** Files containing `{{ define "name" }}` are registered with that exact name. Files without `define` are registered under their file path relative to `core/templates/site` (e.g., `domains/linker/showPage.gohtml`).

## Adding New Templates

1.  Identify whether it's a full page (put in `pages/` or `domains/{feature}/`) or a reusable fragment (put in `partials/` or `domains/{feature}/`).
2.  Reference the template by its defined name (`{{ template "definedName" }}`) or by its relative path if not explicitly defined (`{{ template "domains/news/postPage.gohtml" }}`).
3.  Ensure the template renders properly by checking the UI or running regression tests.

## Regression Testing

To prevent unexpected changes when modifying core layouts or components, we have a test suite located in `core/templates/templates_regression_test.go`. It renders select core pages against stored "golden" files to catch visual or structural HTML regressions.

**Running the comparison tests:**
```bash
go test -v ./core/templates/... -run TestTemplateRegression
```
