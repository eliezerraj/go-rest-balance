# go-rest-balance

POC for test purposes.

CRUD a balance data

## Database

        CREATE TABLE balance (
            id              SERIAL PRIMARY KEY,
            account_id      varchar(200) UNIQUE NULL,
            person_id       varchar(200) NULL,
            currency        varchar(10) NULL,   
            amount          float8 NULL,
            create_at       timestamptz NULL,
            update_at       timestamptz NULL,
            tenant_id       varchar(200) null,
            user_last_update	varchar(200) NULL);

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

        curl svc01.domain.com/get/ACC-001 | jq

+ GET /header

+ GET /list/P-002

        curl svc01.domain.com/list/P-003 | jq

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

+ POST /minus

        {
            "account_id": "ACC-003",
            "amount": 100.00,
            "tenant_id": "TENANT-001"
        }

## K8 local

Add in hosts file /etc/hosts the lines below

    127.0.0.1   svc01.domain.com

## AWS

Create a public apigw
