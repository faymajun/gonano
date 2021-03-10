package mongo

import (
	"encoding/hex"
	"fmt"
	"github.com/faymajun/gonano/core"
	"github.com/faymajun/gonano/core/coroutine"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/faymajun/gonano/config"
	"github.com/faymajun/gonano/core/routine"

	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/net/context"
)

const (
	Local_Mongo  = "LocalMongo"
	Battle_Mongo = "BattleMongo"
)

func InitDefaultConfig() {
	redisConf := MongoConfig{Addr: config.Content.String("mongo_addr"), Name: Local_Mongo}
	MongoMgr.Add(redisConf)
}

func LocalMongo() *Mongo {
	return MongoMgr.GetMongo(Local_Mongo)
}

func InitBattleConfig() {
	redisConf := MongoConfig{Addr: config.Content.String("battle_mongo_addr"), Name: Battle_Mongo}
	MongoMgr.Add(redisConf)
}

func BattleMongo() *Mongo {
	return MongoMgr.GetMongo(Battle_Mongo)
}

func InsertOne(c *mongo.Collection, document interface{}) {
	routine.Go(func() {
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		if _, error := c.InsertOne(ctx, document); error != nil {
			WriteOne(document, c.Database().Name(), c.Name(), error)
		}
	})
}

func InsertMany(c *mongo.Collection, documents []interface{}) {
	routine.Go(func() {
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		if _, error := c.InsertMany(ctx, documents); error != nil {
			WriteOne(documents, c.Database().Name(), c.Name(), error)
		}
	})
}

//有多线程问题
func Update(c *mongo.Collection, filter interface{}, document interface{}) {
	routine.Go(func() {
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		opt := options.Update().SetUpsert(true)
		if _, error := c.UpdateOne(ctx, filter, document, opt); error != nil {
			logger.Errorf("mongo update error, collection=%s, document=%v, error=%s", c.Name(), document, error)
		}
	})
}

//有多线程问题
func Del(c *mongo.Collection, filter interface{}) {
	routine.Go(func() {
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		if _, error := c.DeleteOne(ctx, filter); error != nil {
			logger.Errorf("mongo del error, collection=%s, error=%s", c.Name(), error)
		}
	})
}

const RoutineCount = 10 * 32

var routines []*coroutine.Coroutine

func InitClubConfig() {
	routines = make([]*coroutine.Coroutine, RoutineCount)
	for i := 0; i < RoutineCount; i++ {
		routines[i] = coroutine.NewCoroutine(64, int64(i))
	}
}

func UpdateSelectRoutine(key int64, c *mongo.Collection, filter interface{}, document interface{}) {
	index := key % RoutineCount
	_ = routines[index].PushTask(func() {
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		opt := options.Update().SetUpsert(true)
		if _, error := c.UpdateOne(ctx, filter, document, opt); error != nil {
			logger.Errorf("mongo update error, collection=%s, document=%v, error=%s", c.Name(), document, error)
		}
	}, false)
}

func DelSelectRoutine(key int64, c *mongo.Collection, filter interface{}) {
	index := key % RoutineCount
	_ = routines[index].PushTask(func() {
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		if _, error := c.DeleteOne(ctx, filter); error != nil {
			logger.Errorf("mongo del error, collection=%s, error=%s", c.Name(), error)
		}
	}, false)
}

func GetObjectId(data []byte) (primitive.ObjectID, error) {
	IDByte := [12]byte{}
	if len(data) != 24 {
		return primitive.ObjectID(IDByte), fmt.Errorf("GetObjectId data len is actual is %d, is not expected 24", len(data))
	}

	hexArr := make([]byte, 12)
	hex.Decode(hexArr, data)

	copy(IDByte[:], hexArr)
	return primitive.ObjectID(IDByte), nil
}

func GetUnixTime(data []byte) (time.Time, error) {
	if len(data) < 8 {
		return time.Unix(0, 0), fmt.Errorf("GetUnixTime data len is actual is %d, is not expected 8", len(data))
	}

	val, err := strconv.ParseInt(string(data[0:8]), 16, 32)
	if err != nil {
		return time.Unix(0, 0), fmt.Errorf("GetUnixTime ParseInt Error=%s", err)
	}

	return time.Unix(val, 0), nil
}

// 每天生成不同的collection
func GetCollectionOnDay(now time.Time) string {
	return core.Sprintf("date%d%d%d",
		now.Year(), now.Month(), now.Day())
}

// 获取collection name
func GetCollectionName(data []byte) (string, error) {
	time, err := GetUnixTime(data)
	if err != nil {
		return "", err
	}
	name := GetCollectionOnDay(time)
	return name, nil
}

//func BattleInsert(collection string, document interface{}, opts ...*options.InsertOneOptions) {
//	c := LocalMongo().Database("battle").Collection(collection)
//	insertOne(c, document)
//}
