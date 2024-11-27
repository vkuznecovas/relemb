# A go CLI to update related posts on a hugo blog

The CLI finds all posts (markdown files) on my blog, fetches their embedding values, calculates consine similarity for each pair and saves the top 3 as related in the frontmatter.

To run `go run cli/cmd/main.go update-related`. 

## Configuration

| **CLI Option**             | **Description**                                       | **Default Value**                |
|----------------------------|-------------------------------------------------------|----------------------------------|
| `--embed-api-url`           | The fully qualified URL for the embed API.             | `http://localhost:5555`          |
| `--embed-api-token`         | The token for the embed API (optional).                | *(None)*                         |
| `--post-dir`                | The directory containing the posts.                    | `./content/posts`                |
| `--help, -h`                | Displays help information for the command.             | *(N/A)*                          |


This code is not production ready and only serves as an example.