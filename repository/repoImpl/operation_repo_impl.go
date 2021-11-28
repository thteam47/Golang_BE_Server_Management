package repoimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/smtp"
	"os"
	"strconv"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/go-redis/cache/v8"
	"github.com/tealeg/xlsx"
	models "github.com/thteam47/server_management/model"
	repo "github.com/thteam47/server_management/repository"
	"github.com/vigneshuvi/GoDateFormat"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type OperationRepositoryImpl struct {
	DB           *mongo.Database
	MyRediscache *cache.Cache
	Elas         *elasticsearch.Client
}

func NewOperationRepo(mongodb *mongo.Database, redis *cache.Cache, elas *elasticsearch.Client) repo.OperationRepository {
	return &OperationRepositoryImpl{
		DB:           mongodb,
		MyRediscache: redis,
		Elas:         elas,
	}
}
func (o *OperationRepositoryImpl) Connect(username string, password string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var elem models.Server
	collection := o.DB.Collection(vi.GetString("collectionName"))
	err := collection.FindOne(ctx, bson.M{"username": username, "password": password}).Decode(&elem)
	if err != nil {
		return errors.New("Username or password incorrect")
	}
	var r map[string]interface{}
	var buf bytes.Buffer
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match": map[string]interface{}{
				"_id": elem.ID.Hex(),
			},
		},
	}
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return err
	}
	res, err := o.Elas.Search(
		o.Elas.Search.WithContext(context.Background()),
		o.Elas.Search.WithIndex(vi.GetString("indexName")),
		o.Elas.Search.WithBody(&buf),
		o.Elas.Search.WithTrackTotalHits(true),
		o.Elas.Search.WithPretty(),
	)
	if err != nil {
		return err
	}
	if !res.IsError() {
		defer res.Body.Close()
		if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
			return err
		}

		var detailSV models.ListStatus
		for _, hit := range r["hits"].(map[string]interface{})["hits"].([]interface{}) {
			m, _ := json.Marshal(hit.(map[string]interface{})["_source"])
			err := json.Unmarshal(m, &detailSV)
			if err != nil {
				return err
			}
		}
		detailSV.ChangeStatus = append(detailSV.ChangeStatus, models.StatusDetail{
			Status: "On",
			Time:   time.Now(),
		},
		)
		info := &models.ListStatus{
			ChangeStatus: detailSV.ChangeStatus,
		}
		var inInterface map[string]interface{}
		inter, _ := json.Marshal(info)
		json.Unmarshal(inter, &inInterface)
		var buf bytes.Buffer
		doc := map[string]interface{}{
			"query": map[string]interface{}{
				"match": map[string]interface{}{
					"_id": elem.ID.Hex(),
				},
			},
			"script": map[string]interface{}{
				"source": "ctx._source.changeStatus=params.changeStatus;",
				"params": inInterface,
			},
		}
		if err := json.NewEncoder(&buf).Encode(doc); err != nil {
			return err
		}
		res, err := o.Elas.UpdateByQuery(
			[]string{vi.GetString("indexName")},
			o.Elas.UpdateByQuery.WithBody(&buf),
			o.Elas.UpdateByQuery.WithContext(context.Background()),
			o.Elas.UpdateByQuery.WithPretty(),
		)
		if err != nil {
			return err
		}
		defer res.Body.Close()
	}
	RemoveData(o.MyRediscache)
	return nil
}
func (o *OperationRepositoryImpl) Disconnect(idServer string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	id, _ := primitive.ObjectIDFromHex(idServer)
	collection := o.DB.Collection(vi.GetString("collectionName"))
	var elem models.Server
	err := collection.FindOne(ctx, bson.M{"_id": id}).Decode(&elem)
	if err != nil {
		return err
	}
	var r map[string]interface{}
	var buf bytes.Buffer
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match": map[string]interface{}{
				"_id": elem.ID.Hex(),
			},
		},
	}
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		log.Fatalf("Error encoding query: %s", err)
	}
	res, err := o.Elas.Search(
		o.Elas.Search.WithContext(context.Background()),
		o.Elas.Search.WithIndex(vi.GetString("indexName")),
		o.Elas.Search.WithBody(&buf),
		o.Elas.Search.WithTrackTotalHits(true),
		o.Elas.Search.WithPretty(),
	)
	if err != nil {
		return nil
	}
	if !res.IsError() {
		defer res.Body.Close()
		if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
			log.Fatalf("Error parsing the response body: %s", err)
		}

		var detailSV models.ListStatus
		for _, hit := range r["hits"].(map[string]interface{})["hits"].([]interface{}) {
			m, _ := json.Marshal(hit.(map[string]interface{})["_source"])
			err := json.Unmarshal(m, &detailSV)
			if err != nil {
				log.Fatalf("Error getting response: %s", err)
			}
		}
		detailSV.ChangeStatus = append(detailSV.ChangeStatus, models.StatusDetail{
			Status: "Off",
			Time:   time.Now(),
		},
		)
		info := &models.ListStatus{
			ChangeStatus: detailSV.ChangeStatus,
		}
		var inInterface map[string]interface{}
		inter, _ := json.Marshal(info)
		json.Unmarshal(inter, &inInterface)
		var buf bytes.Buffer
		doc := map[string]interface{}{
			"query": map[string]interface{}{
				"match": map[string]interface{}{
					"_id": elem.ID.Hex(),
				},
			},
			"script": map[string]interface{}{
				"source": "ctx._source.changeStatus=params.changeStatus;",
				"params": inInterface,
			},
		}
		if err := json.NewEncoder(&buf).Encode(doc); err != nil {
			log.Fatalf("Error update: %s", err)
		}
		res, err := o.Elas.UpdateByQuery(
			[]string{vi.GetString("indexName")},
			o.Elas.UpdateByQuery.WithBody(&buf),
			o.Elas.UpdateByQuery.WithContext(context.Background()),
			o.Elas.UpdateByQuery.WithPretty(),
		)
		if err != nil {
			log.Fatalf("Error update: %s", err)
		}
		defer res.Body.Close()
	}
	RemoveData(o.MyRediscache)
	return nil
}
func (o *OperationRepositoryImpl) Export(check bool, limitpage int64, numberpage int64) string {
	file := xlsx.NewFile()
	date := time.Now().Format(GoDateFormat.ConvertFormat("yyyy-MMM-dd-hh-MM-ss"))
	sheet, _ := file.AddSheet("ServerManagement")
	row := sheet.AddRow()
	colName := [7]string{"Server name", "Username", "Password", "Ip", "Port", "Description", "Status"}
	for i := 0; i < len(colName); i++ {
		cell := row.AddCell()
		cell.Value = colName[i]
	}
	collection := o.DB.Collection(vi.GetString("collectionName"))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var cur *mongo.Cursor
	var err error
	if check == false {
		cur, err = collection.Find(ctx, bson.M{})
	} else {
		page := numberpage
		limit := limitpage
		findOptions := options.Find()
		findOptions.SetSort(bson.M{"created_at": -1})
		if page == 1 {
			findOptions.SetSkip(0)
			findOptions.SetLimit(limit)
		} else {
			findOptions.SetSkip((page - 1) * limit)
			findOptions.SetLimit(limit)
		}
		cur, err = collection.Find(ctx, bson.M{}, findOptions)
	}
	if err != nil {
		log.Fatal(err)
	}
	for cur.Next(context.TODO()) {
		// create a value into which the single document can be decoded
		row = sheet.AddRow()
		var elem models.Server
		er := cur.Decode(&elem)
		if er != nil {
			log.Fatal(err)
		}
		listStatus := ""
		//search
		var r map[string]interface{}
		var buf bytes.Buffer
		query := map[string]interface{}{
			"query": map[string]interface{}{
				"match": map[string]interface{}{
					"_id": elem.ID.Hex(),
				},
			},
		}
		if err := json.NewEncoder(&buf).Encode(query); err != nil {
			log.Fatalf("Error encoding query: %s", err)
		}
		res, err := o.Elas.Search(
			o.Elas.Search.WithContext(context.Background()),
			o.Elas.Search.WithIndex(vi.GetString("indexName")),
			o.Elas.Search.WithBody(&buf),
			o.Elas.Search.WithTrackTotalHits(true),
			o.Elas.Search.WithPretty(),
		)
		if err != nil {
			log.Fatalf("Error getting response: %s", err)
		}
		defer res.Body.Close()
		if !res.IsError() {
			if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
				log.Fatalf("Error parsing the response body: %s", err)
			}
			var detailSV models.ListStatus
			for _, hit := range r["hits"].(map[string]interface{})["hits"].([]interface{}) {
				m, _ := json.Marshal(hit.(map[string]interface{})["_source"])
				err := json.Unmarshal(m, &detailSV)
				if err != nil {
					log.Fatalf("Error getting response: %s", err)
				}
			}

			listStatus = detailSV.ChangeStatus[len(detailSV.ChangeStatus)-1].Time.String() + ": " + detailSV.ChangeStatus[len(detailSV.ChangeStatus)-1].Status
			result := [7]string{elem.ServerName, elem.Username, elem.Password, elem.Ip, strconv.FormatInt(elem.Port, 10), elem.Description, listStatus}
			for i := 0; i < len(colName); i++ {
				cell := row.AddCell()
				cell.Value = result[i]
			}
		}
	}
	host, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	os.Setenv("DB_HOST", host)
	fileName := "swaggerui/export/" + date + ".xlsx"

	err = file.Save(fileName)
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}
	os.Setenv("FILENAME", fileName)

	return os.ExpandEnv("$DB_HOST:9090/$FILENAME")
}

func (o *OperationRepositoryImpl) SendMail() {
	from := vi.GetString("from")
	password := vi.GetString("password")
	smtpHost := vi.GetString("smtpHost")
	smtpPort := vi.GetString("smtpPort")
	to := []string{}
	auth := smtp.PlainAuth("", from, password, smtpHost)
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		curUser, err := o.DB.Collection(vi.GetString("collectionUser")).Find(ctx, bson.M{})
		if err != nil {
			log.Fatal(err)
		}
		if curUser != nil {
			for curUser.Next(context.TODO()) {
				var elem models.User
				er := curUser.Decode(&elem)
				if er != nil {
					log.Fatal(err)
				}
				if elem.Role == "admin" {
					to = append(to, elem.Email)
				}
			}
			collection := o.DB.Collection(vi.GetString("collectionName"))
			cur, err := collection.Find(ctx, bson.M{})
			result := ""
			if err != nil {
				log.Fatal(err)
			}
			if cur != nil {
				for cur.Next(context.TODO()) {
					var elem models.Server
					er := cur.Decode(&elem)
					if er != nil {
						log.Fatal(err)
					}
					var r map[string]interface{}
					var buf bytes.Buffer
					query := map[string]interface{}{
						"query": map[string]interface{}{
							"match": map[string]interface{}{
								"_id": elem.ID.Hex(),
							},
						},
					}
					if err := json.NewEncoder(&buf).Encode(query); err != nil {
						log.Fatalf("Error encoding query: %s", err)
					}
					res, err := o.Elas.Search(
						o.Elas.Search.WithContext(context.Background()),
						o.Elas.Search.WithIndex("server-elastest"),
						o.Elas.Search.WithBody(&buf),
						o.Elas.Search.WithTrackTotalHits(true),
						o.Elas.Search.WithPretty(),
					)
					if err != nil {
						log.Fatalf("Error getting response: %s", err)
					}
					defer res.Body.Close()
					if !res.IsError() {
						if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
							log.Fatalf("Error parsing the response body: %s", err)
						}
						var detailSV models.ListStatus
						for _, hit := range r["hits"].(map[string]interface{})["hits"].([]interface{}) {
							m, _ := json.Marshal(hit.(map[string]interface{})["_source"])
							err := json.Unmarshal(m, &detailSV)
							if err != nil {
								log.Fatalf("Error getting response: %s", err)
							}
						}
						if len(detailSV.ChangeStatus) > 0 {
							result += "Id: " + elem.ID.Hex() + ", Server name: " + elem.ServerName + ", Status: " + detailSV.ChangeStatus[len(detailSV.ChangeStatus)-1].Status + "\n"
						}
					}
				}
			}
			msg := []byte("To: THteaM" + "\r\n" +
				"Subject: Daily monitoring report of server status\r\n" +
				"\r\n" +
				result + "\r\n")

			err = smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, msg)
			if err != nil {
				log.Fatalf("Error getting response: %s", err)
			}
		}

		time.Sleep(24 * time.Hour)
	}
}
