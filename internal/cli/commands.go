package cli

import (
	"context"
	"github.com/bruma1994/dyngo/internal/repository"
	"github.com/spf13/cobra"
	"strconv"
)

func NewCLI(notesRepo repository.NotesRepository) *CLI {
	return &CLI{notesRepo: notesRepo}
}

type CLI struct {
	notesRepo repository.NotesRepository
}

func (c *CLI) InitCommands(ctx context.Context) map[string]*cobra.Command {
	var commands = make(map[string]*cobra.Command)

	commands["add"] = &cobra.Command{
		Use:   "add",
		Short: "Add new note to database",
		Run: func(cmd *cobra.Command, args []string) {
			c.notesRepo.AddNote(ctx, args[0], args[1])
		},
	}

	commands["del"] = &cobra.Command{
		Use:   "del",
		Short: "Delete note from database by ID",
		Run: func(cmd *cobra.Command, args []string) {
			id, _ := strconv.Atoi(args[0])
			c.notesRepo.DeleteNote(ctx, int32(id))
		},
	}

	return commands
}
