package main

import (
	"context"
	"math/rand"
)

const set = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func generateSID(ctx context.Context, q Queries) (string, error) {
	setSize := len(set)
	var sid [64]byte
	for i := 0; i < 63; i++ {
		sid[i] = set[rand.Intn(setSize)]
	}
	sid[63] = 0
	for {
		r, err := q.SIDExpired(ctx, sid)
		if err != nil {
			return "", err
		}
		for i := 0; i < 63; i++ {
			sid[i] = a.set[rand.Intn(setSize)]
		}
	}
	return string(sid[:])
}
