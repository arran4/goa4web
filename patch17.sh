sed -i 's/if t, ok := child.(\*ast.Text); ok \&\& strings.TrimSpace(t.Value) != "" {/if strings.TrimSpace(t.Value) != "" {/g' a4code/quote.go
