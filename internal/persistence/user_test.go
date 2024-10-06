package persistence

import (
	"context"
	"testing"
	"time"

	"github.com/RobinHood3082/simplebank/util"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

func createRandomUser(t *testing.T) User {
	hashedPassword, err := util.HashPassword(util.RandomString(6))
	require.NoError(t, err)

	arg := CreateUserParams{
		Username:       util.RandomOwner(),
		HashedPassword: hashedPassword,
		FullName:       util.RandomOwner(),
		Email:          util.RandomEmail(),
	}

	user, err := testQueries.CreateUser(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, user)

	require.Equal(t, arg.Username, user.Username)
	require.Equal(t, arg.HashedPassword, user.HashedPassword)
	require.Equal(t, arg.FullName, user.FullName)
	require.Equal(t, arg.Email, user.Email)

	require.True(t, user.PasswordChangedAt.Time.IsZero())
	require.NotZero(t, user.CreatedAt)

	return user
}

func TestCreateUser(t *testing.T) {
	createRandomUser(t)
}

func TestGetUser(t *testing.T) {
	user1 := createRandomUser(t)
	user2, err := testQueries.GetUser(context.Background(), user1.Username)
	require.NoError(t, err)
	require.NotEmpty(t, user2)

	require.Equal(t, user1.Username, user2.Username)
	require.Equal(t, user1.HashedPassword, user2.HashedPassword)
	require.Equal(t, user1.Email, user2.Email)
	require.Equal(t, user1.FullName, user2.FullName)
	require.WithinDuration(t, user1.CreatedAt.Time, user2.CreatedAt.Time, time.Second)
	require.WithinDuration(t, user1.PasswordChangedAt.Time, user2.PasswordChangedAt.Time, time.Second)
}

func TestUpdateUserOnlyFullName(t *testing.T) {
	oldUser := createRandomUser(t)

	newFullName := util.RandomOwner()
	updatedUser, err := testQueries.UpdateUser(context.Background(),
		UpdateUserParams{
			Username: oldUser.Username,
			FullName: pgtype.Text{
				String: newFullName,
				Valid:  true,
			},
		},
	)

	require.NoError(t, err)
	require.NotEmpty(t, updatedUser)

	require.Equal(t, oldUser.Username, updatedUser.Username)
	require.Equal(t, oldUser.HashedPassword, updatedUser.HashedPassword)
	require.Equal(t, oldUser.Email, updatedUser.Email)
	require.Equal(t, newFullName, updatedUser.FullName)
	require.WithinDuration(t, oldUser.CreatedAt.Time, updatedUser.CreatedAt.Time, time.Second)
	require.WithinDuration(t, oldUser.PasswordChangedAt.Time, updatedUser.PasswordChangedAt.Time, time.Second)

	require.NotEqual(t, oldUser.FullName, updatedUser.FullName)
}

func TestUpdateUserOnlyEmail(t *testing.T) {
	oldUser := createRandomUser(t)

	newEmail := util.RandomEmail()
	updatedUser, err := testQueries.UpdateUser(context.Background(),
		UpdateUserParams{
			Username: oldUser.Username,
			Email: pgtype.Text{
				String: newEmail,
				Valid:  true,
			},
		},
	)

	require.NoError(t, err)
	require.NotEmpty(t, updatedUser)

	require.Equal(t, oldUser.Username, updatedUser.Username)
	require.Equal(t, oldUser.HashedPassword, updatedUser.HashedPassword)
	require.Equal(t, oldUser.FullName, updatedUser.FullName)
	require.Equal(t, newEmail, updatedUser.Email)
	require.WithinDuration(t, oldUser.CreatedAt.Time, updatedUser.CreatedAt.Time, time.Second)
	require.WithinDuration(t, oldUser.PasswordChangedAt.Time, updatedUser.PasswordChangedAt.Time, time.Second)

	require.NotEqual(t, oldUser.Email, updatedUser.Email)
}

func TestUpdateUserOnlyPassword(t *testing.T) {
	oldUser := createRandomUser(t)

	newPassword := util.RandomString(6)
	newHashedPassword, err := util.HashPassword(newPassword)
	require.NoError(t, err)

	updatedUser, err := testQueries.UpdateUser(context.Background(),
		UpdateUserParams{
			Username: oldUser.Username,
			HashedPassword: pgtype.Text{
				String: newHashedPassword,
				Valid:  true,
			},
		},
	)

	require.NoError(t, err)
	require.NotEmpty(t, updatedUser)

	require.Equal(t, oldUser.Username, updatedUser.Username)
	require.Equal(t, updatedUser.HashedPassword, newHashedPassword)
	require.Equal(t, oldUser.FullName, updatedUser.FullName)
	require.Equal(t, oldUser.Email, updatedUser.Email)
	require.WithinDuration(t, oldUser.CreatedAt.Time, updatedUser.CreatedAt.Time, time.Second)

	require.NotEqual(t, oldUser.HashedPassword, updatedUser.HashedPassword)
}

func TestUpdateUserAllFields(t *testing.T) {
	oldUser := createRandomUser(t)

	newFullName := util.RandomOwner()
	newEmail := util.RandomEmail()
	newPassword := util.RandomString(6)
	newHashedPassword, err := util.HashPassword(newPassword)
	require.NoError(t, err)

	updatedUser, err := testQueries.UpdateUser(context.Background(),
		UpdateUserParams{
			Username: oldUser.Username,
			FullName: pgtype.Text{
				String: newFullName,
				Valid:  true,
			},
			Email: pgtype.Text{
				String: newEmail,
				Valid:  true,
			},
			HashedPassword: pgtype.Text{
				String: newHashedPassword,
				Valid:  true,
			},
		},
	)

	require.NoError(t, err)
	require.NotEmpty(t, updatedUser)

	require.Equal(t, oldUser.Username, updatedUser.Username)
	require.Equal(t, newHashedPassword, updatedUser.HashedPassword)
	require.Equal(t, newFullName, updatedUser.FullName)
	require.Equal(t, newEmail, updatedUser.Email)
	require.WithinDuration(t, oldUser.CreatedAt.Time, updatedUser.CreatedAt.Time, time.Second)

	require.NotEqual(t, oldUser.FullName, updatedUser.FullName)
	require.NotEqual(t, oldUser.Email, updatedUser.Email)
	require.NotEqual(t, oldUser.HashedPassword, updatedUser.HashedPassword)
}
