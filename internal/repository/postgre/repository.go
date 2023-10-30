package db_postgre

import (
	"context"
	"time"

	_ "github.com/lib/pq"

	"github.com/go-rest-balance/internal/core"
	"github.com/go-rest-balance/internal/erro"

)

type WorkerRepository struct {
	databaseHelper DatabaseHelper
}

func NewWorkerRepository(databaseHelper DatabaseHelper) WorkerRepository {
	childLogger.Debug().Msg("NewWorkerRepository")
	return WorkerRepository{
		databaseHelper: databaseHelper,
	}
}

func (w WorkerRepository) Ping() (bool, error) {
	childLogger.Debug().Msg("++++++++++++++++++++++++++++++++")
	childLogger.Debug().Msg("Ping")
	childLogger.Debug().Msg("++++++++++++++++++++++++++++++++")

	ctx, cancel := context.WithTimeout(context.Background(), 1000)
	defer cancel()

	client, _ := w.databaseHelper.GetConnection(ctx)
	err := client.Ping()
	if err != nil {
		return false, erro.ErrConnectionDatabase
	}

	return true, nil
}

func (w WorkerRepository) Add(balance core.Balance) (*core.Balance, error){
	childLogger.Debug().Msg("Add")

	ctx, cancel := context.WithTimeout(context.Background(), 30 * time.Second)
	defer cancel()

	client, _ := w.databaseHelper.GetConnection(ctx)

	stmt, err := client.Prepare(`INSERT INTO balance ( 	account_id, 
														person_id, 
														currency,
														amount,
														create_at,
														tenant_id) 
									VALUES($1, $2, $3, $4, $5, $6) `)
	if err != nil {
		childLogger.Error().Err(err).Msg("INSERT statement")
		return nil, erro.ErrInsert
	}
	_, err = stmt.Exec(	balance.AccountID, 
						balance.PersonID,
						balance.Currency,
						balance.Amount,
						time.Now(),
						balance.TenantID)
	if err != nil {
		childLogger.Error().Err(err).Msg("Exec statement")
		return nil, erro.ErrInsert
	}

	return &balance , nil
}

func (w WorkerRepository) Get(balance core.Balance) (*core.Balance, error){
	childLogger.Debug().Msg("Get")

	ctx, cancel := context.WithTimeout(context.Background(), 30 * time.Second)
	defer cancel()

	client, _ := w.databaseHelper.GetConnection(ctx)

	result_query := core.Balance{}
	rows, err := client.Query(`SELECT id, account_id, person_id, currency, amount, create_at, update_at, tenant_id
								FROM balance 
								WHERE account_id =$1`, balance.AccountID)
	if err != nil {
		childLogger.Error().Err(err).Msg("Query statement")
		return nil, erro.ErrNotFound
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan( 	&result_query.ID, 
							&result_query.AccountID, 
							&result_query.PersonID, 
							&result_query.Currency,
							&result_query.Amount,
							&result_query.CreateAt,
							&result_query.UpdateAt,
							&result_query.TenantID,
						)
		if err != nil {
			childLogger.Error().Err(err).Msg("Scan statement")
			return nil, erro.ErrNotFound
        }
		return &result_query, nil
	}

	return nil, erro.ErrNotFound
}

func (w WorkerRepository) Update(balance core.Balance) (bool, error){
	childLogger.Debug().Msg("Update...")

	//childLogger.Debug().Interface("balance : ", balance).Msg("balance")

	ctx, cancel := context.WithTimeout(context.Background(), 30 * time.Second)
	defer cancel()

	client, _ := w.databaseHelper.GetConnection(ctx)

	stmt, err := client.Prepare(`Update balance
									set account_id = $1, 
										person_id = $2, 
										currency = $3, 
										amount = $4, 
										update_at = $5
								where id = $6 `)
	if err != nil {
		childLogger.Error().Err(err).Msg("UPDATE statement")
		return false, erro.ErrUpdate
	}
	_, err = stmt.Exec(	balance.AccountID, 
						balance.PersonID,
						balance.Currency,
						balance.Amount,
						time.Now(),
						balance.ID,
					)
	if err != nil {
		childLogger.Error().Err(err).Msg("Exec statement")
		return false, erro.ErrUpdate
	}

	return true , nil
}

func (w WorkerRepository) Delete(balance core.Balance) (bool, error){
	childLogger.Debug().Msg("Delete")

	ctx, cancel := context.WithTimeout(context.Background(), 30 * time.Second)
	defer cancel()

	client, _ := w.databaseHelper.GetConnection(ctx)

	stmt, err := client.Prepare(`Delete from balance where id = $1 `)
	if err != nil {
		childLogger.Error().Err(err).Msg("DELETE statement")
		return false, erro.ErrDelete
	}
	_, err = stmt.Exec(	balance.ID )
	if err != nil {
		childLogger.Error().Err(err).Msg("Exec statement")
		return false, erro.ErrDelete
	}

	return true , nil
}

func (w WorkerRepository) List(balance core.Balance) (*[]core.Balance, error){
	childLogger.Debug().Msg("List")

	ctx, cancel := context.WithTimeout(context.Background(), 30 * time.Second)
	defer cancel()

	client, _ := w.databaseHelper.GetConnection(ctx)

	result_query := core.Balance{}
	balance_list := []core.Balance{}

	rows, err := client.Query(`SELECT id, account_id, person_id, currency, amount, create_at, update_at, tenant_id
								FROM balance 
								WHERE person_id =$1`, balance.PersonID)
	if err != nil {
		childLogger.Error().Err(err).Msg("SELECT statement")
		return nil, erro.ErrNotFound
	}

	for rows.Next() {
		err := rows.Scan( 	&result_query.ID, 
							&result_query.AccountID, 
							&result_query.PersonID, 
							&result_query.Currency,
							&result_query.Amount,
							&result_query.CreateAt,
							&result_query.UpdateAt,
							&result_query.TenantID,
						)
		if err != nil {
			childLogger.Error().Err(err).Msg("Scan statement")
			return nil, erro.ErrNotFound
        }
		balance_list = append(balance_list, result_query)
	}

	return &balance_list , nil
}

func (w WorkerRepository) Sum(balance core.Balance) (bool, error){
	childLogger.Debug().Msg("Sum")

	ctx, cancel := context.WithTimeout(context.Background(), 30 * time.Second)
	defer cancel()

	client, _ := w.databaseHelper.GetConnection(ctx)

	stmt, err := client.Prepare(`Update balance
									set amount = amount + $1, 
										update_at = $2
								where id = $3 `)
	if err != nil {
		childLogger.Error().Err(err).Msg("UPDATE statement")
		return false, erro.ErrUpdate
	}
	_, err = stmt.Exec(	balance.Amount,
						time.Now(),
						balance.ID,
					)
	if err != nil {
		childLogger.Error().Err(err).Msg("Exec statement")
		return false, erro.ErrUpdate
	}

	return true , nil
}

func (w WorkerRepository) Minus(balance core.Balance) (bool, error){
	childLogger.Debug().Msg("Minus")

	ctx, cancel := context.WithTimeout(context.Background(), 30 * time.Second)
	defer cancel()

	client, _ := w.databaseHelper.GetConnection(ctx)

	stmt, err := client.Prepare(`Update balance
									set amount = amount - $1, 
										update_at = $2
								where id = $3 `)
	if err != nil {
		childLogger.Error().Err(err).Msg("UPDATE statement")
		return false, erro.ErrUpdate
	}
	_, err = stmt.Exec(	balance.Amount,
						time.Now(),
						balance.ID,
					)
	if err != nil {
		childLogger.Error().Err(err).Msg("Exec statement")
		return false, erro.ErrUpdate
	}

	return true , nil
}