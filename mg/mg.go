/**
* @Author: cl
* @Date: 2021/1/16 11:26
 */
package mg

import (
	"context"
	"github.com/ChenLong-dev/gobase/mlog"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	Client  *mongo.Client
	MongoDB *mongo.Database
	err     error
)

func Connect(host, username, password, dbname string) error {
	// Set client options IP:port
	clientOptions := options.Client().ApplyURI(host)

	//Set account and password
	credential := options.Credential{
		Username: username,
		Password: password,
	}
	clientOptions.SetAuth(credential)

	// Connect to MongoDB
	Client, err = mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		mlog.Error(err)
		return err
	}

	// Check the connection
	err = Client.Ping(context.TODO(), nil)
	if err != nil {
		mlog.Error(err)
		return err
	}

	mlog.Info("Connected to MongoDB!")
	MongoDB = Client.Database(dbname)

	return nil
}

func Disconnect() error {
	err = Client.Disconnect(context.TODO())
	if err != nil {
		mlog.Error(err)
		return err
	}

	mlog.Info("Connection to MongoDB closed.")
	return nil
}

func CreateCollection(name string) *mongo.Collection {
	return MongoDB.Collection(name)
}

func DropCollection(col *mongo.Collection) {
	col.Drop(context.TODO())
}

func InsertOne(col *mongo.Collection, doc interface{}) (*mongo.InsertOneResult, error) {
	return col.InsertOne(context.TODO(), doc)
}

func InsertMany(col *mongo.Collection, docs []interface{}) (*mongo.InsertManyResult, error) {
	return col.InsertMany(context.TODO(), docs)
}

func UpdateOne(col *mongo.Collection, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	return col.UpdateOne(context.TODO(), filter, update, opts ...)
}

func UpdateMany(col *mongo.Collection, filter interface{}, update interface{}) (*mongo.UpdateResult, error) {
	return col.UpdateMany(context.TODO(), filter, update)
}

func FindOne(col *mongo.Collection, v interface{}, filter interface{}) error {
	return col.FindOne(context.TODO(), filter).Decode(v)
}

func Find(col *mongo.Collection, filter interface{}, limit int64, skip int64) (*mongo.Cursor, error) {
	findOptions := options.Find()
	findOptions.SetLimit(limit)
	findOptions.SetSkip(skip)
	return col.Find(context.TODO(), filter, findOptions)
}

func DeleteOne(col *mongo.Collection, filter interface{}) (*mongo.DeleteResult, error) {
	return col.DeleteOne(context.TODO(), filter)
}

func DeleteMany(col *mongo.Collection, filter interface{}) (*mongo.DeleteResult, error) {
	return col.DeleteMany(context.TODO(), filter)
}
