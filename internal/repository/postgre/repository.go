package db_postgre

import (
	"context"
	"time"
	"errors"

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

func (w WorkerRepository) SetSessionVariable(userCredential string) (bool, error) {
	childLogger.Debug().Msg("++++++++++++++++++++++++++++++++")
	childLogger.Debug().Msg("SetSessionVariable")
	childLogger.Debug().Msg("++++++++++++++++++++++++++++++++")

	ctx, cancel := context.WithTimeout(context.Background(), 30 * time.Second)
	defer cancel()

	client := w.databaseHelper.GetConnection(ctx)
	
	stmt, err := client.Prepare("SET sess.user_credential to '" + userCredential+ "'")
	if err != nil {
		childLogger.Error().Err(err).Msg("SET SESSION statement ERROR")
		return false, errors.New(err.Error())
	}

	_, err = stmt.Exec()
	if err != nil {
		childLogger.Error().Err(err).Msg("Exec statement")
		return false, errors.New(err.Error())
	}

	return true, nil
}

func (w WorkerRepository) GetSessionVariable() (string, error) {
	childLogger.Debug().Msg("++++++++++++++++++++++++++++++++")
	childLogger.Debug().Msg("GetSessionVariable")
	childLogger.Debug().Msg("++++++++++++++++++++++++++++++++")

	ctx, cancel := context.WithTimeout(context.Background(), 30 * time.Second)
	defer cancel()

	client := w.databaseHelper.GetConnection(ctx)

	var res_balance string
	rows, err := client.Query("SELECT current_setting('sess.user_credential')" )
	if err != nil {
		childLogger.Error().Err(err).Msg("Prepare statement")
		return "", errors.New(err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan( &res_balance )
		if err != nil {
			childLogger.Error().Err(err).Msg("Scan statement")
			return "", errors.New(err.Error())
        }
		return res_balance, nil
	}

	return "", erro.ErrNotFound
}


func (w WorkerRepository) Ping() (bool, error) {
	childLogger.Debug().Msg("++++++++++++++++++++++++++++++++")
	childLogger.Debug().Msg("Ping")
	childLogger.Debug().Msg("++++++++++++++++++++++++++++++++")

	ctx, cancel := context.WithTimeout(context.Background(), 1000)
	defer cancel()

	client := w.databaseHelper.GetConnection(ctx)

	err := client.Ping()
	if err != nil {
		return false, errors.New(err.Error())
	}

	return true, nil
}

func (w WorkerRepository) Add(balance core.Balance) (*core.Balance, error){
	childLogger.Debug().Msg("Add")

	ctx, cancel := context.WithTimeout(context.Background(), 30 * time.Second)
	defer cancel()

	client := w.databaseHelper.GetConnection(ctx)

	userLastUpdate, _ := w.GetSessionVariable()

	stmt, err := client.Prepare(`INSERT INTO balance ( 	account_id, 
														person_id, 
														currency,
														amount,
														create_at,
														tenant_id,
														user_last_update) 
									VALUES($1, $2, $3, $4, $5, $6, $7) `)
	if err != nil {
		childLogger.Error().Err(err).Msg("INSERT statement")
		return nil, errors.New(err.Error())
	}
	defer stmt.Close()
	
	_, err = stmt.Exec(	balance.AccountID, 
						balance.PersonID,
						balance.Currency,
						balance.Amount,
						time.Now(),
						balance.TenantID,
						userLastUpdate)
	if err != nil {
		childLogger.Error().Err(err).Msg("Exec statement")
		return nil, errors.New(err.Error())
	}

	return &balance , nil
}

func (w WorkerRepository) Get(balance core.Balance) (*core.Balance, error){
	childLogger.Debug().Msg("Get")

	ctx, cancel := context.WithTimeout(context.Background(), 30 * time.Second)
	defer cancel()

	client := w.databaseHelper.GetConnection(ctx)

	result_query := core.Balance{}
	rows, err := client.Query(`SELECT id, account_id, person_id, currency, amount, create_at, update_at, tenant_id, user_last_update
								FROM balance 
								WHERE account_id =$1`, balance.AccountID)
	if err != nil {
		childLogger.Error().Err(err).Msg("Query statement")
		return nil, errors.New(err.Error())
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
							&result_query.UserLastUpdate,
						)
		if err != nil {
			childLogger.Error().Err(err).Msg("Scan statement")
			return nil, errors.New(err.Error())
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

	client := w.databaseHelper.GetConnection(ctx)

	userLastUpdate, _ := w.GetSessionVariable()

	stmt, err := client.Prepare(`Update balance
									set account_id = $1, 
										person_id = $2, 
										currency = $3, 
										amount = $4, 
										update_at = $5,
										user_last_update =$7
								where id = $6 `)
	if err != nil {
		childLogger.Error().Err(err).Msg("UPDATE statement")
		return false, errors.New(err.Error())
	}
	defer stmt.Close()

	result, err := stmt.Exec(	balance.AccountID, 
						balance.PersonID,
						balance.Currency,
						balance.Amount,
						time.Now(),
						balance.ID,
						userLastUpdate,
					)
	if err != nil {
		childLogger.Error().Err(err).Msg("Exec statement")
		return false, errors.New(err.Error())
	}

	rowsAffected, _ := result.RowsAffected()
	childLogger.Debug().Int("rowsAffected : ",int(rowsAffected)).Msg("")

	return true , nil
}

func (w WorkerRepository) Delete(balance core.Balance) (bool, error){
	childLogger.Debug().Msg("Delete")

	ctx, cancel := context.WithTimeout(context.Background(), 30 * time.Second)
	defer cancel()

	client := w.databaseHelper.GetConnection(ctx)

	stmt, err := client.Prepare(`Delete from balance where id = $1 `)
	if err != nil {
		childLogger.Error().Err(err).Msg("DELETE statement")
		return false, errors.New(err.Error())
	}
	defer stmt.Close()

	result, err := stmt.Exec(	balance.ID )
	if err != nil {
		childLogger.Error().Err(err).Msg("Exec statement")
		return false, errors.New(err.Error())
	}

	rowsAffected, _ := result.RowsAffected()
	childLogger.Debug().Int("rowsAffected : ",int(rowsAffected)).Msg("")

	return true , nil
}

func (w WorkerRepository) List(balance core.Balance) (*[]core.Balance, error){
	childLogger.Debug().Msg("List")

	ctx, cancel := context.WithTimeout(context.Background(), 30 * time.Second)
	defer cancel()

	client:= w.databaseHelper.GetConnection(ctx)

	result_query := core.Balance{}
	balance_list := []core.Balance{}

	rows, err := client.Query(`SELECT id, account_id, person_id, currency, amount, create_at, update_at, tenant_id, user_last_update
								FROM balance 
								WHERE person_id =$1`, balance.PersonID)
	if err != nil {
		childLogger.Error().Err(err).Msg("SELECT statement")
		return nil, errors.New(err.Error())
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
							&result_query.UserLastUpdate,
						)
		if err != nil {
			childLogger.Error().Err(err).Msg("Scan statement")
			return nil, errors.New(err.Error())
        }
		balance_list = append(balance_list, result_query)
	}

	return &balance_list , nil
}

func (w WorkerRepository) Sum(balance core.Balance) (bool, error){
	childLogger.Debug().Msg("Sum")

	ctx, cancel := context.WithTimeout(context.Background(), 30 * time.Second)
	defer cancel()

	client := w.databaseHelper.GetConnection(ctx)
	stmt, err := client.Prepare(`Update balance
									set amount = amount + $1, 
										update_at = $2
								where id = $3 `)
	if err != nil {
		childLogger.Error().Err(err).Msg("UPDATE statement")
		return false, errors.New(err.Error())
	}
	defer stmt.Close()

	result, err := stmt.Exec(	balance.Amount,
								time.Now(),
								balance.ID,
							)
	if err != nil {
		childLogger.Error().Err(err).Msg("Exec statement")
		return false, errors.New(err.Error())
	}

	rowsAffected, _ := result.RowsAffected()
	childLogger.Debug().Int("rowsAffected : ",int(rowsAffected)).Msg("")

	return true , nil
}

func (w WorkerRepository) Minus(balance core.Balance) (bool, error){
	childLogger.Debug().Msg("Minus")

	ctx, cancel := context.WithTimeout(context.Background(), 30 * time.Second)
	defer cancel()

	client := w.databaseHelper.GetConnection(ctx)

	stmt, err := client.Prepare(`Update balance
									set amount = amount - $1, 
										update_at = $2
								where id = $3 `)
	if err != nil {
		childLogger.Error().Err(err).Msg("UPDATE statement")
		return false, errors.New(err.Error())
	}
	defer stmt.Close()

	result, err := stmt.Exec(	balance.Amount,
						time.Now(),
						balance.ID,
					)
	if err != nil {
		childLogger.Error().Err(err).Msg("Exec statement")
		return false, errors.New(err.Error())
	}

	rowsAffected, _ := result.RowsAffected()
	childLogger.Debug().Int("rowsAffected : ",int(rowsAffected)).Msg("")

	return true , nil
}