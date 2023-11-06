package main

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Database struct {
	*mongo.Client
}

func connectDatabase(mongoURI string) *Database {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("Error connecting to database: %s\n", err)
	}

	return &Database{client}
}

func (db *Database) GetLeaderboard(categories []Category, cursor *primitive.ObjectID, page int) (*Leaderboard, error) {

	projection := bson.M{
		"_id": 1,
	}

	for _, category := range categories {
		field := category.ToMongoField()
		projection[field] = bson.M{
			"$slice": bson.A{
				"$" + field,
				(page - 1) * 50,
				50,
			},
		}
	}

	opts := options.Aggregate().SetAllowDiskUse(true)

	pipeline := bson.A{
		bson.M{
			"$sort": bson.M{
				"_id": -1,
			},
		},
		bson.M{
			"$project": projection,
		},
	}

	if cursor != nil {
		pipeline = append(pipeline, bson.M{
			"$match": bson.M{
				"_id": cursor,
			},
		})
	} else {
		pipeline = append(pipeline, bson.M{
			"$limit": 1,
		})
	}

	timeNow := time.Now()

	cur, err := db.Database("battlebit").Collection("leaderboards").Aggregate(context.TODO(), pipeline, opts)
	if err != nil {
		return nil, err
	}

	log.Printf("Query took %s\n", time.Since(timeNow))

	defer cur.Close(context.Background())

	lb := &Leaderboard{}

	if cur.Next(context.Background()) {
		err = cur.Decode(lb)
		if err != nil {
			return nil, err
		}
	}

	return lb, nil
}

func (db *Database) AddLeaderboard(leaderboard *Leaderboard) {
	_, err := db.Database("battlebit").Collection("leaderboards").InsertOne(context.TODO(), leaderboard)
	if err != nil {
		log.Printf("Error inserting leaderboard: %s\n", err)
	}
}

func (db *Database) RemoveOldEntries() {
	cutoffTime := primitive.NewObjectIDFromTimestamp(time.Now().Add(-24 * time.Hour))

	filter := bson.M{"_id": bson.M{"$lt": cutoffTime}}

	_, err := db.Database("battlebit").Collection("leaderboards").DeleteMany(context.TODO(), filter)
	if err != nil {
		log.Printf("Error removing old entries: %s\n", err)
	}
}

func (db *Database) RemoveOldEntriesJob() {
	for {
		db.RemoveOldEntries()
		time.Sleep(15 * time.Minute)
	}
}
