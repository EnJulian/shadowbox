package app

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// ErrSelectionCancelled is returned when the user aborts an interactive prompt.
var ErrSelectionCancelled = errors.New("selection cancelled")

// PromptOption is one row in an interactive selection list.
type PromptOption struct {
	Label  string // primary line, e.g. "Believer — Imagine Dragons"
	Detail string // secondary line, e.g. "Evolve (2017) · 3:24"
}

// PromptRequest asks the user to pick one of several options.
type PromptRequest struct {
	Title   string
	Options []PromptOption
}

// SelectFunc returns the chosen index for a prompt, or an error if cancelled.
type SelectFunc func(ctx context.Context, req PromptRequest) (int, error)

// choose picks an option: auto-selects when exactly one, prompts when multiple,
// and returns an error when there are no options.
func choose(ctx context.Context, opts Options, req PromptRequest) (int, error) {
	n := len(req.Options)
	switch n {
	case 0:
		return -1, fmt.Errorf("no options to choose from")
	case 1:
		return 0, nil
	}

	if opts.Select != nil {
		return opts.Select(ctx, req)
	}
	return chooseCLI(ctx, req)
}

func chooseCLI(ctx context.Context, req PromptRequest) (int, error) {
	fmt.Println()
	fmt.Println(req.Title)
	for i, o := range req.Options {
		if err := ctx.Err(); err != nil {
			return -1, err
		}
		line := fmt.Sprintf("  %d) %s", i+1, o.Label)
		if o.Detail != "" {
			line += " — " + o.Detail
		}
		fmt.Println(line)
	}
	fmt.Print("\nEnter number (1-", len(req.Options), "): ")

	reader := bufio.NewReader(os.Stdin)
	for {
		if err := ctx.Err(); err != nil {
			return -1, err
		}
		line, err := reader.ReadString('\n')
		if err != nil {
			return -1, err
		}
		line = strings.TrimSpace(line)
		if line == "" {
			return -1, ErrSelectionCancelled
		}
		idx, err := strconv.Atoi(line)
		if err != nil || idx < 1 || idx > len(req.Options) {
			fmt.Print("Invalid choice, try again: ")
			continue
		}
		return idx - 1, nil
	}
}
