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

func makeCamera(uuid string, isPaired bool) *models.Camera {
	return &models.Camera{
		UUID:            uuid,
		CameraName:      "Test Camera",
		Manufacturer:    "Test Manufacturer",
		Model:           "Test Model",
		SerialNumber:    "SN123",
		HardwareId:      "HW456",
		Addr:            "192.168.1.100",
		IsPaired:        isPaired,
		FirmwareVersion: "1.0.0",
	}
}

func setupCameraRepoTest(t *testing.T) (*PgxCameraRepo, pgxmock.PgxPoolIface, context.Context) {
	t.Helper()

	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock pool: %v", err)
	}

	t.Cleanup(func() {
		mock.Close()
	})

	repo := NewPgxCameraRepo(mock)
	ctx := context.Background()

	return repo, mock, ctx
}

func TestUpsertCamera(t *testing.T) {
	repo, mock, ctx := setupCameraRepoTest(t)

	t.Run("nil camera should return error", func(t *testing.T) {
		mock.ExpectBegin()
		tx, err := mock.Begin(ctx)
		if err != nil {
			t.Fatalf("failed to create transaction: %v", err)
		}
		defer tx.Rollback(ctx)

		err = repo.UpsertCameraTx(ctx, tx, nil)
		if err == nil {
			t.Error("expected error for nil camera, got nil")
		}
	})

	t.Run("valid camera should succeed", func(t *testing.T) {
		camera := makeCamera("1", true)
		mock.ExpectBegin()

		mock.ExpectExec(`INSERT INTO cameras`).
			WithArgs(
				camera.UUID,
				camera.CameraName,
				camera.Manufacturer,
				camera.Model,
				camera.FirmwareVersion,
				camera.SerialNumber,
				camera.HardwareId,
				camera.Addr,
				camera.IsPaired,
			).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		tx, err := mock.Begin(ctx)
		if err != nil {
			t.Fatalf("failed to create transaction: %v", err)
		}
		defer tx.Rollback(ctx)

		err = repo.UpsertCameraTx(ctx, tx, camera)
		if err != nil {
			t.Errorf("upsertCamera failed: %v", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unfulfilled expectations: %v", err)
		}
	})

	t.Run("database error should be returned", func(t *testing.T) {
		camera := makeCamera("1", true)

		mock.ExpectBegin()
		mock.ExpectExec(`INSERT INTO cameras`).
			WithArgs(
				camera.UUID,
				camera.CameraName,
				camera.Manufacturer,
				camera.Model,
				camera.FirmwareVersion,
				camera.SerialNumber,
				camera.HardwareId,
				camera.Addr,
				camera.IsPaired,
			).
			WillReturnError(errors.New("database connection failed"))

		tx, err := mock.Begin(ctx)
		if err != nil {
			t.Fatalf("failed to create transaction: %v", err)
		}
		defer tx.Rollback(ctx)

		err = repo.UpsertCameraTx(ctx, tx, camera)
		if err == nil {
			t.Error("expected database error, got nil")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unfulfilled expectations: %v", err)
		}
	})
}

func TestFindExistingPaired(t *testing.T) {
	repo, mock, ctx := setupCameraRepoTest(t)

	t.Run("empty should return empty", func(t *testing.T) {
		filtered, err := repo.FindExistingPaired(ctx, []string{})
		if err != nil {
			t.Errorf("expected nil, got err: %v", err)
		}
		if len(filtered) != 0 {
			t.Errorf("expected empty slice, got: %v", filtered)
		}
	})

	t.Run("query returns row - should return true", func(t *testing.T) {
		uuids := []string{"1"}
		b := mock.ExpectBatch()
		b.ExpectQuery(`SELECT id\s+FROM cameras\s+WHERE id = \$1 and ispaired = true`).
			WithArgs("1").
			WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow("1"))

		got, err := repo.FindExistingPaired(ctx, uuids)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expected := []bool{true}
		if !reflect.DeepEqual(got, expected) {
			t.Errorf("got: %v, want: %v", got, expected)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet expectations: %v", err)
		}
	})

	t.Run("query returns no rows - should return false", func(t *testing.T) {
		uuids := []string{"1"}
		b := mock.ExpectBatch()
		b.ExpectQuery(`SELECT id\s+FROM cameras\s+WHERE id = \$1 and ispaired = true`).
			WithArgs("1").
			WillReturnError(pgx.ErrNoRows)

		got, err := repo.FindExistingPaired(ctx, uuids)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expected := []bool{false}
		if !reflect.DeepEqual(got, expected) {
			t.Errorf("got: %v, want: %v", got, expected)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet expectations: %v", err)
		}
	})

	t.Run("mixed results - some true others false", func(t *testing.T) {
		uuids := []string{"exists", "missing", "exists-too"}

		b := mock.ExpectBatch()
		b.ExpectQuery(`SELECT id\s+FROM cameras\s+WHERE id = \$1 and ispaired = true`).
			WithArgs("exists").
			WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow("exists"))

		b.ExpectQuery(`SELECT id\s+FROM cameras\s+WHERE id = \$1 and ispaired = true`).
			WithArgs("missing").
			WillReturnError(pgx.ErrNoRows)

		b.ExpectQuery(`SELECT id\s+FROM cameras\s+WHERE id = \$1 and ispaired = true`).
			WithArgs("exists-too").
			WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow("exists-too"))

		results, err := repo.FindExistingPaired(ctx, uuids)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expected := []bool{true, false, true}
		if !reflect.DeepEqual(results, expected) {
			t.Errorf("got %v, want %v", results, expected)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet expectations: %v", err)
		}
	})

	t.Run("database error should be returned", func(t *testing.T) {
		uuids := []string{"camera-1"}

		b := mock.ExpectBatch()
		b.ExpectQuery(`SELECT id\s+FROM cameras\s+WHERE id = \$1 and ispaired = true`).
			WithArgs("camera-1").
			WillReturnError(errors.New("database connection failed"))

		_, err := repo.FindExistingPaired(ctx, uuids)
		if err == nil {
			t.Error("expected error, got nil")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet expectations: %v", err)
		}
	})

	t.Run("nil input should handle gracefully", func(t *testing.T) {
		results, err := repo.FindExistingPaired(ctx, nil)
		if err != nil {
			t.Errorf("expected nil error for nil input, got: %v", err)
		}
		if len(results) != 0 {
			t.Errorf("expected empty results for nil input, got: %v", results)
		}
	})
}

func TestCamerasFindOne(t *testing.T) {
	repo, mock, ctx := setupCameraRepoTest(t)

	t.Run("camera exists", func(t *testing.T) {
		cam := makeCamera("1", true)
		mock.ExpectQuery(`SELECT.*FROM cameras.*WHERE id = \$1`).
			WithArgs(cam.UUID).
			WillReturnRows(pgxmock.NewRows([]string{
				"id",
				"name",
				"manufacturer",
				"model",
				"firmwareversion",
				"serialnumber",
				"hardwareid",
				"addr",
				"ispaired"}).AddRow(cam.UUID,
				cam.CameraName,
				cam.Manufacturer,
				cam.Model,
				cam.FirmwareVersion,
				cam.SerialNumber,
				cam.HardwareId,
				cam.Addr,
				cam.IsPaired))

		got, err := repo.FindOne(ctx, "1")
		if err != nil {
			t.Fatalf("expected nil, got error:  %v", err)
		}

		if !reflect.DeepEqual(cam, got) {
			t.Errorf("expected: %v, got: %v", cam, got)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet expectations: %v", err)
		}
	})

	t.Run("camera doesnt exist", func(t *testing.T) {
		mock.ExpectQuery(`SELECT.*FROM cameras.*WHERE id = \$1`).
			WithArgs("none").
			WillReturnError(pgx.ErrNoRows)

		got, err := repo.FindOne(ctx, "none")
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !strings.Contains(err.Error(), "no rows in result set") {
			t.Errorf("expected 'no rows' error, got: %v", err)
		}

		if !reflect.DeepEqual(&models.Camera{}, got) {
			t.Errorf("expected empty camera, got: %v", got)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet expectations: %v", err)
		}
	})

	t.Run("empty string", func(t *testing.T) {
		mock.ExpectQuery(`SELECT.*FROM cameras.*WHERE id = \$1`).
			WithArgs("").
			WillReturnError(pgx.ErrNoRows)

		got, err := repo.FindOne(ctx, "")
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !strings.Contains(err.Error(), "no rows in result set") {
			t.Errorf("expected 'no rows' error, got: %v", err)
		}

		if !reflect.DeepEqual(&models.Camera{}, got) {
			t.Errorf("expected empty camera, got: %v", got)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet expectations: %v", err)
		}
	})
}

func TestSave(t *testing.T) {
	repo, mock, ctx := setupCameraRepoTest(t)

	t.Run("successful save", func(t *testing.T) {
		camera := makeCamera("1", false)

		mock.ExpectExec(`UPDATE cameras\s+SET.*WHERE id = \$1`).
			WithArgs(
				camera.UUID,
				camera.CameraName,
				camera.Manufacturer,
				camera.Model,
				camera.FirmwareVersion,
				camera.SerialNumber,
				camera.HardwareId,
				camera.Addr,
				camera.IsPaired,
			).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		err := repo.Save(ctx, camera)
		if err != nil {
			t.Errorf("expected nil error, got: %v", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet expectations: %v", err)
		}
	})

	t.Run("camera not found - no rows affected", func(t *testing.T) {
		camera := makeCamera("1", false)

		mock.ExpectExec(`UPDATE cameras\s+SET.*WHERE id = \$1`).
			WithArgs(
				camera.UUID,
				camera.CameraName,
				camera.Manufacturer,
				camera.Model,
				camera.FirmwareVersion,
				camera.SerialNumber,
				camera.HardwareId,
				camera.Addr,
				camera.IsPaired,
			).
			WillReturnResult(pgxmock.NewResult("UPDATE", 0))

		err := repo.Save(ctx, camera)
		if err == nil {
			t.Fatal("expected error when no rows affected, got nil")
		}

		expectedErrMsg := "save failed: no rows were affected"
		if !strings.Contains(err.Error(), expectedErrMsg) {
			t.Errorf("expected error containing '%s', got: %v", expectedErrMsg, err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet expectations: %v", err)
		}
	})

	t.Run("nil camera", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic when camera is nil, got no panic")
			}
		}()

		err := repo.Save(ctx, nil)
		t.Errorf("should have panicked before returning error, got: %v", err)
	})

	t.Run("multiple rows affected - unexpected", func(t *testing.T) {
		camera := makeCamera("1", false)

		mock.ExpectExec(`UPDATE cameras\s+SET.*WHERE id = \$1`).
			WithArgs(
				camera.UUID,
				camera.CameraName,
				camera.Manufacturer,
				camera.Model,
				camera.FirmwareVersion,
				camera.SerialNumber,
				camera.HardwareId,
				camera.Addr,
				camera.IsPaired,
			).
			WillReturnResult(pgxmock.NewResult("UPDATE", 2))

		err := repo.Save(ctx, camera)
		if err == nil {
			t.Fatal("expected error when multiple rows affected, got nil")
		}

		expectedErrMsg := "save failed: no rows were affected"
		if !strings.Contains(err.Error(), expectedErrMsg) {
			t.Errorf("expected error containing '%s', got: %v", expectedErrMsg, err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet expectations: %v", err)
		}
	})
}
