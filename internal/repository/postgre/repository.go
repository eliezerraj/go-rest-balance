package db_postgre

import (
	"context"
	"time"
	"errors"

	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"

	"github.com/go-rest-balance/internal/core"
	"github.com/go-rest-balance/internal/erro"
	"github.com/aws/aws-xray-sdk-go/xray"

)

var childLogger = log.With().Str("repository", "WorkerRepository").Logger()

type WorkerRepository struct {
	databaseHelper DatabaseHelper
}

func NewWorkerRepository(databaseHelper DatabaseHelper) WorkerRepository {
	childLogger.Debug().Msg("NewWorkerRepository")
	return WorkerRepository{
		databaseHelper: databaseHelper,
	}
}

func (w WorkerRepository) SetSessionVariable(ctx context.Context,userCredential string) (bool, error) {
	childLogger.Debug().Msg("++++++++++++++++++++++++++++++++")
	childLogger.Debug().Msg("SetSessionVariable")
	childLogger.Debug().Msg("++++++++++++++++++++++++++++++++")

	client := w.databaseHelper.GetConnection()
	
	stmt, err := client.Prepare("SET sess.user_credential to '" + userCredential+ "'")
	if err != nil {
		childLogger.Error().Err(err).Msg("SET SESSION statement ERROR")
		return false, errors.New(err.Error())
	}

	_, err = stmt.ExecContext(ctx)
	if err != nil {
		childLogger.Error().Err(err).Msg("Exec statement")
		return false, errors.New(err.Error())
	}

	return true, nil
}

func (w WorkerRepository) GetSessionVariable(ctx context.Context) (string, error) {
	childLogger.Debug().Msg("++++++++++++++++++++++++++++++++")
	childLogger.Debug().Msg("GetSessionVariable")
	childLogger.Debug().Msg("++++++++++++++++++++++++++++++++")

	client := w.databaseHelper.GetConnection()

	var res_balance string
	rows, err := client.QueryContext(ctx, "SELECT current_setting('sess.user_credential')" )
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

func (w WorkerRepository) Ping(ctx context.Context) (bool, error) {
	childLogger.Debug().Msg("++++++++++++++++++++++++++++++++")
	childLogger.Debug().Msg("Ping")
	childLogger.Debug().Msg("++++++++++++++++++++++++++++++++")

	client := w.databaseHelper.GetConnection()

	err := client.PingContext(ctx)
	if err != nil {
		return false, errors.New(err.Error())
	}

	return true, nil
}

func (w WorkerRepository) Add(ctx context.Context, balance core.Balance) (*core.Balance, error){
	childLogger.Debug().Msg("Add")

	_, root := xray.BeginSubsegment(ctx, "SQL.Add-Balance")
	defer root.Close(nil)

	client := w.databaseHelper.GetConnection()

	userLastUpdate, _ := w.GetSessionVariable(ctx)

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
	
	_, err = stmt.ExecContext(	ctx,	
								balance.AccountID, 
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

func (w WorkerRepository) Get(ctx context.Context, balance core.Balance) (*core.Balance, error){
	childLogger.Debug().Msg("Get")

	_, root := xray.BeginSubsegment(ctx, "SQL.Get-Balance")
	defer root.Close(nil)

	client := w.databaseHelper.GetConnection()

	result_query := core.Balance{}
	rows, err := client.QueryContext(ctx, `SELECT id, account_id, person_id, currency, amount, create_at, update_at, tenant_id, user_last_update FROM balance WHERE account_id =$1`, balance.AccountID)
	if err != nil {
		childLogger.Error().Err(err).Msg("Query statement")
		return nil, errors.New(err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan( &result_query.ID, 
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

func (w WorkerRepository) Update(ctx context.Context, balance core.Balance) (bool, error){
	childLogger.Debug().Msg("Update...")
	//childLogger.Debug().Interface("balance : ", balance).Msg("balance")

	_, root := xray.BeginSubsegment(ctx, "SQL.Update-Balance")
	defer root.Close(nil)

	client := w.databaseHelper.GetConnection()

	userLastUpdate, _ := w.GetSessionVariable(ctx)

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

	result, err := stmt.ExecContext(ctx,	
									balance.AccountID, 
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

func (w WorkerRepository) Delete(ctx context.Context, balance core.Balance) (bool, error){
	childLogger.Debug().Msg("Delete")

	_, root := xray.BeginSubsegment(ctx, "SQL.Update-Balance")
	defer root.Close(nil)
	
	client := w.databaseHelper.GetConnection()

	stmt, err := client.Prepare(`Delete from balance where id = $1 `)
	if err != nil {
		childLogger.Error().Err(err).Msg("DELETE statement")
		return false, errors.New(err.Error())
	}
	defer stmt.Close()

	result, err := stmt.ExecContext(ctx,balance.ID )
	if err != nil {
		childLogger.Error().Err(err).Msg("Exec statement")
		return false, errors.New(err.Error())
	}

	rowsAffected, _ := result.RowsAffected()
	childLogger.Debug().Int("rowsAffected : ",int(rowsAffected)).Msg("")

	return true , nil
}

func (w WorkerRepository) List(ctx context.Context, balance core.Balance) (*[]core.Balance, error){
	childLogger.Debug().Msg("List")

	_, root := xray.BeginSubsegment(ctx, "SQL.List-Balance")
	defer root.Close(nil)

	client:= w.databaseHelper.GetConnection()
	
	result_query := core.Balance{}
	balance_list := []core.Balance{}
	rows, err := client.QueryContext(ctx, `SELECT id, account_id, person_id, currency, amount, create_at, update_at, tenant_id, user_last_update FROM balance WHERE person_id =$1`, balance.PersonID)
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

func (w WorkerRepository) Sum(ctx context.Context, balance core.Balance) (bool, error){
	childLogger.Debug().Msg("Sum")

	_, root := xray.BeginSubsegment(ctx, "SQL.Sum-Balance")
	defer root.Close(nil)

	client := w.databaseHelper.GetConnection()

	stmt, err := client.Prepare(`Update balance
									set amount = amount + $1, 
										update_at = $2
								where id = $3 `)
	if err != nil {
		childLogger.Error().Err(err).Msg("UPDATE statement")
		return false, errors.New(err.Error())
	}
	defer stmt.Close()

	result, err := stmt.ExecContext(ctx,
									balance.Amount,
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
