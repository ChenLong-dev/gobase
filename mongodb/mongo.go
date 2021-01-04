/*
 * @Description:
 * @Author: Chen Long
 * @Date: 2020-12-17 09:19:26
 * @LastEditTime: 2020-12-17 09:19:26
 * @LastEditors: Chen Long
 * @Reference:
 */

package mongodb

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ChenLong-dev/gobase/config"
	"github.com/ChenLong-dev/gobase/etcd"
	"github.com/ChenLong-dev/gobase/mlog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
	"time"
)

var GlobalMongoClient *mongo.Client
var GlobalDatabase *mongo.Database

func init() {
	//判断config中mongo字段中key是否为空
	if config.Conf.Mongo.Key != "" {
		etcd.Connect()
		defer etcd.Close()
		//从etcd获取字段
		mongoConf, err := etcd.Get(config.Conf.Mongo.Key)
		if err != nil {
			panic(fmt.Sprintf("[MONGODB] Get etcd_config err: %s \n", err))
		}
		if err = json.Unmarshal(mongoConf, &config.Conf.Mongo); err != nil {
			panic(fmt.Sprintf("[MONGODB] Parse etcd_config err: %s \n", err))
		}
	}
	mlog.Debugf("[MONGODB] GET MONGO CONF HOST: %s\n", config.Conf.Mongo.Host)
	if _, err := newClientFromConf(config.Conf.Mongo); err != nil {
		panic(fmt.Sprintf("[MONGODB] NEW CONNECT FAIL: %s \n", err))
	}
}

func newClientFromConf(mongoConf *config.Mongo) (*mongo.Client, error) {

	if mongoConf.MaxConns == 0 {
		mongoConf.MaxConns = 100
	}
	Readconcern := false
	Writeconcern := false

	opts := &options.ClientOptions{}

	opts.SetMaxPoolSize(mongoConf.MaxConns)
	opts.ApplyURI(mongoConf.Host)

	if Readconcern == true {
		opts.SetReadConcern(readconcern.Majority())
	}

	if Writeconcern == true {
		wc := writeconcern.New(writeconcern.WMajority())
		opts.SetWriteConcern(wc)
	}
	if mongoConf.UserName != "" {
		credential := options.Credential{
			Username:   mongoConf.UserName,
			Password:   mongoConf.Password,
			AuthSource: mongoConf.AuthDB,
		}
		opts.SetAuth(credential)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Second*10))
	defer cancel()
	mongoCli, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, err
	}

	mlog.Info(opts)
	GlobalMongoClient = mongoCli
	GlobalDatabase = GlobalMongoClient.Database(mongoConf.DbName)

	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = mongoCli.Ping(ctx, nil)
	if err != nil {
		mlog.Fatal(err)
	}

	return GlobalMongoClient, nil
}

func Disconnect() error {
	err := GlobalMongoClient.Disconnect(context.TODO())
	if err != nil {
		mlog.Errorf("[MONGODB] DISCONNECT ERROR: %v \n", err)
		return err
	}

	mlog.Debug("[MONGODB]Connection to MongoDB closed.")
	return nil
}

func CreateCollection(name string) *mongo.Collection {
	mlog.Debugf("[MONGODB]Create Collection: %v \n", name)
	return GlobalDatabase.Collection(name)
}

func DropCollection(col *mongo.Collection) {
	col.Drop(context.TODO())
	mlog.Debugf("[MONGODB]Drop Collection: %v \n", col)
}

func InsertOne(col *mongo.Collection, doc interface{}) (*mongo.InsertOneResult, error) {
	mlog.Debugf("[MONGODB]Insert One Document: %v \n", doc)
	return col.InsertOne(context.TODO(), doc)
}

func InsertMany(col *mongo.Collection, docs []interface{}) (*mongo.InsertManyResult, error) {
	mlog.Debugf("[MONGODB]Insert Many Document: %v \n", docs)
	return col.InsertMany(context.TODO(), docs)
}

func UpdateOne(col *mongo.Collection, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	mlog.Debugf("[MONGODB]Update One Document filter: %v update: %v \n", filter, update)
	return col.UpdateOne(context.TODO(), filter, update, opts...)
}

func UpdateMany(col *mongo.Collection, filter interface{}, update interface{}) (*mongo.UpdateResult, error) {
	mlog.Debugf("[MONGODB]Update UpdateMany Document filter: %v update: %v \n", filter, update)
	return col.UpdateMany(context.TODO(), filter, update)
}

func FindOne(col *mongo.Collection, v interface{}, filter interface{}) error {
	mlog.Debugf("[MONGODB]FindOne Document filter: %v \n", filter)
	err := col.FindOne(context.TODO(), filter).Decode(v)
	if err == mongo.ErrNilDocument || err == mongo.ErrNoDocuments {
		mlog.Infof("[MONGODB]FindOne Document Data Not Found: %v \n", v)
		return nil
	}
	return err
}

func Find(col *mongo.Collection, filter interface{}, limit int64) (*mongo.Cursor, error) {
	mlog.Debugf("[MONGODB]Find Document filter: %v \n", filter)
	findOptions := options.Find()
	findOptions.SetLimit(limit)
	return col.Find(context.TODO(), filter, findOptions)
}

func FindMany(col *mongo.Collection, filter interface{}, findOptions *options.FindOptions) (results []bson.M, err error) {
	mlog.Debugf("[MONGODB]Find Many Document filters: %v \n", filter)
	//findOptions.SetLimit(f.Limit)
	//findOptions.SetSkip(f.Skip)
	//findOptions.SetProjection(projection)
	//findOptions.SetSort(sort)
	cursor, err := col.Find(context.TODO(), filter, findOptions)
	if err != nil {
		if err == mongo.ErrNilDocument {
			mlog.Errorf("[MONGO] Find Many: %s", err)
			return nil, nil
		}
		return nil, err
	}
	defer cursor.Close(context.TODO())
	for cursor.Next(context.TODO()) {
		//解码文档
		var elem bson.M
		err = cursor.Decode(&elem)
		if err != nil {
			mlog.Errorf("data: %s, error: %s", cursor.Current, err)
			continue
		}
		results = append(results, elem)
	}

	return results, err
}

func DeleteOne(col *mongo.Collection, filter interface{}) (*mongo.DeleteResult, error) {
	mlog.Debugf("[MONGODB]Delete One Document filter: %v \n", filter)
	return col.DeleteOne(context.TODO(), filter)
}

func DeleteMany(col *mongo.Collection, filter interface{}) (*mongo.DeleteResult, error) {
	mlog.Debugf("[MONGODB]Delete Many Document filter: %v \n", filter)
	return col.DeleteMany(context.TODO(), filter)
}
