package v1

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/pashagolub/pgxmock/v4"
)

func TestPairingValidation(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock pool: %v", err)
	}
	defer mock.Close()

	svc := CameraService{
		DB:     mock,
		Logger: slog.New(slog.NewTextHandler(os.Stdout, nil)),
	}
	uuids := []string{"u1", "u2", "u3"}

	const q = `SELECT id\s+FROM cameras\s+WHERE id = \$1`

	b := mock.ExpectBatch()
	b.ExpectQuery(q).WithArgs("u1").
		WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow("u1"))
	b.ExpectQuery(q).WithArgs("u2").
		WillReturnRows(pgxmock.NewRows([]string{"id"}))
	b.ExpectQuery(q).WithArgs("u3").
		WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow("u3"))

	got, err := svc.AreCamerasPaired(context.Background(), uuids)
	if err != nil {
		t.Fatalf("AreCamerasPaired error: %v", err)
	}

	expected := []bool{true, false, true}
	if len(got) != len(expected) {
		t.Fatalf("len mismatch: got %d, want %d", len(got), len(expected))
	}
	for i := range expected {
		if got[i] != expected[i] {
			t.Fatalf("idx %d: got %v, want %v", i, got[i], expected[i])
		}
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
