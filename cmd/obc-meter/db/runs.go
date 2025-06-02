package db

import (
	"context"
	"fmt"
	"strconv"
	"strings"
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

type GetRunsArgs struct {
	Ids      *[]string
	FromTime *time.Time
	ToTime   *time.Time
	Trigger  *string
}

func GetRuns(args GetRunsArgs) (*[]Run, error) {
	whereStatements := []string{}
	sqlVars := []interface{}{}

	if args.Ids != nil {
		whereStatements = append(whereStatements, "id = ANY($"+strconv.Itoa(len(whereStatements)+1)+")")
		sqlVars = append(sqlVars, *args.Ids)
	}

	if args.Trigger != nil {
		whereStatements = append(whereStatements, "trigger = $"+strconv.Itoa(len(whereStatements)+1))
		sqlVars = append(sqlVars, *args.Trigger)
	}

	if args.FromTime != nil {
		whereStatements = append(whereStatements, "end_time > $"+strconv.Itoa(len(whereStatements)+1))
		sqlVars = append(sqlVars, *args.FromTime)
	}

	if args.ToTime != nil {
		whereStatements = append(whereStatements, "start_time < $"+strconv.Itoa(len(whereStatements)+1))
		sqlVars = append(sqlVars, *args.ToTime)
	}

	sql := `
			SELECT id, start_time, end_time, all_uids, failed_uids, error_messages, trigger 
			FROM runs
			`

	if len(whereStatements) > 0 {
		sql = sql + "WHERE " + strings.Join(whereStatements, " AND ")
	}

	rows, err := pool.Query(context.TODO(), sql, sqlVars...)
	defer rows.Close()

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	var runs []Run
	for rows.Next() {
		var run Run
		err := rows.Scan(
			&run.ID,
			&run.StartTime,
			&run.EndTime,
			&run.AllUids,
			&run.FailedUids,
			&run.ErrorMessages,
			&run.Trigger,
		)

		if err != nil {
			fmt.Println(err)
			return nil, err
		}

		runs = append(runs, run)
	}

	return &runs, nil

}
