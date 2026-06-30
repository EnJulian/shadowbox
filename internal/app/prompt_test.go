package app

import (
	"context"
	"errors"
	"testing"
)

func TestChooseAutoPicksSingle(t *testing.T) {
	idx, err := choose(context.Background(), Options{}, PromptRequest{
		Title:   "Test",
		Options: []PromptOption{{Label: "only"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if idx != 0 {
		t.Fatalf("idx = %d, want 0", idx)
	}
}

func TestChooseErrorsOnEmpty(t *testing.T) {
	_, err := choose(context.Background(), Options{}, PromptRequest{Title: "Test"})
	if err == nil {
		t.Fatal("expected error for empty options")
	}
}

func TestChooseUsesSelectFunc(t *testing.T) {
	opts := Options{
		Select: func(ctx context.Context, req PromptRequest) (int, error) {
			if req.Title != "Pick one" || len(req.Options) != 2 {
				t.Fatalf("unexpected request: %+v", req)
			}
			return 1, nil
		},
	}
	idx, err := choose(context.Background(), opts, PromptRequest{
		Title: "Pick one",
		Options: []PromptOption{
			{Label: "a"},
			{Label: "b"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if idx != 1 {
		t.Fatalf("idx = %d, want 1", idx)
	}
}

func TestChoosePropagatesCancel(t *testing.T) {
	opts := Options{
		Select: func(ctx context.Context, req PromptRequest) (int, error) {
			return -1, ErrSelectionCancelled
		},
	}
	_, err := choose(context.Background(), opts, PromptRequest{
		Title:   "Pick one",
		Options: []PromptOption{{Label: "a"}, {Label: "b"}},
	})
	if !errors.Is(err, ErrSelectionCancelled) {
		t.Fatalf("err = %v, want ErrSelectionCancelled", err)
	}
}
