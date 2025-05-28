package utils

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

var pool *pgxpool.Pool

func ConnectPostgres() {
	uri := os.Getenv("POSTGRES_URI")
	if uri == "" {
		log.Fatal("Missing 'POSTGRES_URI' environment variable.")
	}

	var err error
	ctx := context.Background()

	pool, err = pgxpool.New(ctx, uri)
	if err != nil {
		fmt.Println(err)
		log.Fatal("Unable to connect to database")
	}

	if err := pool.Ping(ctx); err != nil {
		fmt.Println(err)
		log.Fatal("Unable to ping database")
	}

	log.Println("Successfully connected to PG database")

	// record, err := appendBucketUsageRecord(appendBucketUsageRecordArgs{
	// 	bucketUid:    "692e149b-4393-4aa8-8b54-72dfe267d203",
	// 	objectsCount: 1,
	// 	bytesTotal:   1000,
	// })

	// if err != nil {
	// 	fmt.Println(err)
	// 	fmt.Println("Failed to update record")
	// } else {
	// 	fmt.Println(record)
	// }

}
