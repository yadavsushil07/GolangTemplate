package repository_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yadavsushil07/GolangTemplate/internal/repository"
	"github.com/yadavsushil07/GolangTemplate/internal/testutil"
)

func TestUserRepository(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	repo := repository.NewUserRepository(tdb.Pool)
	ctx := context.Background()

	t.Run("Create and FindByIdentifier", func(t *testing.T) {
		tdb.TruncateTables(t, "users")

		user, err := repo.Create(ctx, "user@example.com", "customer")
		require.NoError(t, err)
		require.NotNil(t, user)
		assert.Equal(t, "user@example.com", user.Identifier)
		assert.Equal(t, "customer", user.Role)
		assert.NotZero(t, user.ID)

		found, err := repo.FindByIdentifier(ctx, "user@example.com")
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, user.ID, found.ID)
	})

	t.Run("FindByIdentifier returns nil for unknown", func(t *testing.T) {
		found, err := repo.FindByIdentifier(ctx, "nobody@example.com")
		require.NoError(t, err)
		assert.Nil(t, found)
	})

	t.Run("FindByID", func(t *testing.T) {
		tdb.TruncateTables(t, "users")
		user, err := repo.Create(ctx, "findme@example.com", "vendor")
		require.NoError(t, err)

		found, err := repo.FindByID(ctx, user.ID)
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, "vendor", found.Role)
	})

	t.Run("FindByID returns nil for unknown", func(t *testing.T) {
		found, err := repo.FindByID(ctx, 99999)
		require.NoError(t, err)
		assert.Nil(t, found)
	})
}
