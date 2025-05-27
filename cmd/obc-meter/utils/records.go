package utils

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Record struct {
	ID           bson.ObjectID `bson:"_id,omitempty" json:"_id"`
	BucketUid    string        `bson:"bucket_uid" json:"bucket_uid"`
	PeriodStart  int64         `bson:"period_start" json:"period_start"`
	PeriodEnd    *int64        `bson:"period_end" json:"period_end"`
	ObjectsCount int32         `bson:"objects_count" json:"objects_count"`
	BytesTotal   int64         `bson:"bytes_count" json:"bytes_count"`
}

type createArgs struct {
	bucketId     string
	objectsCount int32
	bytesTotal   int64
}

func CreateRecord(args createArgs, ctx context.Context) (*Record, error) {
	coll := *getCollection()

	record := Record{
		BucketUid:    args.bucketId,
		ObjectsCount: args.objectsCount,
		BytesTotal:   args.bytesTotal,
		PeriodStart:  time.Now().Unix(),
	}

	res, err := coll.InsertOne(ctx, record)

	if err != nil {
		return nil, err
	}

	record.ID = res.InsertedID.(bson.ObjectID)

	return &record, nil
}

type deltaArgs struct {
	bucketId     string
	objectsCount int32
	bytesTotal   int64
}

func GetDeltaRecord(args deltaArgs, ctx context.Context) *Record {
	// coll := *getCollection()
	// col.find({bucketId, periodEnd: {$exists: false}, $or: [{objectsCount: {$ne: args.objectsCount}}, {bytesTotal: {$ne: args.bytesTotal}}]})

	return nil
}
