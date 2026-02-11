package ast

// Walk traverses the node tree depth-first without modifying nodes.
func Walk(n Node, fn func(Node) error) error {
	if n == nil {
		return nil
	}
	if err := fn(n); err != nil {
		return err
	}
	if p, ok := n.(parent); ok {
		for _, c := range *p.childrenPtr() {
			if err := Walk(c, fn); err != nil {
				return err
			}
		}
	}
	return nil
}
