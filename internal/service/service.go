package service

import (
	"github.com/rs/zerolog/log"

	"github.com/go-rest-balance/internal/erro"
	"github.com/go-rest-balance/internal/core"
	"github.com/go-rest-balance/internal/repository/postgre"

)

var childLogger = log.With().Str("service", "service").Logger()

type WorkerService struct {
	workerRepository 		*db_postgre.WorkerRepository
}

func NewWorkerService(workerRepository *db_postgre.WorkerRepository) *WorkerService{
	childLogger.Debug().Msg("NewWorkerService")

	return &WorkerService{
		workerRepository:	workerRepository,
	}
}

func (s WorkerService) SetSessionVariable(userCredential string) (bool, error){
	childLogger.Debug().Msg("SetSessionVariable")

	res, err := s.workerRepository.SetSessionVariable(userCredential)
	if err != nil {
		return false, err
	}

	return res, nil
}

func (s WorkerService) Add(balance core.Balance) (*core.Balance, error){
	childLogger.Debug().Msg("Add")

	res, err := s.workerRepository.Add(balance)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s WorkerService) Get(balance core.Balance) (*core.Balance, error){
	childLogger.Debug().Msg("Get")

	res, err := s.workerRepository.Get(balance)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s WorkerService) Update(balance core.Balance) (*core.Balance, error){
	childLogger.Debug().Msg("Update")

	res_balance, err := s.workerRepository.Get(balance)
	if err != nil {
		return nil, err
	}

	balance.ID = res_balance.ID
	isUpdated, err := s.workerRepository.Update(balance)
	if err != nil {
		return nil, err
	}
	if (isUpdated == false) {
		return nil, erro.ErrUpdate
	}

	res_balance, err = s.workerRepository.Get(balance)
	if err != nil {
		return nil, err
	}
	return res_balance, nil
}

func (s WorkerService) Delete(balance core.Balance) (bool, error){
	childLogger.Debug().Msg("Delete")

	res_balance, err := s.workerRepository.Get(balance)
	if err != nil {
		return false, err
	}

	balance.ID = res_balance.ID
	isDelete, err := s.workerRepository.Delete(balance)
	if err != nil {
		return false, err
	}
	if (isDelete == false) {
		return false, erro.ErrDelete
	}
	return true, nil
}

func (s WorkerService) List(balance core.Balance) (*[]core.Balance, error){
	childLogger.Debug().Msg("List")

	res, err := s.workerRepository.List(balance)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s WorkerService) Sum(balance core.Balance) (*core.Balance, error){
	childLogger.Debug().Msg("Sum")

	res_balance, err := s.workerRepository.Get(balance)
	if err != nil {
		return nil, err
	}

	balance.ID = res_balance.ID
	isUpdated, err := s.workerRepository.Sum(balance)
	if err != nil {
		return nil, err
	}
	if (isUpdated == false) {
		return nil, erro.ErrUpdate
	}

	res_balance, err = s.workerRepository.Get(balance)
	if err != nil {
		return nil, err
	}
	return res_balance, nil
}

func (s WorkerService) Minus(balance core.Balance) (*core.Balance, error){
	childLogger.Debug().Msg("Minus")

	res_balance, err := s.workerRepository.Get(balance)
	if err != nil {
		return nil, err
	}

	balance.ID = res_balance.ID
	isUpdated, err := s.workerRepository.Minus(balance)
	if err != nil {
		return nil, err
	}
	if (isUpdated == false) {
		return nil, erro.ErrUpdate
	}

	res_balance, err = s.workerRepository.Get(balance)
	if err != nil {
		return nil, err
	}
	return res_balance, nil
}