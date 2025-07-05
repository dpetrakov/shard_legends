#!/bin/bash

# Получить JWT токен
JWT_TOKEN=$(curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"telegram_id": 56851083}' | jq -r '.access_token')

echo "JWT Token: $JWT_TOKEN"

# Сделать запрос к daily-chest/claim
curl -X POST http://localhost:8080/deck/daily-chest/claim \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -d '{"combo": 15, "chest_indices": [1]}' | jq .
