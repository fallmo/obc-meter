package utils

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
}

func GetBucketCurrentRecord(bucketUid string) (*Record, error) {
	sql := `
		SELECT id, bucket_uid, period_start, period_end, objects_count, bytes_total
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

	err := pool.QueryRow(context.TODO(), sql, bucketUid).Scan(
		&ID,
		&BucketUid,
		&PeriodStart,
		&PeriodEnd,
		&ObjectsCount,
		&BytesTotal,
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
	}

	return &record, nil
}

type AppendBucketUsageRecordArgs struct {
	BucketUid    string
	ObjectsCount uint64
	BytesTotal   uint64
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
		`INSERT INTO records (bucket_uid, objects_count, bytes_total)
		VALUES ($1, $2, $3)
		RETURNING id, period_start`,
		args.BucketUid,
		args.ObjectsCount,
		args.BytesTotal,
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
	}

	return &record, nil
}

type GetBucketRecordsArgs struct {
	Uid         string
	From_period *time.Time
	To_period   *time.Time
}

func GetBucketUsageRecords(args GetBucketRecordsArgs) (*[]Record, error) {
	whereStatements := []string{"WHERE bucket_uid = $1"}
	sqlVars := []interface{}{}
	sqlVars = append(sqlVars, args.Uid)

	if args.From_period != nil {
		whereStatements = append(whereStatements, "(period_end > "+"$"+strconv.Itoa(len(whereStatements)+1)+" OR period_end IS NULL)")
		sqlVars = append(sqlVars, args.From_period)
	}

	if args.To_period != nil {
		whereStatements = append(whereStatements, "period_start < "+"$"+strconv.Itoa(len(whereStatements)+1))
		sqlVars = append(sqlVars, args.To_period)
	}

	sql := `
		SELECT id, bucket_uid, period_start, period_end, objects_count, bytes_total
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
		)

		if err != nil {
			fmt.Println(err)
			return nil, err
		}

		if args.From_period != nil && record.PeriodStart.Before(*args.From_period) {
			record.PeriodStart = *args.From_period
		}

		if args.To_period != nil && (record.PeriodEnd == nil || record.PeriodEnd.After(*args.To_period)) {
			record.PeriodEnd = *&args.To_period
		}

		records = append(records, record)
	}

	return &records, nil
}
