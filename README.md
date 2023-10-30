# go-rest-balance

POC for encryption purposes.

CRUD a balance data

## Database

    CREATE TABLE balance (
        id              serial not null,
        account_id      varchar(200) NULL,
        person_id       varchar(200) NULL,
        currency        varchar(10) NULL,   
        amount          float8 NULL,
        create_at       timestamptz NULL,
        update_at       timestamptz NULL,
        tenant_id       varchar(200) NULL
    );

## Endpoints

+ POST /add

        {
            "account_id": "ACC-100",
            "person_id": "P-100",
            "currency": "BRL",
            "amount": 100,
            "create_at": "0001-01-01T00:00:00Z",
            "tenant_id": "TENANT-001"
        }


+ GET /get/ACC-003

+ GET /header

+ GET /list/P-002

+ POST /update/ACC-003

        {
            "person_id": "P-002",
            "currency": "BRL",
            "amount": 200.99,
            "tenant_id": "TENANT-001"
        }

+ DELETE /delete/ACC-001

+ POST /sum

        {
            "account_id": "ACC-003",
            "amount": 100.00,
            "tenant_id": "TENANT-001"
        }