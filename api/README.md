# A python API to return a text embedding 

The API requires an embedding model to operate. I used [jina-embeddings-v](https://huggingface.co/jinaai/jina-embeddings-v3).

For a full writeup, please refer to [my blog post](https://dizzy.zone/2024/12/02/ML-for-related-posts-on-Hugo/).

To build, execute `docker build .`.

The API server only includes a single endpoint:

```
curl --request POST \
  --url http://localhost:5555/embedding \
  --header 'Content-Type: application/json' \
  --data '{
  "text": "Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean massa."
}'

[0.0701325461268425,...]
```

## Configuration

| **Environment Variable** | **Description**                             | **Default Value**                                   |
|--------------------------|---------------------------------------------|-----------------------------------------------------|
| `PORT`                   | Port number for the server to run on.        | `5555`                                              |
| `BEARER_TOKEN`           | Bearer token for authentication. Disabled if empty            | `''` (empty string)                                 |
| `MODEL_DIR`              | Path to the model directory.                 | `../jina-embeddings-v3`                             |
| `POSTGRES_DSN`           | PostgreSQL Data Source Name (DSN) for the database connection. | `postgresql://postgres:postgres@localhost:5432/embedding` |


This code is not production ready and only serves as an example.

