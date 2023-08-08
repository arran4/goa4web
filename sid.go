package main

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"time"
)

const set = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func generateSID(ctx context.Context, q Queries) (string, error) {
	setSize := len(set)
	var sid [64]byte
	for i := 0; i < 64; i++ {
		sid[i] = set[rand.Intn(setSize)]
	}
	for {
		r, err := q.SIDExpired(ctx, sql.NullString{
			String: string(sid[:]),
			Valid:  true,
		})
		if err != nil {
			return "", fmt.Errorf("generateSID: %w", err)
		}
		if r == nil || !r.Logintime.Valid || r.Logintime.Time.Before(time.Now()) {
			break
		}
		for i := 0; i < 64; i++ {
			sid[i] = set[rand.Intn(setSize)]
		}
	}
	return string(sid[:]), nil
}
