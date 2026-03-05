# test-assignment

## Запуск проекта
1. git clone https://github.com/wrong89/test-assignment.git
2. cd test-assignment
3. docker-compose up --build

Bash command:
```bash
git clone https://github.com/wrong89/test-assignment.git && cd test-assignment && docker-compose up --build
```

## API
### 1. Создание пользователя

**POST /v1/user**

- **Описание:** создаёт нового пользователя с указанным балансом.  
- **Body (JSON):**

```json
{
  "user_id": 1,
  "balance": 2500
}
```

### 2. Создание вывода средств (withdrawal)

**POST /v1/withdrawals**

- **Описание:** создаёт withdrawal для пользователя.  
- **Body (JSON):**

```json
{
  "user_id": 1,
  "amount": 100,
  "currency": "USDT",
  "destination": "wallet",
  "idempotency_key": "unique-key-123"
}
```

| Условие                                           | HTTP Status |
|--------------------------------------------------|------------|
| Успех (новый withdrawal)                         | 200        |
| amount <= 0                                      | 400        |
| Недостаточный баланс                              | 409        |
| Повторный idempotency_key + тот же payload       | 200        |
| Повторный idempotency_key + другой payload      | 422        |

## Пример успешного ответа

```json
{
  "id": 1,
  "user_id": 1,
  "amount": "100",
  "currency": "USDT",
  "destination": "wallet",
  "status": "pending"
}
```

### 3. Получение информации о withdrawal

**GET /v1/withdrawals/{id}**

**Пример ответа:**
```json
{
  "id": 1,
  "user_id": 1,
  "amount": "100",
  "currency": "USDT",
  "destination": "wallet",
  "status": "pending"
}
```
Ошибка:
- Если withdrawal с таким id не найден HTTP 404

