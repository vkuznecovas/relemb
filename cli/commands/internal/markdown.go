package internal

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/russross/blackfriday/v2"
	"gopkg.in/yaml.v3"
)

type Frontmatter struct {
	Title        string    `yaml:"title"`
	Date         time.Time `yaml:"date"`
	Draft        bool      `yaml:"draft"`
	Author       string    `yaml:"author"`
	Tags         []string  `yaml:"tags"`
	Description  string    `yaml:"description"`
	SimilarPosts []string  `yaml:"similar_posts"`
}

type Content []byte

var shortcodeRegex = regexp.MustCompile(`(?s)\{\{<\s*[^>]+>\}\}.*?\{\{<\s*/[^>]+>\}\}`)

func (c Content) Strip() (string, error) {
	html := blackfriday.Run(c)
	doc, err := goquery.NewDocumentFromReader(bytes.NewBuffer(html))
	if err != nil {
		return "", err
	}

	text := doc.Text()
	cleanedText := shortcodeRegex.ReplaceAllString(text, "")
	return cleanedText, nil
}

type MarkdownPost struct {
	RawContent  []byte // frontmatter + separators + content
	Path        string
	Frontmatter Frontmatter // frontmatter
	Content     Content     // does not include frontmatter
	Embedding   []float64
}

func (mp MarkdownPost) Hash() string {
	hash := sha256.Sum256(mp.RawContent)
	return fmt.Sprintf("%x", hash)
}

func (m *MarkdownPost) Save() error {
	frontmatterYaml, err := yaml.Marshal(m.Frontmatter)
	if err != nil {
		return fmt.Errorf("failed to marshal frontmatter: %w", err)
	}

	var buffer bytes.Buffer
	buffer.WriteString("---\n")
	buffer.Write(frontmatterYaml)
	buffer.WriteString("---")
	buffer.Write(m.Content)

	if err := os.WriteFile(m.Path, buffer.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
