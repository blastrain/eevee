package main

import (
	"context"
	"relation/mock/model/factory"
	"relation/mock/repository"
	"testing"
)

func TestUserGet(t *testing.T) {
	repo := repository.NewMock()
	ctx := context.Background()
	repo.UserMock().EXPECT().FindByID(ctx, 1).Return(factory.DefaultUser(), nil)
	u, err := repo.User().FindByID(ctx, 1)
	if err != nil {
		t.Fatalf("%+v", err)
	}
	if u.ID != factory.DefaultUser().ID {
		t.Fatal("invalid user")
	}
}
