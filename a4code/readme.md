# a4code

## Purpose

Package `a4code` is the root package for the custom A4Code markup engine. It defines the core parser, tokenization, and entry points for evaluating A4Code strings.

## Why It Exists

Goa4Web uses a custom, lightweight markup language (A4Code) rather than allowing raw HTML to ensure strict security and prevent XSS, while offering a simpler syntax than full Markdown for core forum features.

## What It Allows

It allows users to safely format their posts (bold, italic, images, links) without exposing the platform to malicious payloads. It acts as the gatekeeper for user content.

## Structure and Components

The primary files and their general responsibilities include:

- `output.go`
- `parser_test.go`
- `sanitize.go`
- `snip.go`
- `parser.go`
- `quote.go`
- `quote_test.go`
- `snip_test.go`
- `substring.go`
- `substring_test.go`
- `a4code.go`
- `common.go`
- `html.go`

### Exported Types and Interfaces

- **`TruncatedQuoteDepth`**:
- **`ScannerInterface`** (Interface): Defines a core contract for this module.
- **`StreamOption`**:
- **`QuoteOption`**:
- **`RestrictedQuoteDepth`**:

### Exported Functions

- `ToCode`
- `ToCleanText`
- `ToText`
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
- `SanitizeURL`
- `Snip`
- `SnipText`
- `SnipWords`
- `SnipTextWords`
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
- `TestSnip`
- `TestSnipText`
- `Substring`
- `TestSubstringIncludesSelectedImage`
- `TestSubstringIncludesImageBetweenText`
- `ToA4Code`
- `ConsumeCodeBlock`
- `GetNextArg`
- `GetNext`
- `ToHTML`

## Usage Examples

Call `a4code.Parse()` on a raw user string. If successful, you receive an Abstract Syntax Tree (AST) that can be passed to various renderers (HTML, Text, Format).

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
