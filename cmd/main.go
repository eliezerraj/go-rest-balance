package main

import(
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/go-rest-balance/internal/handler"
	"github.com/go-rest-balance/internal/core"
	"github.com/go-rest-balance/internal/service"
	"github.com/go-rest-balance/internal/repository/postgre"
	
)

var(
	logLevel 	= zerolog.DebugLevel
	tableName 	= "BALANCE"
	version 	= "GO CRUD BALANCE 1.0"

	envDB	 			core.DatabaseRDS
	httpAppServerConfig 	core.HttpAppServer
	server					core.Server
	dataBaseHelper 			db_postgre.DatabaseHelper
	repoDB					db_postgre.WorkerRepository
)

func init(){
	log.Debug().Msg("init")
	zerolog.SetGlobalLevel(logLevel)

	envDB.Host = "127.0.0.1" //"host.docker.internal"
	envDB.Port = "5432"
	envDB.Schema = "public"
	envDB.DatabaseName = "postgres"
	envDB.User  = "postgres"
	envDB.Password  = "pass123"
	envDB.Db_timeout = 90
	envDB.Postgres_Driver = "postgres"

	server.Port = 5000
	server.ReadTimeout = 60
	server.WriteTimeout = 60
	server.IdleTimeout = 60
	server.CtxTimeout = 60

	httpAppServerConfig.Server = server
}

func main() {
	log.Debug().Msg("main")

	count := 1
	var err error
	for {
		dataBaseHelper, err = db_postgre.NewDatabaseHelper(envDB)
		if err != nil {
			if count < 3 {
				log.Error().Err(err).Msg("Erro na abertura do Database")
			} else {
				log.Error().Err(err).Msg("EERRO FATAL na abertura do Database aborting")
				panic(err)	
			}
			time.Sleep(3 * time.Second)
			count = count + 1
			continue
		}
		break
	}
	repoDB = db_postgre.NewWorkerRepository(dataBaseHelper)

	workerService := service.NewWorkerService(&repoDB)

	httpWorkerAdapter := handler.NewHttpWorkerAdapter(workerService)
	httpServer := handler.NewHttpAppServer(httpAppServerConfig)

	httpServer.StartHttpAppServer(httpWorkerAdapter)
}