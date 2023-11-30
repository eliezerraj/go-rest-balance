package main

import(
	"time"
	"os"
	"strconv"
	"net"
	"io/ioutil"
	"context"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
    "github.com/aws/aws-sdk-go-v2/config"

	"github.com/go-rest-balance/internal/adapter/event"
	"github.com/go-rest-balance/internal/handler"
	"github.com/go-rest-balance/internal/core"
	"github.com/go-rest-balance/internal/service"
	"github.com/go-rest-balance/internal/repository/postgre"
	
)

var(
	logLevel 	= 	zerolog.DebugLevel
	tableName 	= 	"BALANCE"
	version 	= 	"GO CRUD BALANCE 1.0"
	noAZ		=	true // set only if you get to split the xray trace per AZ

	infoPod					core.InfoPod
	envDB	 				core.DatabaseRDS
	envKafka				core.KafkaConfig
	httpAppServerConfig 	core.HttpAppServer
	server					core.Server
	dataBaseHelper 			db_postgre.DatabaseHelper
	repoDB					db_postgre.WorkerRepository
)

func loadLocalEnv(){
	log.Debug().Msg("loadLocalEnv")
	zerolog.SetGlobalLevel(logLevel)

	// LOCAL TEST
	// ------------------------------------------------------------
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

	envKafka.KafkaConfigurations.Username = "admin"
	envKafka.KafkaConfigurations.Password = "admin"
	envKafka.KafkaConfigurations.Protocol = "PLAINTEXT"
	envKafka.KafkaConfigurations.Mechanisms = "PLAINTEXT"

	envKafka.KafkaConfigurations.Clientid = "GO-REST-BALANCE"
	envKafka.KafkaConfigurations.Brokers1 = "b-1.mskarchtest02.9vkh4b.c3.kafka.us-east-2.amazonaws.com:9092"
	envKafka.KafkaConfigurations.Brokers2 = "b-2.mskarchtest02.9vkh4b.c3.kafka.us-east-2.amazonaws.com:9092"

	envKafka.KafkaConfigurations.Partition = 1
	envKafka.KafkaConfigurations.ReplicationFactor = 1
	// ------------------------------------------------------------

}

func init(){
	log.Debug().Msg("init")
	zerolog.SetGlobalLevel(logLevel)
	
	loadLocalEnv()

	// Get Database Secrets
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

	// Load info pod
	// Get IP
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

	// Get AZ only if localtest is true
	if (noAZ != true) {
		cfg, err := config.LoadDefaultConfig(context.TODO())
		if err != nil {
			log.Error().Err(err).Msg("ERRO FATAL get Context !!!")
			os.Exit(3)
		}
		client := imds.NewFromConfig(cfg)
		response, err := client.GetInstanceIdentityDocument(context.TODO(), &imds.GetInstanceIdentityDocumentInput{})
		if err != nil {
			log.Error().Err(err).Msg("Unable to retrieve the region from the EC2 instance !!!")
			os.Exit(3)
		}
		infoPod.AvailabilityZone = response.AvailabilityZone	
	} else {
		infoPod.AvailabilityZone = "LOCALHOST_NO_AZ"
	}
	// Load info pod
	infoPod.Database = &envDB
	infoPod.Kafka	 = &envKafka
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

	if os.Getenv("KAFKA_USER") !=  "" {
		envKafka.KafkaConfigurations.Username = os.Getenv("KAFKA_USER")
	}
	if os.Getenv("KAFKA_PASSWORD") !=  "" {
		envKafka.KafkaConfigurations.Password = os.Getenv("KAFKA_PASSWORD")
	}
	if os.Getenv("KAFKA_PROTOCOL") !=  "" {
		envKafka.KafkaConfigurations.Protocol = os.Getenv("KAFKA_PROTOCOL")
	}
	if os.Getenv("KAFKA_MECHANISM") !=  "" {
		envKafka.KafkaConfigurations.Mechanisms = os.Getenv("KAFKA_MECHANISM")
	}
	if os.Getenv("KAFKA_CLIENT_ID") !=  "" {
		envKafka.KafkaConfigurations.Clientid = os.Getenv("KAFKA_CLIENT_ID")
	}
	if os.Getenv("KAFKA_BROKER_1") !=  "" {
		envKafka.KafkaConfigurations.Brokers1 = os.Getenv("KAFKA_BROKER_1")
	}
	if os.Getenv("KAFKA_BROKER_2") !=  "" {
		envKafka.KafkaConfigurations.Brokers2 = os.Getenv("KAFKA_BROKER_2")
	}
	if os.Getenv("KAFKA_BROKER_3") !=  "" {
		envKafka.KafkaConfigurations.Brokers3 = os.Getenv("KAFKA_BROKER_3")
	}

	if os.Getenv("KAFKA_PARTITION") !=  "" {
		intVar, _ := strconv.Atoi(os.Getenv("KAFKA_PARTITION"))
		envKafka.KafkaConfigurations.Partition = intVar
	}
	if os.Getenv("KAFKA_REPLICATION") !=  "" {
		intVar, _ := strconv.Atoi(os.Getenv("KAFKA_REPLICATION"))
		envKafka.KafkaConfigurations.ReplicationFactor = intVar
	}

	if os.Getenv("NO_AZ") == "false" {	
		noAZ = false
	} else {
		noAZ = true
	}
}

func main() {
	log.Debug().Msg("main")
	log.Debug().Interface("",envDB).Msg("getEnv")
	log.Debug().Msg("--------")
	log.Debug().Interface("",server).Msg("server")
	log.Debug().Msg("--------")

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration( server.ReadTimeout ) * time.Second)
	defer cancel()

	// Open Database
	count := 1
	var err error
	for {
		dataBaseHelper, err = db_postgre.NewDatabaseHelper(ctx, envDB)
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
	
	// Setup workload
	httpAppServerConfig.Server = server
	repoDB = db_postgre.NewWorkerRepository(dataBaseHelper)

	producerWorker, err := event.NewProducerWorker(&envKafka)
	if err != nil {
		log.Error().Err(err).Msg("Erro na abertura do Kafka")
	}

	workerService := service.NewWorkerService(&repoDB,producerWorker)
	httpWorkerAdapter := handler.NewHttpWorkerAdapter(workerService)

	httpAppServerConfig.InfoPod = &infoPod
	httpServer := handler.NewHttpAppServer(httpAppServerConfig)

	httpServer.StartHttpAppServer(ctx, httpWorkerAdapter)
}