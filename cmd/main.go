package main

import(
	"time"
	"os"
	"strconv"
	"net"
	"io/ioutil"

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

	infoPod					core.InfoPod
	envDB	 				core.DatabaseRDS
	httpAppServerConfig 	core.HttpAppServer
	server					core.Server
	dataBaseHelper 			db_postgre.DatabaseHelper
	repoDB					db_postgre.WorkerRepository
)

func init(){
	log.Debug().Msg("init")
	zerolog.SetGlobalLevel(logLevel)

	// Just for easy test
	envDB.Host = "127.0.0.1" //"host.docker.internal"
	envDB.Port = "5432"
	envDB.Schema = "public"
	envDB.DatabaseName = "postgres"
	//envDB.User  = "postgres"
	//envDB.Password  = "pass123"

	envDB.Db_timeout = 90
	envDB.Postgres_Driver = "postgres"

	server.Port = 5000
	server.ReadTimeout = 60
	server.WriteTimeout = 60
	server.IdleTimeout = 60
	server.CtxTimeout = 60
	//Just for easy test

	httpAppServerConfig.Server = server

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Error().Err(err).Msg("Error to get the POD IP address !!!")
		os.Exit(3)
	}
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				infoPod.IPAddress = ipnet.IP.String()
			}
		}
	}
	infoPod.OSPID = strconv.Itoa(os.Getpid())

	file_user, err := ioutil.ReadFile("/var/pod/secret/username")
    if err != nil {
        log.Error().Err(err).Msg("ERRO FATAL recuperacao secret-user")
		os.Exit(3)
    }
	file_pass, err := ioutil.ReadFile("/var/pod/secret/password")
    if err != nil {
        log.Error().Err(err).Msg("ERRO FATAL recuperacao secret-pass")
		os.Exit(3)
    }
	envDB.User = string(file_user)
	envDB.Password = string(file_pass)

	getEnv()
}

func getEnv() {
	log.Debug().Msg("getEnv")

	if os.Getenv("API_VERSION") !=  "" {
		infoPod.ApiVersion = os.Getenv("API_VERSION")
	}
	if os.Getenv("POD_NAME") !=  "" {
		infoPod.PodName = os.Getenv("POD_NAME")
	}

	if os.Getenv("PORT") !=  "" {
		intVar, _ := strconv.Atoi(os.Getenv("PORT"))
		server.Port = intVar
	}
	if os.Getenv("DB_HOST") !=  "" {
		envDB.Host = os.Getenv("DB_HOST")
	}
	if os.Getenv("DB_PORT") !=  "" {
		envDB.Port = os.Getenv("DB_PORT")
	}

	if os.Getenv("DB_NAME") !=  "" {	
		envDB.DatabaseName = os.Getenv("DB_NAME")
	}
	if os.Getenv("DB_SCHEMA") !=  "" {	
		envDB.Schema = os.Getenv("DB_SCHEMA")
	}
}

func main() {
	log.Debug().Msg("main")
	log.Debug().Interface("",envDB).Msg("getEnv")

	count := 1
	var err error
	for {
		dataBaseHelper, err = db_postgre.NewDatabaseHelper(envDB)
		if err != nil {
			if count < 3 {
				log.Error().Err(err).Msg("Erro na abertura do Database")
			} else {
				log.Error().Err(err).Msg("ERRO FATAL na abertura do Database aborting")
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

	httpAppServerConfig.InfoPod = &infoPod
	httpServer := handler.NewHttpAppServer(httpAppServerConfig)

	httpServer.StartHttpAppServer(httpWorkerAdapter)
}