package migration

import (
	"testing"
	"testing/fstest"
)

func TestGetAllFilenames(t *testing.T) {
	t.Run("lists files and skips directories", func(t *testing.T) {
		fsys := fstest.MapFS{
			"00001_users.sql":       {Data: []byte("-- +goose Up")},
			"00002_predictions.sql": {Data: []byte("-- +goose Up")},
			"sub/00003_orders.sql":  {Data: []byte("-- +goose Up")},
		}

		got, err := getAllFilenames(fsys)
		if err != nil {
			t.Fatalf("expected nil error, got: %v", err)
		}

		want := map[string]bool{
			"00001_users.sql":       true,
			"00002_predictions.sql": true,
			"sub/00003_orders.sql":  true,
		}
		if len(got) != len(want) {
			t.Fatalf("expected %d filenames, got: %v", len(want), got)
		}
		for _, name := range got {
			if !want[name] {
				t.Errorf("unexpected filename: %s", name)
			}
		}
	})
}
