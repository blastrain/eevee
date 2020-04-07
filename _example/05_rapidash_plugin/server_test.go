package main

import (
	"context"
	"daoplugin/mock/model"
	"daoplugin/mock/repository"
	"testing"
)

func TestUserGet(t *testing.T) {
	repo := repository.NewMock()
	ctx := context.Background()
	repo.User().EXPECT().FindByID(ctx, 1).Return(model.DefaultUser(), nil)
	u, err := repo.User().FindByID(ctx, 1)
	if err != nil {
		t.Fatalf("%+v", err)
	}
	if u.ID != model.DefaultUser().ID {
		t.Fatal("invalid user")
	}
}
