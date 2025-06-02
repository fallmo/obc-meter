package db

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Record struct {
	ID           int        `json:"id"`
	BucketUid    string     `json:"bucket_uid"`
	PeriodStart  time.Time  `json:"period_start"`
	PeriodEnd    *time.Time `json:"period_end"`
	ObjectsCount uint64     `json:"objects_count"`
	BytesTotal   uint64     `json:"bytes_count"`
	RunId        int        `json:"run_id"`
}

func GetBucketCurrentRecord(bucketUid string) (*Record, error) {
	sql := `
		SELECT id, bucket_uid, period_start, period_end, objects_count, bytes_total, run_id
		FROM records
		WHERE bucket_uid = $1 AND period_end IS NULL
		LIMIT 1
	`

	var ID int
	var BucketUid string
	var PeriodStart time.Time
	var PeriodEnd *time.Time
	var ObjectsCount uint64
	var BytesTotal uint64
	var RunId int

	err := pool.QueryRow(context.TODO(), sql, bucketUid).Scan(
		&ID,
		&BucketUid,
		&PeriodStart,
		&PeriodEnd,
		&ObjectsCount,
		&BytesTotal,
		&RunId,
	)

	if err != nil {
		// not a real error
		if strings.Contains(err.Error(), "no rows in result set") {
			return nil, nil
		}

		return nil, err
	}

	record := Record{
		ID:           ID,
		BucketUid:    BucketUid,
		PeriodStart:  PeriodStart,
		PeriodEnd:    PeriodEnd,
		ObjectsCount: ObjectsCount,
		BytesTotal:   BytesTotal,
		RunId:        RunId,
	}

	return &record, nil
}

type AppendBucketUsageRecordArgs struct {
	BucketUid    string
	ObjectsCount uint64
	BytesTotal   uint64
	RunId        int
}

func AppendBucketUsageRecord(args AppendBucketUsageRecordArgs) (*Record, error) {
	ctx := context.Background()
	tx, err := pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, "UPDATE records SET period_end = NOW() WHERE bucket_uid = $1 AND period_end IS NULL", args.BucketUid)
	if err != nil {
		fmt.Println("Failed to close previous records")
		tx.Rollback(ctx)
		return nil, err
	}

	var id int
	var period_start time.Time

	err = tx.QueryRow(
		ctx,
		`INSERT INTO records (bucket_uid, objects_count, bytes_total, run_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id, period_start`,
		args.BucketUid,
		args.ObjectsCount,
		args.BytesTotal,
		args.RunId,
	).Scan(&id, &period_start)

	if err != nil {
		fmt.Println("Failed to insert new record")
		tx.Rollback(ctx)
		return nil, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		fmt.Println("Failed to commit transaction")
		tx.Rollback(ctx)
		return nil, err
	}

	record := Record{
		ID:           id,
		BucketUid:    args.BucketUid,
		PeriodStart:  period_start,
		PeriodEnd:    nil,
		ObjectsCount: args.ObjectsCount,
		BytesTotal:   args.BytesTotal,
		RunId:        args.RunId,
	}

	return &record, nil
}

type GetRecordsArgs struct {
	Uids       *[]string
	FromPeriod *time.Time
	ToPeriod   *time.Time
	RunIds     *[]string
}

func GetUsageRecords(args GetRecordsArgs) (*[]Record, error) {
	whereStatements := []string{}
	sqlVars := []interface{}{}

	if args.Uids != nil {
		whereStatements = append(whereStatements, "bucket_uid = ANY($"+strconv.Itoa(len(whereStatements)+1)+")")
		sqlVars = append(sqlVars, *args.Uids)
	}

	if args.RunIds != nil {
		whereStatements = append(whereStatements, "run_id = ANY($"+strconv.Itoa(len(whereStatements)+1)+")")
		sqlVars = append(sqlVars, *args.RunIds)
	}

	if args.FromPeriod != nil {
		whereStatements = append(whereStatements, "(period_end > "+"$"+strconv.Itoa(len(whereStatements)+1)+" OR period_end IS NULL)")
		sqlVars = append(sqlVars, args.FromPeriod)
	}

	if args.ToPeriod != nil {
		whereStatements = append(whereStatements, "period_start < "+"$"+strconv.Itoa(len(whereStatements)+1))
		sqlVars = append(sqlVars, args.ToPeriod)
	}

	sql := `
		SELECT id, bucket_uid, period_start, period_end, objects_count, bytes_total, run_id
		FROM records
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

	var records []Record
	for rows.Next() {
		var record Record
		err := rows.Scan(
			&record.ID,
			&record.BucketUid,
			&record.PeriodStart,
			&record.PeriodEnd,
			&record.ObjectsCount,
			&record.BytesTotal,
			&record.RunId,
		)

		if err != nil {
			fmt.Println(err)
			return nil, err
		}

		if args.FromPeriod != nil && record.PeriodStart.Before(*args.FromPeriod) {
			record.PeriodStart = *args.FromPeriod
		}

		if args.ToPeriod != nil && (record.PeriodEnd == nil || record.PeriodEnd.After(*args.ToPeriod)) {
			record.PeriodEnd = *&args.ToPeriod
		}

		records = append(records, record)
	}

	return &records, nil
}

type GetBucketRecordsArgs struct {
	Uid        string
	FromPeriod *time.Time
	ToPeriod   *time.Time
	RunIds     *[]string
}

// redundant code, GetUsageRecords covers function
func GetBucketUsageRecords(args GetBucketRecordsArgs) (*[]Record, error) {
	whereStatements := []string{"WHERE bucket_uid = $1"}
	sqlVars := []interface{}{}
	sqlVars = append(sqlVars, args.Uid)

	if args.RunIds != nil {
		whereStatements = append(whereStatements, "run_id = ANY($"+strconv.Itoa(len(whereStatements)+1)+")")
		sqlVars = append(sqlVars, *args.RunIds)
	}

	if args.FromPeriod != nil {
		whereStatements = append(whereStatements, "(period_end > "+"$"+strconv.Itoa(len(whereStatements)+1)+" OR period_end IS NULL)")
		sqlVars = append(sqlVars, args.FromPeriod)
	}

	if args.ToPeriod != nil {
		whereStatements = append(whereStatements, "period_start < "+"$"+strconv.Itoa(len(whereStatements)+1))
		sqlVars = append(sqlVars, args.ToPeriod)
	}

	sql := `
		SELECT id, bucket_uid, period_start, period_end, objects_count, bytes_total, run_id
		FROM records
		` + strings.Join(whereStatements, " AND ")

	rows, err := pool.Query(context.TODO(), sql, sqlVars...)
	defer rows.Close()

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	var records []Record
	for rows.Next() {
		var record Record
		err := rows.Scan(
			&record.ID,
			&record.BucketUid,
			&record.PeriodStart,
			&record.PeriodEnd,
			&record.ObjectsCount,
			&record.BytesTotal,
			&record.RunId,
		)

		if err != nil {
			fmt.Println(err)
			return nil, err
		}

		if args.FromPeriod != nil && record.PeriodStart.Before(*args.FromPeriod) {
			record.PeriodStart = *args.FromPeriod
		}

		if args.ToPeriod != nil && (record.PeriodEnd == nil || record.PeriodEnd.After(*args.ToPeriod)) {
			record.PeriodEnd = *&args.ToPeriod
		}

		records = append(records, record)
	}

	return &records, nil
}
