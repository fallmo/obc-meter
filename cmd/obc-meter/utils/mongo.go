package utils

import (
	"context"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var recordsCollection **mongo.Collection

func ConnectDatabase() {
	uri := os.Getenv("MONGO_URI")

	if uri == "" {
		log.Fatalln("Missing environment variable 'MONGO_URI'")
	}

	client, err := mongo.Connect(options.Client().ApplyURI(uri))

	if err != nil {
		panic(err)
	}

	defer func() {
		if err := client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()

	coll := client.Database("obc-meter").Collection("records")

	recordsCollection = &coll

	// var result bson.M

	// err = coll.FindOne(context.TODO(), bson.M{}).Decode(&result)

	// if err == mongo.ErrNoDocuments {
	// 	fmt.Println("No documents available")
	// 	return
	// }

	// if err != nil {
	// 	panic(err)
	// }

	// jsonData, err := json.MarshalIndent(result, "", "   ")
	// if err != nil {
	// 	panic(err)
	// }

	// fmt.Printf("%s\n", jsonData)

}

func getCollection() **mongo.Collection {
	if recordsCollection == nil {
		log.Panic("Missing database collection.")
	}

	return recordsCollection
}
