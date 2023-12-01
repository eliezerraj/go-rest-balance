package service

import (
	"context"
	"time"
	"github.com/rs/zerolog/log"

	"github.com/go-rest-balance/internal/erro"
	"github.com/go-rest-balance/internal/core"
	"github.com/go-rest-balance/internal/repository/postgre"
	"github.com/go-rest-balance/internal/adapter/event"
	"github.com/aws/aws-xray-sdk-go/xray"

)

var childLogger = log.With().Str("service", "service").Logger()

type WorkerService struct {
	workerRepository 		*db_postgre.WorkerRepository
	producerWorker			*event.ProducerWorker
}

func NewWorkerService(	workerRepository 	*db_postgre.WorkerRepository,
						producerWorker		*event.ProducerWorker) *WorkerService{
	childLogger.Debug().Msg("NewWorkerService")

	return &WorkerService{
		workerRepository:	workerRepository,
		producerWorker: 	producerWorker,
	}
}

func (s WorkerService) SetSessionVariable(ctx context.Context, userCredential string) (bool, error){
	childLogger.Debug().Msg("SetSessionVariable")

	res, err := s.workerRepository.SetSessionVariable(ctx, userCredential)
	if err != nil {
		return false, err
	}

	return res, nil
}

func (s WorkerService) Add(ctx context.Context, balance core.Balance) (*core.Balance, error){
	childLogger.Debug().Msg("Add")

	_, root := xray.BeginSubsegment(ctx, "Service.Add")
	defer func() {
		root.Close(nil)
	}()

	res, err := s.workerRepository.Add(ctx, balance)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s WorkerService) Get(ctx context.Context,balance core.Balance) (*core.Balance, error){
	childLogger.Debug().Msg("Get")

	_, root := xray.BeginSubsegment(ctx, "Service.Get")
	defer func() {
		root.Close(nil)
	}()

	res, err := s.workerRepository.Get(ctx, balance)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s WorkerService) Update(ctx context.Context, balance core.Balance) (*core.Balance, error){
	childLogger.Debug().Msg("Update")

	_, root := xray.BeginSubsegment(ctx, "Service.Update")
	defer func() {
		root.Close(nil)
	}()

	res_balance, err := s.workerRepository.Get(ctx, balance)
	if err != nil {
		return nil, err
	}

	balance.ID = res_balance.ID
	isUpdated, err := s.workerRepository.Update(ctx, balance)
	if err != nil {
		return nil, err
	}
	if (isUpdated == false) {
		return nil, erro.ErrUpdate
	}

	res_balance, err = s.workerRepository.Get(ctx, balance)
	if err != nil {
		return nil, err
	}
	return res_balance, nil
}

func (s WorkerService) Delete(ctx context.Context,balance core.Balance) (bool, error){
	childLogger.Debug().Msg("Delete")

	_, root := xray.BeginSubsegment(ctx, "Service.Delete")
	defer func() {
		root.Close(nil)
	}()

	res_balance, err := s.workerRepository.Get(ctx,balance)
	if err != nil {
		return false, err
	}

	balance.ID = res_balance.ID
	isDelete, err := s.workerRepository.Delete(ctx, balance)
	if err != nil {
		return false, err
	}
	if (isDelete == false) {
		return false, erro.ErrDelete
	}
	return true, nil
}

func (s WorkerService) List(ctx context.Context, balance core.Balance) (*[]core.Balance, error){
	childLogger.Debug().Msg("List")

	_, root := xray.BeginSubsegment(ctx, "Service.List")
	defer root.Close(nil)

	res, err := s.workerRepository.List(ctx, balance)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s WorkerService) Sum(ctx context.Context,balance core.Balance) (*core.Balance, error){
	childLogger.Debug().Msg("Sum")

	_, root := xray.BeginSubsegment(ctx, "Service.Sum")
	defer func() {
		root.Close(nil)
	}()

	res_balance, err := s.workerRepository.Get(ctx, balance)
	if err != nil {
		return nil, err
	}

	balance.ID = res_balance.ID
	isUpdated, err := s.workerRepository.Sum(ctx, balance)
	if err != nil {
		return nil, err
	}
	if (isUpdated == false) {
		return nil, erro.ErrUpdate
	}

	res_balance, err = s.workerRepository.Get(ctx, balance)
	if err != nil {
		return nil, err
	}

	eventData := core.EventData{res_balance}

	event := core.Event{
		ID: 1,
		EventDate: time.Now(),
		EventType: "topic.x",
		EventData:	&eventData,	
	}

	err = s.producerWorker.Producer(ctx, event)
	if err != nil {
		return nil, err
	}

	return res_balance, nil
}
