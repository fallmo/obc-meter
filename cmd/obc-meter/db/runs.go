package db

import (
	"context"
	"fmt"
	"time"
)

type Run struct {
	ID            int        `json:"id"`
	StartTime     time.Time  `json:"start_time"`
	EndTime       *time.Time `json:"end_time"`
	AllUids       []string   `json:"all_uids"`
	FailedUids    []string   `json:"failed_uids"`
	ErrorMessages []string   `json:"error_messages"`
	Trigger       string     `json:"trigger"`
}

func OpenRun(trigger string) (*int, error) {
	sql := `INSERT INTO	runs (all_uids, failed_uids, error_messages, trigger)
	VALUES ($1, $2, $3, $4)
	RETURNING id
	`

	var id int

	err := pool.QueryRow(
		context.TODO(),
		sql,
		[]string{},
		[]string{},
		[]string{},
		trigger,
	).Scan(&id)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return &id, nil
}

type CloseRunArgs struct {
	AllUids       []string
	FailedUids    []string
	ErrorMessages []string
}

func CloseRun(id int, args CloseRunArgs) error {
	sql := `UPDATE runs
			SET end_time = NOW(), all_uids = $2, failed_uids = $3, error_messages = $4
			WHERE id = $1
`

	_, err := pool.Exec(context.TODO(), sql, id, args.AllUids, args.FailedUids, args.ErrorMessages)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}
