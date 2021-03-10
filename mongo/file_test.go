package mongo

//func TestFile(t *testing.T) {
//	InitMongoFile("mongo")
//	defer CloseMongoLog()
//	record := &message.StoreRecord{
//		RoleId:     1,
//		Name:       "222",
//		RecordTime: timer.NowTime(),
//	}
//	redisConf := MongoConfig{Addr: "mongodb://192.168.1.100:27017", Name: Local_Mongo}
//	err := MongoMgr.Add(redisConf)
//	if err != nil {
//		t.Errorf("The error is not nil, but it should be, err=%v", err)
//	}
//
//	collection := LocalMongo().Database("battle").Collection("record")
//
//	InsertOne(collection, record)
//	InsertMany(collection, []interface{}{record, record})
//
//	time.Sleep(8 * time.Second)
//}
//
//func TestFile2(t *testing.T) {
//	InitMongoFile("mongo")
//	record := &message.StoreRecord{
//		RoleId:     1,
//		Name:       "222",
//		RecordTime: timer.NowTime(),
//	}
//	WriteOne(record, "db", "c", errors.New("error"))
//	CloseMongoLog()
//	time.Sleep(20 * time.Second)
//	WriteOne([]interface{}{record, record}, "db", "c", errors.New("error"))
//
//	time.Sleep(8 * time.Second)
//}
