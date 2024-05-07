package main

import (
	"database/sql"
	"log"
	"math/rand"
	"time"

	dbmigration "github.com/Just-Goo/grpc-go-server/db"
	"github.com/Just-Goo/grpc-go-server/internal/adapter/database"
	mygrpc "github.com/Just-Goo/grpc-go-server/internal/adapter/grpc"
	app "github.com/Just-Goo/grpc-go-server/internal/application"
	"github.com/Just-Goo/grpc-go-server/internal/application/domain/bank"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	log.SetFlags(0)
	log.SetOutput(logWriter{})

	db, err := sql.Open("pgx", "postgres://root:password@localhost:5432/grpc?sslmode=disable")
	if err != nil {
		log.Fatalln("can't connect to database", err)
	}

	dbmigration.Migrate(db) // run database migration

	// create new database adapter instance
	dbAdapter, err := database.NewDatabaseAdapter(db)

	if err != nil {
		log.Fatalln("can't create database adapter", err)
	}

	hs := &app.HelloService{}
	bs := app.NewBankService(dbAdapter)

	go generateExchangeRates(bs, "USD", "IDR", 5 * time.Second) // launch a separate goroutine and generate exchange rates every 5 second

	grpcAdapter := mygrpc.NewGrpcAdapter(hs, bs, 9090)

	grpcAdapter.Run()
}

// func runDummyData(da *database.DatabaseAdapter) {
// 	now := time.Now()

// 	uuid, _ := da.Save(
// 		&database.DummyOrm{
// 			UserID:    uuid.New(),
// 			Username:  "Dave " + now.Format("15:04:05"),
// 			CreatedAt: now,
// 			UpdatedAt: now,
// 		},
// 	)

// 	res, _ := da.GetByUUID(&uuid)

// 	log.Println("res : ", res)
// }

func generateExchangeRates(bs *app.BankService, fromCurrency, toCurrency string, duration time.Duration) {
	ticker := time.NewTicker(duration)

	for range ticker.C {
		now := time.Now()
		validFrom := now.Truncate(time.Second).Add(3 * time.Second)
		validTo := validFrom.Add(duration).Add(-1 * time.Millisecond)

		dummyRate := bank.ExchangeRate{
			FromCurrency:       fromCurrency,
			ToCurrency:         toCurrency,
			ValidFromTimestamp: validFrom,
			ValidToTimestamp:   validTo,
			Rate:               2000 + float64(rand.Intn(300)),
		}

		bs.CreateExchangeRate(dummyRate)

	}
}
