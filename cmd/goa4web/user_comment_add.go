package main

import (
	"flag"
)

// userCommentAddCmd implements "user comment add".
type userCommentAddCmd struct {
	*userCommentCmd
	fs      *flag.FlagSet
	request adminUserCommentAddRequest
}

func parseUserCommentAddCmd(parent *userCommentCmd, args []string) (*userCommentAddCmd, error) {
	c := &userCommentAddCmd{userCommentCmd: parent}
	if err := parseAdminUserCommentAddFlags("add", args, &c.request, &c.fs); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *userCommentAddCmd) Run() error {
	return runAdminUserCommentAdd(c.rootCmd, c.request)
}
