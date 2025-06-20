package main

import (
	"context"
	"log"
)

type subscriberLister interface {
	ListSubscribersForThread(ctx context.Context, arg ListSubscribersForThreadParams) ([]*ListSubscribersForThreadRow, error)
}

func queueThreadNotifications(ctx context.Context, q subscriberLister, threadID, excludeUser int32, url string) {
	rows, err := q.ListSubscribersForThread(ctx, ListSubscribersForThreadParams{
		ForumthreadIdforumthread: threadID,
		Idusers:                  excludeUser,
	})
	if err != nil {
		log.Printf("Error: listSubscribersForThread: %s", err)
		return
	}
	for _, row := range rows {
		enqueueEmail(row.Username.String, url)
	}
}
