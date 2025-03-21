# Электронный кошелек (Wallet Service)

Сервис для управления балансом пользователей с поддержкой транзакций.

## Особенности

- **Транзакционная модель**: Гарантирует целостность данных при депозитах/снятиях.
- **Retry-логика**: Автоматические повторы операций при ошибках сериализации PostgreSQL.
- **Масштабируемость**: Горизонтальное масштабирование через Docker-контейнеры.
- **Производительность**: Пулы соединений, оптимизированные SQL-запросы, индексы.
- **Логирование**: Детальный мониторинг операций через logrus.

## Технологии

- **Язык**: Go 1.24+

- **Базы данных**: PostgreSQL (основная)
- **Роутинг**: Gorilla Mux
- **Тестирование**: Testify, sqlmock
- **Инфраструктура**: Docker, Docker Compose
- **Библиотеки**:
  - pgx (драйвер PostgreSQL)
  - decimal (точная арифметика)
  - logrus (логирование)

## Установка

### Требования

- Docker 20.10+

- Go 1.24+ (для разработки)

### Запуск

1. Склонируйте репозиторий:

   ```bash
   git clone https://github.com/sunriseex/test_wallet.git
   cd test_wallet
    ```

2. Настройте окружение:

    ```bash
    cp .env.example .env
    ```

3. Запустите docker compose:

   ```bash
    docker-compose up --build
    ```

## Endpoints

POST    `/api/v1/wallet` - Депозит/снятие средств

GET    `/api/v1/wallets/{walletId}` - Получить баланс

### Пример запроса

```bash
# Депозит
curl -X POST http://localhost:8080/api/v1/wallet \
  -H "Content-Type: application/json" \
  -d '{
    "walletId": "550e8400-e29b-41d4-a716-446655440000",
    "operationType": "DEPOSIT",
    "amount": "150.50"
  }'

# Получить баланс
curl http://localhost:8080/api/v1/wallets/550e8400-e29b-41d4-a716-446655440000

   ```

## Тестирование

Запуск unit-тестов:

```bash
go test -v ./...

```

Запуск теста нагрузки:

<span style="color:red"> Необходимо установить [k6](https://grafana.com/docs/k6/latest/set-up/install-k6/) </span>

```bash
k6 run tests/k6test.js
```

## Покрытие тестами

- Валидация входных данных
- Обработка ошибок БД
- Транзакционные сценарии
- Retry-логика
