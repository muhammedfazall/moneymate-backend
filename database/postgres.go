package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/moneymate-2026/moneymate-backend/shared/config"
)

var Pool *pgxpool.Pool

func ConnectDB(cfg *config.Config)(*pgxpool.Pool,error){

	poolConfig,err:=pgxpool.ParseConfig(cfg.Database.DSN)
	if err !=nil{
		log.Fatal("Failed to parse database config")
	}
	poolConfig.MaxConns = int32(cfg.Database.MaxOpenConns)
	poolConfig.MinConns = int32(cfg.Database.MinOpenConns)
	poolConfig.MaxConnIdleTime = cfg.Database.MaxIdleTime
	poolConfig.MaxConnLifetime = cfg.Database.MaxConnLifetime

	pool,err:=pgxpool.NewWithConfig(context.Background(),poolConfig)
	if err!=nil{
		return nil, fmt.Errorf("Unable to create connection pool:%w ",err)
	}
	ctx,cancel:=context.WithTimeout(context.Background(),time.Second * 5)
	defer cancel()

	if err:=pool.Ping(ctx); err!=nil{
		return nil, fmt.Errorf("unable to connect database: %w",err)
	}
	log.Println("Database connected succesfully")
	return pool,nil
}