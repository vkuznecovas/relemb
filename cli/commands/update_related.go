package commands

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/urfave/cli/v2"
	"github.com/vkuznecovas/relemb/cli/commands/internal"
	"gopkg.in/yaml.v3"
)

func NewUpdateRelated() *cli.Command {
	return &cli.Command{
		Name:    "update-related",
		Aliases: []string{"ur"},
		Usage:   "updates related posts",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "embed-api-url",
				Usage: "the fully qualified embedded api url",
				Value: "http://localhost:5555",
			},
			&cli.StringFlag{
				Name:  "embed-api-token",
				Usage: "the token for the embed api (optional)",
				Value: "",
			},
			&cli.PathFlag{
				Name:  "post-dir",
				Value: "./content/posts",
				Usage: "the directory of the posts",
			},
		},
		Action: func(cCtx *cli.Context) error {
			dir := cCtx.String("post-dir")
			posts, err := loadMarkdownPosts(dir)
			if err != nil {
				log.Fatalf("Error loading posts: %v", err)
			}

			client := internal.NewEmbedClient(cCtx.String("embed-api-url"), cCtx.String("embed-api-token"))
			err = loadEmbeddings(cCtx.Context, client, posts)
			if err != nil {
				log.Fatalf("Error loading embeddings: %v", err)
			}

			err = findTop3SimilarPosts(posts)
			if err != nil {
				log.Fatalf("Error finding top 3 similar posts: %v", err)
			}

			for _, post := range posts {
				err := post.Save()
				if err != nil {
					log.Fatalf("Could not save file: %v", err)
				}
			}

			return nil
		},
	}
}

func parseMarkdownFile(path string) (internal.MarkdownPost, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return internal.MarkdownPost{}, err
	}

	parts := bytes.SplitN(data, []byte("---"), 3)
	if len(parts) < 3 {
		return internal.MarkdownPost{}, fmt.Errorf("invalid frontmatter in file: %s", path)
	}

	var frontmatter internal.Frontmatter
	if err := yaml.Unmarshal(parts[1], &frontmatter); err != nil {
		return internal.MarkdownPost{}, fmt.Errorf("error parsing YAML in file %s: %w", path, err)
	}

	return internal.MarkdownPost{
		RawContent:  data,
		Path:        path,
		Frontmatter: frontmatter,
		Content:     internal.Content(parts[2]), // content without frontmatter
	}, nil
}

func loadMarkdownPosts(dir string) ([]*internal.MarkdownPost, error) {
	var posts []*internal.MarkdownPost

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.EqualFold(d.Name(), "index.md") {
			return nil
		}

		post, err := parseMarkdownFile(path)
		if err != nil {
			log.Printf("Error parsing %s: %v", path, err)
			return nil
		}

		posts = append(posts, &post)
		return nil
	})

	return posts, err
}

func loadEmbeddings(ctx context.Context, client *internal.EmbedClient, posts []*internal.MarkdownPost) error {
	sema := make(chan struct{}, 5)
	defer close(sema)

	wg := sync.WaitGroup{}

	var firstErr error
	var once sync.Once

	cctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for _, v := range posts {
		sema <- struct{}{}
		wg.Add(1)

		go func(post *internal.MarkdownPost) {
			defer func() {
				wg.Done()
				<-sema
			}()

			select {
			case <-cctx.Done():
				return
			default:
				// Proceed normally
			}

			stripped, err := v.Content.Strip()
			if err != nil {
				once.Do(func() {
					firstErr = fmt.Errorf("failed to strip %q: %v", post.Path, err)
					cancel()
				})
				return
			}

			embedding, err := client.GetEmbedding(cctx, []byte(stripped))
			if err != nil {
				once.Do(func() {
					firstErr = fmt.Errorf("failed to get embedding %q: %v", post.Path, err)
					cancel()
				})
				return
			}

			v.Embedding = embedding
		}(v)
	}

	wg.Wait()

	return firstErr
}

func findTop3SimilarPosts(posts []*internal.MarkdownPost) error {
	for _, post := range posts {
		similarities := make(map[string]float64)

		for _, otherPost := range posts {
			if post.Path == otherPost.Path {
				continue
			}

			if otherPost.Frontmatter.Draft {
				continue
			}

			if otherPost.Frontmatter.Date.After(time.Now()) {
				continue
			}

			similarity, err := internal.CosineSimilarity(post.Embedding, otherPost.Embedding)
			if err != nil {
				return fmt.Errorf("failed to compute similarity for %q and %q: %v", post.Path, otherPost.Path, err)
			}

			similarities[otherPost.Path] = similarity
		}

		type postSimilarity struct {
			Path       string
			Similarity float64
		}
		similarityList := make([]postSimilarity, 0, len(similarities))
		for path, similarity := range similarities {
			similarityList = append(similarityList, postSimilarity{Path: path, Similarity: similarity})
		}

		sort.Slice(similarityList, func(i, j int) bool {
			return similarityList[i].Similarity > similarityList[j].Similarity
		})

		top3 := similarityList
		if len(similarityList) > 3 {
			top3 = similarityList[:3]
		}

		post.Frontmatter.SimilarPosts = make([]string, 0, 3)
		for _, entry := range top3 {
			path := strings.TrimPrefix(entry.Path, "content")
			path = strings.TrimSuffix(path, "/index.md")

			post.Frontmatter.SimilarPosts = append(post.Frontmatter.SimilarPosts, path)
		}
	}

	return nil
}
