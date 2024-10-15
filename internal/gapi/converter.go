package gapi

import (
	"github.com/RobinHood3082/simplebank/internal/pb"
	"github.com/RobinHood3082/simplebank/internal/persistence"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func convertUser(user persistence.User) *pb.User {
	return &pb.User{
		Username:          user.Username,
		FullName:          user.FullName,
		Email:             user.Email,
		PasswordChangedAt: timestamppb.New(user.PasswordChangedAt.Time),
		CreatedAt:         timestamppb.New(user.CreatedAt.Time),
	}
}

func convertAccount(account persistence.Account) *pb.Account {
	return &pb.Account{
		Id:        account.ID,
		Owner:     account.Owner,
		Balance:   account.Balance,
		Currency:  account.Currency,
		CreatedAt: timestamppb.New(account.CreatedAt.Time),
	}
}
