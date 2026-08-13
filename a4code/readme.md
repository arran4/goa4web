# a4code

## Purpose

Package `a4code` is the root package for the custom A4Code markup engine. It defines the core parser, tokenization, and entry points for evaluating A4Code strings.

## Context and Use Cases (How and Why)

**Why it exists:** Goa4Web uses a custom, lightweight markup language (A4Code) rather than allowing raw HTML to ensure strict security and prevent XSS, while offering a simpler syntax than full Markdown for core forum features.
**What this allows:** It allows users to safely format their posts (bold, italic, images, links) without exposing the platform to malicious payloads.
**How to use it:** Call `a4code.Parse()` on a raw user string. If successful, you receive an Abstract Syntax Tree (AST) that can be passed to various renderers (HTML, Text, Format).

## Structure and Components

The primary files and their general responsibilities include:

- `common.go`
- `parser_test.go`
- `snip.go`
- `snip_test.go`
- `substring_test.go`
- `a4code.go`
- `html.go`
- `output.go`
- `parser.go`
- `quote.go`
- `quote_test.go`
- `sanitize.go`
- `substring.go`

### Exported Types and Interfaces

- **`RestrictedQuoteDepth`**:
- **`TruncatedQuoteDepth`**:
- **`ScannerInterface`** (Interface): Defines a core contract for this module.
- **`StreamOption`**:
- **`QuoteOption`**:

### Exported Functions

- `ConsumeCodeBlock`
- `GetNextArg`
- `GetNext`
- `TestParseToHTML`
- `TestParseImage`
- `TestRoundTrip`
- `TestParseNodes`
- `TestOffsets`
- `TestQuoteHTML`
- `TestQuoteOfHTML`
- `TestInlineCode`
- `TestBlockCode`
- `TestInlineCodeWithBrackets`
- `TestInlineQuote`
- `TestBlockQuote`
- `TestQuoteOfAlwaysBlock`
- `TestCodeIn`
- `TestCodeWhitespace`
- `TestCodeInGenerator`
- `TestCodeWithNestedQuote`
- `TestQOMarkup`
- `TestInvalidTags`
- `TestUpdateBlockStatus`
- `TestQuoteAdjacentLinkBoundaries`
- `TestCodeBlockEscaping`
- `TestLinkIsImmediateClose`
- `TestToText_Code`
- `TestBlockInlineInteractions`
- `TestTxtarBlockInline`
- `Snip`
- `SnipText`
- `SnipWords`
- `SnipTextWords`
- `TestSnip`
- `TestSnipText`
- `TestSubstringIncludesSelectedImage`
- `TestSubstringIncludesImageBetweenText`
- `ToA4Code`
- `ToHTML`
- `ToCode`
- `ToCleanText`
- `ToText`
- `WithDepth`
- `WithAllNodes`
- `Stream`
- `Parse`
- `ParseString`
- `ParseNodesReader`
- `ParseNodes`
- `WithParagraphQuote`
- `WithTrimSpace`
- `WithRestrictedQuoteDepth`
- `WithTruncatedQuoteDepth`
- `WithFullQuote`
- `QuoteText`
- `IsQuoteBlock`
- `QuoteReduce`
- `TestQuote`
- `TestQuoteFullParagraphs`
- `TestQuoteFullEscaping`
- `TestQuoteFullImage`
- `TestQuoteTrim`
- `TestQuoteOfWithSpaces`
- `TestQuoteOfWithSpacesAndEscapedQuote`
- `TestQuoteRoundTripComplexName`
- `TestSubstring`
- `TestQuoteDepthOptions`
- `TestQuoteTxtar`
- `TestIsPureQuote`
- `SanitizeURL`
- `Substring`

## Usage Examples

The typical workflow involves parsing an input string into an AST, then handing that AST off to a renderer (like `a4code2html` or `markdown`).

```go
import "goa4web/a4code"

// 1. Parse raw input string into an AST
astRoot, err := a4code.Parse("Some [b]input[/b] text")
if err != nil {
    // handle parser errors
}

// 2. The AST is now ready to be traversed or rendered.
```

## Limitations and Constraints

- **Internal Dependencies**: Specific limitations depend on the internal implementations of the exposed functions. Agents should not modify core interfaces without strictly considering downstream dependencies within the Goa4Web repository.
