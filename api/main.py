import os
from flask import Flask, request, jsonify
from transformers import AutoModel
from sklearn.metrics.pairwise import cosine_similarity
import psycopg2
from psycopg2 import pool, sql
import hashlib
from psycopg2.extras import execute_values


PORT = int(os.getenv("PORT", 5555)) 
BEARER_TOKEN = os.getenv("BEARER_TOKEN", '')
MODEL_DIR = os.getenv("MODEL_DIRECTORY", "../jina-embeddings-v3")
POSTGRES_DSN = os.getenv("POSTGRES_DSN", 'postgresql://postgres:postgres@localhost:5432/embedding')

app = Flask(__name__)

connection_pool = psycopg2.pool.SimpleConnectionPool(
    1,  
    10, 
    dsn=POSTGRES_DSN
)

model = AutoModel.from_pretrained(MODEL_DIR, trust_remote_code=True)

@app.errorhandler(Exception)
def handle_exception(e):
    print(e)
    return jsonify({'error': 'Internal Server Error'}), 500


@app.before_request
def check_authentication():
    if BEARER_TOKEN == '':
        return None
    auth_header = request.headers.get('Authorization')
    if not auth_header or auth_header != f'Bearer {BEARER_TOKEN}':
        return jsonify({'error': 'Unauthorized'}), 401


@app.route('/embedding', methods=['POST'])
def create_embedding():
    data = request.get_json()
    if not data or 'text' not in data:
        return jsonify({'error': 'Bad Request, missing "text" field'}), 400
    
    hash = hash_to_sha512(data['text'])
    model_name = model.config.model_type

    embedding = fetch_embedding(hash, model_name)
    if embedding is not None:
        return embedding[2]

    texts = [data['text']]
    embeddings = model.encode(texts, task="text-matching")
    embedding_values = embeddings.tolist()[0]

    insert_embedding(hash, embedding_values, model_name)

    return jsonify(embedding_values), 200


def migrate():
    create_table_query = """
    CREATE TABLE IF NOT EXISTS embeddings (
        id SERIAL PRIMARY KEY,
        hash VARCHAR(128) UNIQUE NOT NULL,
        embedding FLOAT[] NOT NULL,
        model VARCHAR(512) NOT NULL
    );
    CREATE INDEX IF NOT EXISTS idx_hash ON embeddings(hash, model);
    """
    conn = connection_pool.getconn()
    try:
        with conn:
            with conn.cursor() as cursor:
                cursor.execute(create_table_query)
    finally:
        connection_pool.putconn(conn)

def fetch_embedding(hash_value, model_value):
    query = """
    SELECT id, hash, embedding, model FROM embeddings
    WHERE hash = %s AND model = %s;
    """
    conn = connection_pool.getconn()
    try:
        with conn:
            with conn.cursor() as cursor:
                cursor.execute(query, (hash_value, model_value))
                result = cursor.fetchone()
                return result
    finally:
        connection_pool.putconn(conn)

def insert_embedding(hash_value, embedding, model_value):
    insert_query = """
    INSERT INTO embeddings (hash, embedding, model)
    VALUES (%s, %s, %s)
    ON CONFLICT (hash) DO NOTHING;
    """
    conn = connection_pool.getconn()
    try:
        with conn:
            with conn.cursor() as cursor:
                cursor.execute(insert_query, (hash_value, embedding, model_value))
    finally:
        connection_pool.putconn(conn)

def hash_to_sha512(input_string):
    byte_input = input_string.encode('utf-8')
    sha512_hash = hashlib.sha512(byte_input)
    return sha512_hash.hexdigest()

if __name__ == '__main__':
    migrate()
    from waitress import serve
    serve(app, host="0.0.0.0", port=PORT)