package core

import (
	"time"

)

type DatabaseRDS struct {
    Host 				string `json:"host"`
    Port  				string `json:"port"`
	Schema				string `json:"schema"`
	DatabaseName		string `json:"databaseName"`
	User				string `json:"user"`
	Password			string `json:"password"`
	Db_timeout			int	`json:"db_timeout"`
	Postgres_Driver		string `json:"postgres_driver"`
}

type HttpAppServer struct {
	AppInfo 	*AppInfo 		`json:"app_info"`
	Server     	Server     		`json:"server"`
}

type AppInfo struct {
	Name 				string `json:"name"`
	Description 		string `json:"description"`
	Version 			string `json:"version"`
	OSPID				string `json:"os_pid"`
	IpAdress			string `json:"ip_adress"`
}

type Server struct {
	Port 			int `json:"port"`
	ReadTimeout		int `json:"readTimeout"`
	WriteTimeout	int `json:"writeTimeout"`
	IdleTimeout		int `json:"idleTimeout"`
	CtxTimeout		int `json:"ctxTimeout"`
}

type Balance struct {
	ID				int		`json:"id,omitempty"`
	AccountID		string	`json:"account_id,omitempty"`
	PersonID		string  `json:"person_id,omitempty"`
	Currency		string  `json:"currency,omitempty"`
	Amount			float64 `json:"amount,omitempty"`
	CreateAt		time.Time 	`json:"create_at,omitempty"`
	UpdateAt		*time.Time 	`json:"update_at,omitempty"`
	TenantID		string  `json:"tenant_id,omitempty"`
}