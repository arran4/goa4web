sed -i '/newVals.Set("method", r.Method)/d' core/session.go
sed -i '/if r.Method != http.MethodGet {/,+2d' core/session.go
