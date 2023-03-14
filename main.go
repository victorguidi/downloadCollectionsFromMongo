package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var configEnv string

type Mongodb struct {
	Databases  []string
	Client     *mongo.Client
	ctx        context.Context
	connString string
}

func (m *Mongodb) Connect() error {
	m.ctx = context.TODO()
	client, err := mongo.Connect(m.ctx, options.Client().ApplyURI(m.connString))
	if err != nil {
		return err
	}
	m.Client = client
	return nil
}

func (m *Mongodb) GetDatabases() {
	databases, err := m.Client.ListDatabaseNames(m.ctx, bson.M{})
	if err != nil {
		panic(err)
	}
	m.Databases = databases
}

func (m *Mongodb) RunFindInAllCollections() {
loop:
	for _, database := range m.Databases {
		if configEnv == "dev" {
			if database == "network-energy" || database == "QBR" || database == "qbr" || database == "ScalaDrillTest" || database == "notasFiscais_dev" || database == "test" || database == "test_powerBI" || database == "ScalaCosts" {
				continue loop
			}
		}
		if configEnv == "prod" {
			if database == "test" || database == "geographic" {
				continue loop
			}
		}
		collections, err := m.Client.Database(database).ListCollectionNames(m.ctx, bson.M{})
		if err != nil {
			panic(err)
		}
		for _, collection := range collections {

			if collection == "clients" || collection == "data-center-power-consumption" || collection == "manual-data-snapshot" || collection == "data-centers-to-crawl" || collection == "crawler-history" || collection == "EnergyData" || configEnv == "prod" && collection == "users" {
				continue
			}
			cursor, err := m.Client.Database(database).Collection(collection).Find(m.ctx, bson.M{})
			if err != nil {
				panic(err)
			}
			defer cursor.Close(m.ctx)

			var results []bson.M
			if err = cursor.All(m.ctx, &results); err != nil {
				panic(err)
			}

			now := time.Now()
			filename := fmt.Sprintf("%s_%s_%s.json", database, collection, now.Format("20060102150405"))
			file, err := os.Create("/home/victorguidi/projects/golang/backupFiles/backups/" + filename)
			if err != nil {
				panic(err)
			}
			defer file.Close()

			encoder := json.NewEncoder(file)
			encoder.SetIndent("", "  ")
			if err := encoder.Encode(results); err != nil {
				panic(err)
			}
		}
	}
}

func main() {
	flag.StringVar(&configEnv, "c", "", "config file path")
	flag.Parse()

	if configEnv == "" {
		panic("config file path is required")
	}

	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	if configEnv == "dev" {
		mongodb := Mongodb{
			connString: os.Getenv("MONGODB_DEV"),
		}
		err := mongodb.Connect()
		if err != nil {
			panic(err)
		}
		mongodb.GetDatabases()
		mongodb.RunFindInAllCollections()
	} else if configEnv == "prod" {
		mongodb := Mongodb{
			connString: os.Getenv("MONGODB_PROD"),
		}
		err := mongodb.Connect()
		if err != nil {
			panic(err)
		}
		mongodb.GetDatabases()
		mongodb.RunFindInAllCollections()
	}
}
