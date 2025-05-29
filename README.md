# Simple Wallet App
A REST API made in Go for financial transactions.

## Requirements
- Go 1.23 or higher
- Docker installed

## Running locally
There is a `docker-compose.yml` for postgres and the backend service, but you can choose to run the postgres using docker & run the service manually

### Only use Docker for Postgres
1. Run `docker compose up postgres -d`
2. Download dependencies using `go mod download`
3. Run `go run cmd/main.go`

### Use Docker for Postgres & backend service
Run `docker compose up` (or `docker compose up --build` after any code change, to ensure changed code is rebuilt)

## API endpoints
### Creating a wallet/account
| Method | Path     |
|--------|----------|
| POST   | /wallets |

Request body
```json
{
    "account_name": "John Doe"
}
```
Response
```json
{
    "account_id": 1,
    "account_name": "John Doe"
}
```

### Deposit/Withdraw funds
| Method | Path                              |
|--------|-----------------------------------|
| POST   | /wallets/:account_id/transactions |

Request body
```json
{
    "amount": "0.112233445566778899",
    "description": "My first deposit",
    "transaction_type": "deposit"
}
```
`transaction_type` should be "deposit" or "withdrawal" 

Response
```json
{
    "message": "New transaction successful"
}
```

### Transfer to another account
| Method | Path       |
|--------|------------|
| POST   | /transfers |

Request body
```json
{
    "from_account_id": 1,
    "to_account_id": 2,
    "amount": "0.1",
    "description": "Transfer to Jack"
}
```

Response
```json
{
    "message": "Transfer successful"
}
```

### Get wallet balance
| Method | Path                 |
|--------|----------------------|
| GET    | /wallets/:account_id |

Response
```json
{
    "account_id": 1,
    "balance": "0.112233445566778899"
}
```

### Get wallet history
| Method | Path                              |
|--------|-----------------------------------|
| GET    | /wallets/:account_id/transactions |

Accepts two optional query params:
- start_date: YYYY-MM-DD format, searches for transactions that are later than `start_date` (inclusive)
- end_date: YYYY-MM-DD format, searches for transactions that are earlier than `end_date` (inclusive)

Response
```json
{
    "account_id": 1,
    "start_date": "2025-05-28",
    "end_date": "2025-05-30",
    "transactions": [
        {
            "transaction_id": 3,
            "transaction_date": "2025-05-29T10:40:04.384171Z",
            "description": "Transfer to Jane",
            "ledger_id": 3,
            "account_id": 1,
            "amount": "0.00000000000000009",
            "is_credit": true
        },
        {
            "transaction_id": 2,
            "transaction_date": "2025-05-29T10:29:48.655115Z",
            "description": "My first withdrawal",
            "ledger_id": 2,
            "account_id": 1,
            "amount": "0.1122334455667788",
            "is_credit": true
        },
        {
            "transaction_id": 1,
            "transaction_date": "2025-05-29T10:29:39.184188Z",
            "description": "My first deposit",
            "ledger_id": 1,
            "account_id": 1,
            "amount": "0.112233445566778899",
            "is_credit": false
        }
    ]
}
```