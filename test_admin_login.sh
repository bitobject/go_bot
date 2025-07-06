#!/bin/bash

# Тест админ API
echo "=== Тестирование Admin API ==="

# Базовый URL
BASE_URL="http://localhost:8080"

echo "1. Проверка health endpoint..."
curl -s "$BASE_URL/health" | jq .

echo -e "\n2. Проверка readiness endpoint..."
curl -s "$BASE_URL/ready" | jq .

echo -e "\n3. Попытка входа с неправильными данными..."
curl -s -X POST "$BASE_URL/api/admin/login" \
  -H "Content-Type: application/json" \
  -d '{"login": "wrong", "password": "wrong"}' | jq .

echo -e "\n4. Попытка входа с правильными данными (если есть админ)..."
RESPONSE=$(curl -s -X POST "$BASE_URL/api/admin/login" \
  -H "Content-Type: application/json" \
  -d '{"login": "admin", "password": "password"}')

echo "$RESPONSE" | jq .

# Если получили токен, тестируем защищенные endpoint'ы
TOKEN=$(echo "$RESPONSE" | jq -r '.token // empty')
if [ ! -z "$TOKEN" ] && [ "$TOKEN" != "null" ]; then
    echo -e "\n5. Получение профиля администратора..."
    curl -s -X GET "$BASE_URL/api/admin/profile" \
      -H "Authorization: Bearer $TOKEN" | jq .
    
    echo -e "\n6. Попытка доступа без токена..."
    curl -s -X GET "$BASE_URL/api/admin/profile" | jq .
else
    echo -e "\n5. Токен не получен, пропускаем тесты защищенных endpoint'ов"
fi

echo -e "\n=== Тест завершен ===" 