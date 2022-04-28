package commands

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"

	"git.sr.ht/~glorifiedgluer/mata/mataroa"
	"github.com/spf13/cobra"
)

func newPostsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "posts",
		Short: "Manage posts",
	}

	cmd.AddCommand(newPostsCreateCommand())
	cmd.AddCommand(newPostsDeleteCommand())
	cmd.AddCommand(newPostsEditCommand())
	cmd.AddCommand(newPostsGetCommand())
	cmd.AddCommand(newPostsListCommand())
	cmd.AddCommand(newPostsUpdateCommand())

	return cmd
}

func newPostsCreateCommand() *cobra.Command {
	run := func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()

		filePath := args[0]

		if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
			log.Fatalf("%s: error creating new post: file '%s' not found\n", cmd.Use, filePath)
		}

		f, err := ioutil.ReadFile(filePath)
		if err != nil {
			log.Fatalf("%s: error reading markdown file: %s", cmd.Use, err)
		}

		post, err := mataroa.NewPost(f)
		if err != nil {
			log.Fatalf("%s: error creating new post: %s\n", cmd.Use, err)
		}

		c := mataroa.NewMataroaClient()

		resp, err := c.CreatePost(ctx, mataroa.PostsCreateResquest{
			Title:       post.Title,
			PublishedAt: post.PublishedAt,
			Body:        post.Body,
		})
		if err != nil {
			log.Fatalf("%s: error creating new post: %s\n", cmd.Use, err)
		}

		if resp.OK {
			fmt.Printf("%s: '%s' created successfully\n%s\n", cmd.Use, resp.Slug, resp.URL)
			fmt.Printf("%s\n", resp.URL)
			fmt.Printf("\n")
		}
	}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a post",
		Args:  cobra.ExactArgs(1),
		Run:   run,
	}
	return cmd
}

func newPostsDeleteCommand() *cobra.Command {
	run := func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()

		slug := args[0]

		c := mataroa.NewMataroaClient()

		response, err := c.DeletePost(ctx, slug)
		if err != nil {
			log.Fatalf("%s: couldn't delete post: %s", cmd.Use, err)
		}

		if !response.OK {
			log.Fatalf("%s: couldn't delete post '%s': %s", cmd.Use, slug, response.Error)
		}

		fmt.Printf("%s: '%s' deleted sucessfully\n", cmd.Use, slug)
	}

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a post",
		Args:  cobra.ExactArgs(1),
		Run:   run,
	}
	return cmd
}

func newPostsEditCommand() *cobra.Command {
	run := func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		slug := args[0]

		c := mataroa.NewMataroaClient()

		response, err := c.PostBySlug(ctx, slug)
		if err != nil {
			log.Fatalf("%s: couldn't get post '%s': %s", cmd.Use, slug, err)
		}

		file, err := os.CreateTemp("", "mata")
		if err != nil {
			log.Fatalf("%s: couldn't create temp file: %s", cmd.Use, err)
		}

		_, err = file.WriteString(response.Post.ToMarkdown())
		if err != nil {
			log.Fatalf("%s: couldn't write markdown to file: %s", cmd.Use, err)
		}

		editor := os.Getenv("EDITOR")
		if len(editor) == 0 {
			log.Fatalf("%s: couldn't edit post $EDITOR environment variable not set", cmd.Use)
		}

		tempname := file.Name()
		defer os.Remove(tempname)

		shellCommand := exec.Command(editor, tempname)
		shellCommand.Stdin = os.Stdin
		shellCommand.Stdout = os.Stdout
		err = shellCommand.Run()
		if err != nil {
			log.Fatalf("%s: error while spawning $EDITOR: %s", cmd.Use, err)
		}

		_, err = file.Seek(0, 0)
		if err != nil {
			log.Fatalf("%s: error offsetting to the beginning of the file: %s", cmd.Use, err)
		}

		f, err := io.ReadAll(file)
		if err != nil {
			log.Fatalf("%s: error reading temporary markdown file: %s", cmd.Use, err)
		}

		post, err := mataroa.NewPost(f)
		if err != nil {
			log.Fatalf("%s: couldn't read new post body from temp file: %s", cmd.Use, err)
		}

		updateResponse, err := c.UpdatePost(ctx, slug, post)
		if err != nil {
			log.Fatal(err)
		}

		if !mataroa.HasPostChanged(response.Post, post) {
			fmt.Printf("%s: '%s' has not changed, skipping update", cmd.Use, slug)
			return
		}

		if updateResponse.OK {
			fmt.Printf("%s: '%s' updated sucessfully", cmd.Use, slug)
		} else {
			log.Fatalf("%s: couldn't update the post '%s': %s ", cmd.Use, slug, updateResponse.Error)
		}
	}

	cmd := &cobra.Command{
		Use:   "edit",
		Short: "Edit a post",
		Args:  cobra.ExactArgs(1),
		Run:   run,
	}
	return cmd
}

func newPostsGetCommand() *cobra.Command {
	run := func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()

		slug := args[0]

		c := mataroa.NewMataroaClient()
		response, err := c.PostBySlug(ctx, slug)
		if err != nil {
			log.Fatalf("%s: couldn't get post '%s': %s", cmd.Use, slug, err)
		}

		if !response.OK {
			log.Fatalf("%s: couldn't get post '%s': %s", cmd.Use, slug, response.Error)
		}

		md := response.Post.ToMarkdown()
		fmt.Println(md)
	}

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a post",
		Args:  cobra.ExactArgs(1),
		Run:   run,
	}
	return cmd
}

func newPostsUpdateCommand() *cobra.Command {
	run := func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		slug := args[0]
		filePath := args[1]

		f, err := ioutil.ReadFile(filePath)
		if err != nil {
			log.Fatalf("%s: error reading markdown file: %s", cmd.Use, err)
		}

		post, err := mataroa.NewPost(f)
		if err != nil {
			log.Fatal(err)
		}
		c := mataroa.NewMataroaClient()

		response, err := c.UpdatePost(ctx, slug, post)
		if err != nil {
			log.Fatal(err)
		}

		if response.OK {
			fmt.Printf("%s: '%s' successfully updated\n", cmd.Use, slug)
		} else {
			log.Fatalf("%s: couldn't update '%s': %s", cmd.Use, slug, response.Error)
		}
	}

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a post",
		Args:  cobra.ExactArgs(2), // TODO: Add slug flag
		Run:   run,
	}
	return cmd
}

func newPostsListCommand() *cobra.Command {
	run := func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		c := mataroa.NewMataroaClient()

		posts, err := c.ListPosts(ctx)
		if err != nil {
			log.Fatal(err)
		}

		for _, post := range posts {
			fmt.Printf("%s\t%s\t%s\t\n", post.PublishedAt, post.Slug, post.Title)
			fmt.Printf("%s\n", post.URL)
			fmt.Printf("\n")
		}
	}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List posts",
		Args:  cobra.ExactArgs(0),
		Run:   run,
	}
	return cmd
}
