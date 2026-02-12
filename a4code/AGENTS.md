# Guidelines for a4code

When working with `a4code`, adhere to the following rules:

1.  **Lisp-Style Syntax:** Always use `[tag arg content]` syntax.
    *   Correct: `[link http://example.com Title]`
    *   Incorrect: `[link=http://example.com]Title[/link]`

2.  **No Closing Tags:** Do not use `[/tag]`. The parser throws an error if it encounters a tag starting with `/`.

3.  **Testing:** When writing tests for `a4code`:
    *   Use the Lisp-style syntax.
    *   Verify structural properties (`IsBlock`, `IsImmediateClose`) based on this syntax.
    *   `[link url]` (no content) -> `IsImmediateClose = true`
    *   `[link url title]` (has content) -> `IsImmediateClose = false`
