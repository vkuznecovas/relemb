# Relembed 

This is a repository that contains example code for one of my blog posts.

For a full writeup, please refer to [my blog post](https://dizzy.zone/2024/12/02/ML-for-related-posts-on-Hugo/).

It deals with text embeddings for blog posts and updates the relevant markdown files with the paths to related posts.

It contains two parts:

1) The Python HTTP `./api` to calculate embeddings for a given text. [Readme](./api/README.md)
2) The Go `./cli` to update the related posts in markdown files for my own blog. [Readme](./cli/README.md)

This code is not production ready and only serves as an example.

