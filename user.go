package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"
)

func UserAdderMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		// Get the session.
		session, err := store.Get(request, sessionName)
		if err != nil {
			http.Error(writer, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		// TODO login user
		ctx := context.WithValue(request.Context(), ContextValues("session"), session)
		next.ServeHTTP(writer, request.WithContext(ctx))
	})
}

func loginUser(ctx context.Context, queries Queries, username, password string) (int32, error) {
	userID, err := queries.Login(ctx, LoginParams{
		Username: sql.NullString{
			String: username,
			Valid:  true,
		},
		MD5: password,
	})
	if err != nil {
		return -1, fmt.Errorf("loginUser: %w", err)
	}
	return userID, nil
}

func verifyUser(ctx context.Context, w http.ResponseWriter, queries Queries, sid string) (int, error) {
	if sid == "" {
		t := time.Now().AddDate(1, 0, 0)
		var err error
		sid, err = generateSID(ctx, queries)
		if err != nil {
			return 0, fmt.Errorf("verifyUser: %w", err)
		}
		http.SetCookie(w, &http.Cookie{
			Name:    "SID",
			Value:   sid,
			Expires: t,
		})
	}

	user, err := queries.SelectUserBySID(ctx, sql.NullString{
		String: sid,
		Valid:  true,
	})
	if err != nil {
		return 0, fmt.Errorf("verifyUser: %w", err)
	}

	if user.Logintime.Time.AddDate(1, 0, 0).After(time.Now()) {
		return int(user.UsersIdusers), nil
	}

	return 0, nil
}

func registerUser(ctx context.Context, queries Queries, username, password, email string) (int, error) {
	userID, err := queries.Login(ctx, LoginParams{
		Username: sql.NullString{
			String: username,
			Valid:  true,
		},
		MD5: password,
	})
	if userID != 0 || err == nil {
		return -3, fmt.Errorf("registerUser: %w", err)
	}

	result, err := queries.InsertUser(ctx, InsertUserParams{
		Username: sql.NullString{
			Valid:  true,
			String: username,
		},
		MD5: password,
		Email: sql.NullString{
			Valid:  true,
			String: email,
		},
	})
	if err != nil {
		return -3, err // SQL error.
	}

	lastInsertID, err := result.LastInsertId()
	if err != nil {
		return -3, fmt.Errorf("registerUser: %w", err) // SQL error.
	}

	return int(lastInsertID), nil
}
