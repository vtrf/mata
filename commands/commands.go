package commands

import (
	"github.com/spf13/cobra"
)

func SetupCommands() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "mata",
		Short:             "mata is a CLI tool for mataroa.blog",
		CompletionOptions: cobra.CompletionOptions{HiddenDefaultCmd: true},
		DisableAutoGenTag: true,
	}

	cmd.AddCommand(newInitCommand())
	cmd.AddCommand(newPostsCommand())
	cmd.AddCommand(newSyncCommand())

	return cmd
}
