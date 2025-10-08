#!/bin/bash

BASE_URL="http://localhost:8081"

echo "=== Inserting vectors ==="
curl -s -X POST $BASE_URL/vectors -H "Content-Type: application/json" -d '{
  "id": "doc1",
  "vector": [0.1, 0.2, 0.3],
  "metadata": {"title": "AI Doc", "category": "tech"}
}' && echo

curl -s -X POST $BASE_URL/vectors -H "Content-Type: application/json" -d '{
  "id": "doc2",
  "vector": [0.2, 0.1, 0.4],
  "metadata": {"title": "ML Paper", "category": "tech"}
}' && echo

curl -s -X POST $BASE_URL/vectors -H "Content-Type: application/json" -d '{
  "id": "doc3",
  "vector": [0.9, 0.8, 0.7],
  "metadata": {"title": "Cooking Tips", "category": "food"}
}' && echo

echo "=== Searching all vectors ==="
curl -s -X POST $BASE_URL/search -H "Content-Type: application/json" -d '{
  "query": [0.1, 0.2, 0.3],
  "top_k": 3
}' && echo

echo "=== Searching with filter ==="
curl -s -X POST $BASE_URL/search -H "Content-Type: application/json" -d '{
  "query": [0.1, 0.2, 0.3],
  "top_k": 3,
  "filter": {"category": "tech"}
}' && echo

echo "=== Testing pagination (page=1, limit=2) ==="
curl -s -X POST $BASE_URL/search -H "Content-Type: application/json" -d '{
  "query": [0.1, 0.2, 0.3],
  "top_k": 3,
  "page": 1,
  "limit": 2
}' && echo

echo "=== Testing pagination (page=2, limit=2) ==="
curl -s -X POST $BASE_URL/search -H "Content-Type: application/json" -d '{
  "query": [0.1, 0.2, 0.3],
  "top_k": 3,
  "page": 2,
  "limit": 2
}' && echo

