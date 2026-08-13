# a4code

## Purpose

Package `a4code` is the root package for the custom A4Code markup engine. It defines the core parser, tokenization, and entry points for evaluating A4Code strings.

## Structure and Components

The primary files and their general responsibilities include:

- `sanitize.go`
- `snip.go`
- `substring.go`
- `substring_test.go`
- `common.go`
- `output.go`
- `parser.go`
- `quote.go`
- `snip_test.go`
- `a4code.go`
- `html.go`
- `parser_test.go`
- `quote_test.go`

### Exported Types and Interfaces

- **`ScannerInterface`** (Interface): Defines a core contract for this module.
- **`StreamOption`**:
- **`QuoteOption`**:
- **`RestrictedQuoteDepth`**:
- **`TruncatedQuoteDepth`**:

### Exported Functions

- `SanitizeURL`
- `Snip`
- `SnipText`
- `SnipWords`
- `SnipTextWords`
- `Substring`
- `TestSubstringIncludesSelectedImage`
- `TestSubstringIncludesImageBetweenText`
- `ConsumeCodeBlock`
- `GetNextArg`
- `GetNext`
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
- `TestSnip`
- `TestSnipText`
- `ToA4Code`
- `ToHTML`
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

## Usage

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
