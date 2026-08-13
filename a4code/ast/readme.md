# A4Code abstract syntax tree

## Role in the pipeline

This package is the stable interchange between parsing and output generation.
`Node` provides parent links, source positions, and recursive transformation.
`Container` identifies nodes with children; `Block`, `Inline`, and their
blockable/inlinable variants express layout behavior used by parsers and renderers.
`Generator` is the complete visitor contract implemented by every output package.

## Node families

- Containers: `Root`, `Bold`, `Italic`, `Underline`, `Sup`, `Sub`, `Link`,
  `Quote`, `QuoteOf`, `Spoiler`, `Indent`, and `Custom`.
- Leaves: `Text`, `Image`, `Code`, `CodeIn`, and `HR`.
- All concrete nodes satisfy `Node`; container nodes additionally satisfy
  `Container` through `AddChild` and `GetChildren`. Compile-time behavior markers
  (`Block`, `Inline`, `BlockWithInlinable`, `InlineWithBlockable`) should be used
  instead of duplicating type lists in callers.

## Walking and changing trees

Prefer `Walk` when observing all nodes and `Transform` when replacing them:

```go
err := ast.Walk(root, func(n ast.Node) error {
    if container, ok := n.(ast.Container); ok {
        for _, child := range container.GetChildren() { _ = child }
    }
    return nil
})
```

`AddChild` establishes the parent pointer; directly appending to a `Children`
field can violate that invariant. Preserve `GetPos`/`SetPos` values when replacing
nodes because quoting and source-aware HTML rely on offsets. To add a node, update
`Node`, the appropriate behavior interfaces, `Generator`, `Generate`, every
concrete generator, walker/transform tests, and parser tests together.
