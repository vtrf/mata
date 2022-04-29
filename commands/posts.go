package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"

	log "github.com/sirupsen/logrus"

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
			log.WithFields(log.Fields{
				"cmd":  cmd.Use,
				"file": filePath,
			}).Fatal(err)
		}

		f, err := ioutil.ReadFile(filePath)
		if err != nil {
			log.WithFields(log.Fields{
				"cmd":  cmd.Use,
				"file": filePath,
			}).Fatalf("error reading markdown file: %s", err)
		}

		post, err := mataroa.NewPost(f)
		if err != nil {
			log.WithFields(log.Fields{
				"cmd":  cmd.Use,
				"path": filePath,
			}).Fatal(err)
		}

		c, err := mataroa.NewMataroaClient()
		if err != nil {
			log.WithFields(log.Fields{
				"cmd":  cmd.Use,
				"path": filePath,
			}).Fatal(err)
		}

		resp, err := c.CreatePost(ctx, mataroa.PostsCreateResquest{
			Title:       post.Title,
			PublishedAt: post.PublishedAt,
			Body:        post.Body,
		})
		if err != nil {
			log.WithFields(log.Fields{
				"cmd":  cmd.Use,
				"path": filePath,
			}).Fatal(err)
		}

		if !resp.OK {
			log.WithFields(log.Fields{
				"cmd":  cmd.Use,
				"path": filePath,
			}).Fatal(resp.Error)
		}

		log.WithFields(log.Fields{
			"cmd":  cmd.Use,
			"slug": resp.Slug,
			"url":  resp.URL,
		}).Info("created successfully")
	}

	cmd := &cobra.Command{
		Use:     "create",
		Aliases: []string{"c"},
		Short:   "Create a post",
		Args:    cobra.ExactArgs(1),
		Run:     run,
	}
	return cmd
}

func newPostsDeleteCommand() *cobra.Command {
	run := func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		slug := args[0]

		c, err := mataroa.NewMataroaClient()
		if err != nil {
			log.WithFields(log.Fields{
				"cmd":  cmd.Use,
				"slug": slug,
			}).Fatal(err)
		}

		response, err := c.DeletePost(ctx, slug)
		if err != nil {
			log.WithFields(log.Fields{
				"cmd":  cmd.Use,
				"slug": slug,
			}).Fatal(err)
		}

		if !response.OK {
			log.WithFields(log.Fields{
				"cmd":  cmd.Use,
				"slug": slug,
			}).Fatal(response.Error)
		}

		log.WithFields(log.Fields{
			"cmd":  cmd.Use,
			"slug": slug,
		}).Info("deleted sucessfully")
	}

	cmd := &cobra.Command{
		Use:     "delete",
		Aliases: []string{"d"},
		Short:   "Delete a post",
		Args:    cobra.ExactArgs(1),
		Run:     run,
	}
	return cmd
}

func newPostsEditCommand() *cobra.Command {
	run := func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		slug := args[0]

		c, err := mataroa.NewMataroaClient()
		if err != nil {
			log.WithFields(log.Fields{
				"cmd":  cmd.Use,
				"slug": slug,
			}).Fatal(err)
		}

		response, err := c.PostBySlug(ctx, slug)
		if err != nil {
			log.WithFields(log.Fields{
				"cmd":  cmd.Use,
				"slug": slug,
			}).Fatal(err)
		}

		tempFile, err := os.CreateTemp("", "mata")
		if err != nil {
			log.WithFields(log.Fields{
				"cmd":  cmd.Use,
				"slug": slug,
			}).Fatal(err)
		}
		defer tempFile.Close()

		_, err = tempFile.WriteString(response.Post.ToMarkdown())
		if err != nil {
			log.WithFields(log.Fields{
				"cmd":  cmd.Use,
				"slug": slug,
			}).Fatalf("couldn't write markdown to temporary file: %s", err)
		}

		editor := os.Getenv("EDITOR")
		if len(editor) == 0 {
			log.WithFields(log.Fields{
				"cmd":  cmd.Use,
				"slug": slug,
			}).Fatal("couldn't edit post $EDITOR environment variable not set")
		}

		tempname := tempFile.Name()
		defer os.Remove(tempname)

		shellCommand := exec.Command(editor, tempname)
		shellCommand.Stdin = os.Stdin
		shellCommand.Stdout = os.Stdout
		err = shellCommand.Run()
		if err != nil {
			log.WithFields(log.Fields{
				"cmd":  cmd.Use,
				"slug": slug,
			}).Fatalf("error while spawning $EDITOR: %s", err)
		}

		_, err = tempFile.Seek(0, 0)
		if err != nil {
			log.WithFields(log.Fields{
				"cmd":  cmd.Use,
				"slug": slug,
			}).Fatalf("error offsetting to the beginning of the file: %s", err)
		}

		f, err := io.ReadAll(tempFile)
		if err != nil {
			log.WithFields(log.Fields{
				"cmd":  cmd.Use,
				"slug": slug,
			}).Fatalf("error reading temporary markdown file: %s", err)
		}

		post, err := mataroa.NewPost(f)
		if err != nil {
			log.WithFields(log.Fields{
				"cmd":  cmd.Use,
				"slug": slug,
			}).Fatalf("couldn't read new post body from temp file: %s", err)
		}

		updateResponse, err := c.UpdatePost(ctx, slug, post)
		if err != nil {
			log.WithFields(log.Fields{
				"cmd":  cmd.Use,
				"slug": slug,
			}).Fatal(err)
		}

		if !mataroa.HasPostChanged(response.Post, post) {
			log.WithFields(log.Fields{
				"cmd":  cmd.Use,
				"slug": slug,
			}).Info("has not changed, skipping update")
			return
		}

		if !updateResponse.OK {
			log.WithFields(log.Fields{
				"cmd":  cmd.Use,
				"slug": slug,
			}).Fatal(response.Error)
		}

		log.WithFields(log.Fields{
			"cmd":  cmd.Use,
			"slug": slug,
		}).Info("edited successfully")
	}

	cmd := &cobra.Command{
		Use:     "edit",
		Aliases: []string{"e"},
		Short:   "Edit a post",
		Args:    cobra.ExactArgs(1),
		Run:     run,
	}
	return cmd
}

func newPostsGetCommand() *cobra.Command {
	var jsonFlag bool
	run := func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()

		slug := args[0]

		c, err := mataroa.NewMataroaClient()
		if err != nil {
			log.WithFields(log.Fields{
				"cmd":  cmd.Use,
				"slug": slug,
			}).Fatal(err)
		}

		response, err := c.PostBySlug(ctx, slug)
		if err != nil {
			log.WithFields(log.Fields{
				"cmd":  cmd.Use,
				"slug": slug,
			}).Fatal(err)
		}

		if !response.OK {
			log.WithFields(log.Fields{
				"cmd":  cmd.Use,
				"slug": slug,
			}).Fatal(response.Error)
		}

		if jsonFlag {
			output, err := json.MarshalIndent(response.Post, "", "  ")
			if err != nil {
				log.WithFields(log.Fields{
					"cmd": cmd.Use,
				}).Fatal(err)
			}

			fmt.Println(string(output))
			return
		}

		md := response.Post.ToMarkdown()
		fmt.Println(md)
	}

	cmd := &cobra.Command{
		Use:     "get",
		Aliases: []string{"g"},
		Short:   "Get a post",
		Args:    cobra.ExactArgs(1),
		Run:     run,
	}
	cmd.Flags().BoolVar(&jsonFlag, "json", false, "output JSON")
	return cmd
}

func newPostsUpdateCommand() *cobra.Command {
	var slugFlag string
	run := func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		filePath := args[0]

		f, err := ioutil.ReadFile(filePath)
		if err != nil {
			log.WithFields(log.Fields{
				"cmd":  cmd.Use,
				"file": filePath,
				"slug": slugFlag,
			}).Fatal(err)
			return
		}

		post, err := mataroa.NewPost(f)
		if err != nil {
			log.WithFields(log.Fields{
				"cmd":  cmd.Use,
				"file": filePath,
				"slug": slugFlag,
			}).Fatal(err)
			return
		}

		c, err := mataroa.NewMataroaClient()
		if err != nil {
			log.WithFields(log.Fields{
				"cmd":  cmd.Use,
				"file": filePath,
				"slug": slugFlag,
			}).Fatal(err)
		}

		if slugFlag == "" {
			slugFlag = post.Slug
		}

		response, err := c.UpdatePost(ctx, slugFlag, post)
		if err != nil {
			log.WithFields(log.Fields{
				"cmd":  cmd.Use,
				"file": filePath,
				"slug": slugFlag,
			}).Fatal(err)
			return
		}

		if !response.OK {
			log.WithFields(log.Fields{
				"cmd":  cmd.Use,
				"file": filePath,
				"slug": slugFlag,
			}).Fatal(response.Error)
			return
		}

		log.WithFields(log.Fields{
			"cmd":  cmd.Use,
			"file": filePath,
			"slug": slugFlag,
		}).Printf("successfully updated")
	}

	cmd := &cobra.Command{
		Use:     "update",
		Aliases: []string{"u"},
		Short:   "Update a post",
		Args:    cobra.ExactArgs(1), // TODO: Add slug flag
		Run:     run,
	}
	cmd.Flags().StringVarP(&slugFlag, "slug", "s", "", "Post slug")

	return cmd
}

func newPostsListCommand() *cobra.Command {
	var jsonFlag bool

	run := func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()

		c, err := mataroa.NewMataroaClient()
		if err != nil {
			log.WithFields(log.Fields{
				"cmd": cmd.Use,
			}).Fatal(err)
		}

		posts, err := c.ListPosts(ctx)
		if err != nil {
			log.WithFields(log.Fields{
				"cmd": cmd.Use,
			}).Fatal(err)
			return
		}

		if jsonFlag {
			output, err := json.MarshalIndent(posts, "", "  ")
			if err != nil {
				log.WithFields(log.Fields{
					"cmd": cmd.Use,
				}).Fatal(err)
				return
			}

			fmt.Println(string(output))
			return
		}

		for _, post := range posts {
			fmt.Printf("%s\t%s\n", post.Slug, post.URL)
		}
	}

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"l"},
		Short:   "List posts",
		Args:    cobra.ExactArgs(0),
		Run:     run,
	}

	cmd.Flags().BoolVar(&jsonFlag, "json", false, "output JSON")
	return cmd
}
