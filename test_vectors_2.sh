#!/bin/bash

BASE_URL="http://localhost:8080"

echo "========== INSERT VECTORS =========="
curl -s -X POST $BASE_URL/vectors \
-H "Content-Type: application/json" \
-d '{
    "id": "vec001",
    "vector": [0.5, 0.2, 0.1, 0.7],
    "metadata": {"category":"science", "author":"Alice"}
}' &&  echo

echo
curl -s -X POST $BASE_URL/vectors \
-H "Content-Type: application/json" \
-d '{
    "id": "vec002",
    "vector": [0.9, 0.1, 0.4, 0.3],
    "metadata": {"category":"math", "author":"Bob"}
}'  &&  echo

echo
curl -s -X POST $BASE_URL/vectors \
-H "Content-Type: application/json" \
-d '{
    "id": "vec003",
    "vector": [0.2, 0.8, 0.5, 0.1],
    "metadata": {"category":"science", "author":"Charlie"}
}' && echo

echo
echo "========== GET VECTOR =========="
curl -s $BASE_URL/vectors/vec002     echo

echo
echo "========== SEARCH VECTORS =========="
curl -s -X POST $BASE_URL/search \
-H "Content-Type: application/json" \
-d '{
    "query": [0.4, 0.2, 0.1, 0.6],
    "top_k": 2,
    "filter": {"category": "science"},
    "page": 1,
    "limit": 2
}' && echo

echo
echo "========== UPDATE VECTOR =========="
curl -s -X PUT $BASE_URL/vectors/vec003 \
-H "Content-Type: application/json" \
-d '{
    "id": "vec003",
    "vector": [0.3, 0.7, 0.6, 0.2],
    "metadata": {"category": "science", "author": "CharlieUpdated"}
}' && echo

echo
echo "========== DELETE VECTOR =========="
curl -s -X DELETE $BASE_URL/vectors/vec001

echo
echo "========== VERIFY DELETION =========="
curl -s $BASE_URL/vectors/vec001
echo

