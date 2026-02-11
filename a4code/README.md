# a4code

a4code is a custom markup language that uses a Lisp-style syntax with square brackets.

## Syntax

The general syntax is `[tag argument content]`.

*   **No closing tags:** Unlike HTML or BBCode, a4code does not use closing tags like `[/tag]`. Instead, the scope of a tag is determined by the nesting of brackets `[...]`.
*   **Arguments:** Tags can take arguments separated by spaces.
*   **Content:** The content follows the arguments and is closed by the matching `]`.

### Examples

*   **Link:** `[link http://example.com Link Title]`
*   **Bold:** `[b This is bold]`
*   **Quote:** `[quote This is a quote]`
*   **Nested:** `[quote [b Bold inside quote]]`

## Forbidden Syntax

*   Closing tags (e.g., `[/b]`, `[/link]`) are **invalid** and will trigger a parser error.
*   Assignment style (e.g., `[link=http://example.com]`) is deprecated/incorrect for this parser version in favor of space-separated arguments.

## Code Blocks

Code blocks are an exception and may use specific delimiters defined by the parser, but standard usage should follow the bracket structure where possible or use the specific `[code]...[/code]` legacy support if enabled, though `[code ... ]` is preferred if content permits.
