package repos

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
	"tomerab.com/cam-hub/internal/api/v1/models"
)

func makeCreds(uuid string) *models.CameraCreds {
	return &models.CameraCreds{
		UUID:     uuid,
		Username: "test-user",
		Password: "test-pass",
	}
}

func setupCameraCredsRepoTest(t *testing.T) (*PgxCameraCredsRepo, pgxmock.PgxPoolIface, context.Context) {
	t.Helper()

	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock pool: %v", err)
	}

	t.Cleanup(func() {
		mock.Close()
	})

	repo := NewPgxCameraCredsRepo(mock)
	ctx := context.Background()

	return repo, mock, ctx
}

func TestInsertCreds(t *testing.T) {
	repo, mock, ctx := setupCameraCredsRepoTest(t)

	t.Run("non existent - should inset", func(t *testing.T) {
		creds := makeCreds("1")

		mock.ExpectBegin()
		mock.ExpectExec(`.*INSERT INTO camera_creds.*\s*VALUES \(\$1,\$2,\$3\)\s*ON CONFLICT \(id\) DO UPDATE SET(\s*.*)*`).
			WithArgs(creds.UUID, creds.Username, creds.Password).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		tx, err := mock.Begin(ctx)
		if err != nil {
			t.Fatalf("failed to create a transaction: %v", err)
		}
		defer tx.Rollback(ctx)

		if err := repo.InsertCreds(ctx, tx, creds); err != nil {
			t.Fatalf("expected nil, got err: %v", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet expectations: %v", err)
		}
	})

	t.Run("update row - should update", func(t *testing.T) {
		creds := makeCreds("1")

		mock.ExpectBegin()
		mock.ExpectExec(`.*INSERT INTO camera_creds.*\s*VALUES \(\$1,\$2,\$3\)\s*ON CONFLICT \(id\) DO UPDATE SET(\s*.*)*`).
			WithArgs(creds.UUID, creds.Username, creds.Password).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		tx, err := mock.Begin(ctx)
		if err != nil {
			t.Fatalf("failed to create a transaction: %v", err)
		}
		defer tx.Rollback(ctx)

		if err := repo.InsertCreds(ctx, tx, creds); err != nil {
			t.Fatalf("expected nil, got err: %v", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet expectations: %v", err)
		}
	})

	t.Run("pass nil creds - should panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic when camera is nil, got no panic")
			}
		}()

		mock.ExpectBegin()
		tx, err := mock.Begin(ctx)
		if err != nil {
			t.Fatalf("failed to create a transaction: %v", err)
		}
		defer tx.Rollback(ctx)

		err = repo.InsertCreds(ctx, tx, nil)
		t.Errorf("should have panicked before returning error, got: %v", err)
	})

	t.Run("upsert failed - should return error", func(t *testing.T) {
		creds := makeCreds("1")

		expectedMsg := "failed to insert/update creds"
		mock.ExpectBegin()
		mock.ExpectExec(`.*INSERT INTO camera_creds.*\s*VALUES \(\$1,\$2,\$3\)\s*ON CONFLICT \(id\) DO UPDATE SET(\s*.*)*`).
			WithArgs(creds.UUID, creds.Username, creds.Password).
			WillReturnError(errors.New("failed to insert/update creds"))

		tx, err := mock.Begin(ctx)
		if err != nil {
			t.Fatalf("failed to create a transaction: %v", err)
		}
		defer tx.Rollback(ctx)

		err = repo.InsertCreds(ctx, tx, creds)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !strings.Contains(err.Error(), expectedMsg) {
			t.Fatalf("expected error: %s to contain: %s", err, expectedMsg)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet expectations: %v", err)
		}
	})
}

func TestCredsFindOne(t *testing.T) {
	repo, mock, ctx := setupCameraCredsRepoTest(t)

	t.Run("uuid exist - should return creds", func(t *testing.T) {
		expected := makeCreds("1")

		mock.ExpectQuery(`SELECT.*\s*FROM\s*camera_creds.*\s*WHERE id = \$1`).
			WithArgs(expected.UUID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "username", "password"}).
				AddRow(expected.UUID, expected.Username, expected.Password))

		got, err := repo.FindOne(ctx, "1")
		if err != nil {
			t.Fatalf("expected nil, got err: %v", err)
		}

		if !reflect.DeepEqual(expected, got) {
			t.Errorf("expected: %v, got: %v", expected, got)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet expectations: %v", err)
		}
	})

	t.Run("uuid does not exist - should return err", func(t *testing.T) {
		expected := makeCreds("1")

		mock.ExpectQuery(`SELECT.*\s*FROM\s*camera_creds.*\s*WHERE id = \$1`).
			WithArgs(expected.UUID).
			WillReturnError(pgx.ErrNoRows)

		_, err := repo.FindOne(ctx, "1")
		if err == nil {
			t.Fatalf("expected err, got nil")
		}

		if !strings.Contains(err.Error(), "no rows in result set") {
			t.Errorf("expected pgx.ErrNoRows, got: %v", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet expectations: %v", err)
		}
	})
}
