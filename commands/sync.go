package commands

import (
	"io/fs"
	"io/ioutil"
	"path/filepath"

	"git.sr.ht/~glorifiedgluer/mata/mataroa"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func isMarkdownFile(path string) bool {
	return filepath.Ext(path) == ".md" ||
		filepath.Ext(path) == ".markdown"
}

func newSyncCommand() *cobra.Command {
	run := func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()

		c, err := mataroa.NewMataroaClient()
		if err != nil {
			log.WithFields(log.Fields{
				"cmd": cmd.Use,
			}).Fatal(err)
		}

		if err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			isMd := !d.IsDir() && isMarkdownFile(path)
			if !isMd {
				return nil
			}

			f, err := ioutil.ReadFile(path)
			if err != nil {
				log.WithFields(log.Fields{
					"cmd":  cmd.Use,
					"path": path,
				}).Fatal(err)
			}

			post, err := mataroa.NewPost(f)
			if err != nil {
				log.WithFields(log.Fields{
					"cmd":  cmd.Use,
					"path": path,
				}).Fatal(err)
			}

			_, err = c.PostBySlug(ctx, post.Slug)
			if err != nil {
				log.WithFields(log.Fields{
					"cmd":  cmd.Use,
					"path": path,
					"slug": post.Slug,
				}).Warn(err)
				return nil
			}

			resp, err := c.UpdatePost(ctx, post.Slug, post)
			if err != nil {
				log.WithFields(log.Fields{
					"cmd":  cmd.Use,
					"path": path,
					"slug": post.Slug,
				}).Fatal(err)
			}

			if !resp.OK {
				log.WithFields(log.Fields{
					"cmd":  cmd.Use,
					"path": path,
					"slug": post.Slug,
				}).Warn(resp.Error)
				return nil
			}

			log.WithFields(log.Fields{
				"cmd":  cmd.Use,
				"path": path,
				"slug": post.Slug,
			}).Info("updated successfully")

			return nil
		}); err != nil {
			log.WithFields(log.Fields{
				"cmd": cmd.Use,
			}).Fatal(err)
		}
	}

	cmd := &cobra.Command{
		Use:     "sync",
		Aliases: []string{"s"},
		Short:   "sync all your posts",
		Args:    cobra.ExactArgs(0),
		Run:     run,
	}
	return cmd
}
