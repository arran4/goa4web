package forum

import "sort"

// Label represents a public, author, or private label.
type Label struct {
	Text string // The label value.
	Type string // Label type: "public", "author", or "private".
}

// mergeLabels combines public, author, and private labels into one slice sorted by label text.
func mergeLabels(pub, author, priv []string) []Label {
	labels := make([]Label, 0, len(pub)+len(author)+len(priv))
	for _, l := range pub {
		labels = append(labels, Label{Text: l, Type: "public"})
	}
	for _, l := range author {
		labels = append(labels, Label{Text: l, Type: "author"})
	}
	for _, l := range priv {
		labels = append(labels, Label{Text: l, Type: "private"})
	}
	sort.Slice(labels, func(i, j int) bool { return labels[i].Text < labels[j].Text })
	return labels
}
