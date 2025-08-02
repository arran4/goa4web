package db

import "context"

type CustomQueries interface {
	SearchBloggers(ctx context.Context, arg SearchBloggersParams) ([]*BloggerCountRow, error)
	ListBloggers(ctx context.Context, arg ListBloggersParams) ([]*BloggerCountRow, error)
	SearchWriters(ctx context.Context, arg SearchWritersParams) ([]*WriterCountRow, error)
	ListWriters(ctx context.Context, arg ListWritersParams) ([]*WriterCountRow, error)
}
