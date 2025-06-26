package db

// DB exposes the underlying database handle.
func (q *Queries) DB() DBTX { return q.db }
