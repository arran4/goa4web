# a4code

## Purpose

Package `a4code` is the root package for the custom A4Code markup engine. It defines the core parser, tokenization, and entry points for evaluating A4Code strings.

## Structure and Components

The primary files and their general responsibilities include:

- `a4code.go`
- `quote_test.go`
- `snip_test.go`
- `substring.go`
- `substring_test.go`
- `common.go`
- `html.go`
- `output.go`
- `parser.go`
- `parser_test.go`
- `quote.go`
- `sanitize.go`
- `snip.go`

### Exported Types

- `ScannerInterface`
- `StreamOption`
- `QuoteOption`
- `RestrictedQuoteDepth`
- `TruncatedQuoteDepth`

### Exported Functions

- `ToA4Code`
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
- `ConsumeCodeBlock`
- `GetNextArg`
- `GetNext`
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
- `WithParagraphQuote`
- `WithTrimSpace`
- `WithRestrictedQuoteDepth`
- `WithTruncatedQuoteDepth`
- `WithFullQuote`
- `QuoteText`
- `IsQuoteBlock`
- `QuoteReduce`
- `SanitizeURL`
- `Snip`
- `SnipText`
- `SnipWords`
- `SnipTextWords`

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
