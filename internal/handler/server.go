package handler

import (
	"time"
	"encoding/json"
	"net/http"
	"strconv"
	"os"
	"os/signal"
	"syscall"
	"context"
	"fmt"

	"github.com/gorilla/mux"

	"github.com/go-rest-balance/internal/service"
	"github.com/go-rest-balance/internal/core"
	"github.com/aws/aws-xray-sdk-go/xray"

)

type HttpWorkerAdapter struct {
	workerService 	*service.WorkerService
}

func NewHttpWorkerAdapter(workerService *service.WorkerService) *HttpWorkerAdapter {
	childLogger.Debug().Msg("NewHttpWorkerAdapter")
	return &HttpWorkerAdapter{
		workerService: workerService,
	}
}

type HttpServer struct {
	start 			time.Time
	httpAppServer 	core.HttpAppServer
}

func NewHttpAppServer(httpAppServer core.HttpAppServer) HttpServer {
	childLogger.Debug().Msg("NewHttpAppServer")

	return HttpServer{	start: time.Now(), 
						httpAppServer: httpAppServer,
					}
}

func (h HttpServer) StartHttpAppServer(ctx context.Context, httpWorkerAdapter *HttpWorkerAdapter) {
	childLogger.Info().Msg("StartHttpAppServer")
		
	myRouter := mux.NewRouter().StrictSlash(true)

	myRouter.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		childLogger.Debug().Msg("/")
		json.NewEncoder(rw).Encode(h.httpAppServer)
	})
	myRouter.Use(MiddleWareHandlerHeader)

	myRouter.HandleFunc("/info", func(rw http.ResponseWriter, req *http.Request) {
		childLogger.Debug().Msg("/info")
		json.NewEncoder(rw).Encode(h.httpAppServer)
	})
	myRouter.Use(MiddleWareHandlerHeader)
	
	health := myRouter.Methods(http.MethodGet, http.MethodOptions).Subrouter()
    health.HandleFunc("/health", httpWorkerAdapter.Health)

	live := myRouter.Methods(http.MethodGet, http.MethodOptions).Subrouter()
    live.HandleFunc("/live", httpWorkerAdapter.Live)

	header := myRouter.Methods(http.MethodGet, http.MethodOptions).Subrouter()
    header.HandleFunc("/header", httpWorkerAdapter.Header)
	header.Use(MiddleWareHandlerHeader)

	addBalance := myRouter.Methods(http.MethodPost, http.MethodOptions).Subrouter()
    //addBalance.HandleFunc("/add", httpWorkerAdapter.Add)
	addBalance.Handle("/add", 
						xray.Handler(xray.NewFixedSegmentNamer(fmt.Sprintf("%s%s%s", "balance:", h.httpAppServer.InfoPod.AvailabilityZone, ".add")), 
						http.HandlerFunc(httpWorkerAdapter.Add),
						),
	)
	addBalance.Use(httpWorkerAdapter.DecoratorDB)

	getBalance := myRouter.Methods(http.MethodGet, http.MethodOptions).Subrouter()
    //getBalance.HandleFunc("/get/{id}", httpWorkerAdapter.Get)
	getBalance.Handle("/get/{id}",
						xray.Handler(xray.NewFixedSegmentNamer(fmt.Sprintf("%s%s%s", "balance:", h.httpAppServer.InfoPod.AvailabilityZone, ".getId")),
						http.HandlerFunc(httpWorkerAdapter.Get),
						),
	)
	getBalance.Use(MiddleWareHandlerHeader)

	updateBalance := myRouter.Methods(http.MethodPost, http.MethodOptions).Subrouter()
    //updateBalance.HandleFunc("/update/{id}", httpWorkerAdapter.Update)
	updateBalance.Handle("/update/{id}", 
						xray.Handler(xray.NewFixedSegmentNamer(fmt.Sprintf("%s%s%s", "balance:", h.httpAppServer.InfoPod.AvailabilityZone, ".updateId")),
						//xray.Handler(xray.NewFixedSegmentNamer("go-rest-balance.updateId"), 
						http.HandlerFunc(httpWorkerAdapter.Update),
						),
	)
	updateBalance.Use(httpWorkerAdapter.DecoratorDB)

	listBalance := myRouter.Methods(http.MethodGet, http.MethodOptions).Subrouter()
    //listBalance.HandleFunc("/list/{id}", httpWorkerAdapter.List)
	listBalance.Handle("/list/{id}", 
						xray.Handler(xray.NewFixedSegmentNamer(fmt.Sprintf("%s%s%s", "balance:", h.httpAppServer.InfoPod.AvailabilityZone, ".listId")),
						//xray.Handler(xray.NewFixedSegmentNamer("go-rest-balance.listId"), 
						http.HandlerFunc(httpWorkerAdapter.List),
						),
	)
	listBalance.Use(MiddleWareHandlerHeader)

	sumBalance := myRouter.Methods(http.MethodPost, http.MethodOptions).Subrouter()
    //sumBalance.HandleFunc("/sum", httpWorkerAdapter.Sum)
	sumBalance.Handle("/sum", 
						xray.Handler(xray.NewFixedSegmentNamer(fmt.Sprintf("%s%s%s", "balance:", h.httpAppServer.InfoPod.AvailabilityZone, ".Sum")),
						//xray.Handler(xray.NewFixedSegmentNamer("go-rest-balance.Sum"), 
						http.HandlerFunc(httpWorkerAdapter.Sum),
						),
	)
	sumBalance.Use(httpWorkerAdapter.DecoratorDB)
	
	deleteBalance := myRouter.Methods(http.MethodDelete, http.MethodOptions).Subrouter()
    deleteBalance.HandleFunc("/delete/{id}", httpWorkerAdapter.Delete)
	deleteBalance.Use(MiddleWareHandlerHeader)

	srv := http.Server{
		Addr:         ":" +  strconv.Itoa(h.httpAppServer.Server.Port),      	
		Handler:      myRouter,                	          
		ReadTimeout:  time.Duration(h.httpAppServer.Server.ReadTimeout) * time.Second,   
		WriteTimeout: time.Duration(h.httpAppServer.Server.WriteTimeout) * time.Second,  
		IdleTimeout:  time.Duration(h.httpAppServer.Server.IdleTimeout) * time.Second, 
	}

	childLogger.Info().Str("Service Port : ", strconv.Itoa(h.httpAppServer.Server.Port)).Msg("Service Port")

	go func() {
		err := srv.ListenAndServe()
		if err != nil {
			childLogger.Error().Err(err).Msg("Cancel http mux server !!!")
		}
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	<-ch

	if err := srv.Shutdown(ctx); err != nil && err != http.ErrServerClosed {
		childLogger.Error().Err(err).Msg("WARNING Dirty Shutdown !!!")
		return
	}
	childLogger.Info().Msg("Stop Done !!!!")
}