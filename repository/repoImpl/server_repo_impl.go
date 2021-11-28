package repoimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/esapi"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/go-redis/cache/v8"
	models "github.com/thteam47/server_management/model"
	repo "github.com/thteam47/server_management/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ServerRepositoryImpl struct {
	DB           *mongo.Database
	MyRediscache *cache.Cache
	Elas         *elasticsearch.Client
}

func NewServerRepo(mongodb *mongo.Database, redis *cache.Cache, elas *elasticsearch.Client) repo.ServerRepository {
	return &ServerRepositoryImpl{
		DB:           mongodb,
		MyRediscache: redis,
		Elas:         elas,
	}
}

const (
	detailStatusListKey = "detailstatusList_"
	statusServerKey     = "detailstatus_"
	searchKey           = "search_"
	totalSearchKey      = "search_total_"
	checkServerName     = "checkServername_"
	index               = "index_"
	totalIndex          = "index_total_"
)

func (s *ServerRepositoryImpl) Index(limitpage int64, numberpage int64) ([]*models.Server, int64, error) {
	var limit int64 = limitpage
	var page int64 = numberpage

	key := "index_" + strconv.FormatInt(limit, 10) + "_" + strconv.FormatInt(page, 10)
	keyTotal := "totalIndex_" + strconv.FormatInt(limit, 10) + "_" + strconv.FormatInt(page, 10)
	var dt []*models.Server
	var totalSer int64
	data := GetValueCache(s.MyRediscache, key, &dt)
	totalSe := GetValueCache(s.MyRediscache, keyTotal, &totalSer)

	fmt.Println("limit" + strconv.FormatInt(limit, 10))
	fmt.Println(page)
	fmt.Println("err")
	fmt.Println(data)
	fmt.Println(totalSe)
	fmt.Println("data")
	fmt.Println(data)
	fmt.Println(totalSer)
	if data == nil && totalSe == nil {
		return dt, totalSer, nil
	} else {
		fmt.Println("ok")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		findOptions := options.Find()
		findOptions.SetSort(bson.M{"created_at": -1})
		if page == 1 {
			findOptions.SetSkip(0)
			findOptions.SetLimit(limit)
		} else {
			findOptions.SetSkip((page - 1) * limit)
			findOptions.SetLimit(limit)
		}

		cur, err := s.DB.Collection(vi.GetString("collectionName")).Find(ctx, bson.M{}, findOptions)
		if err != nil {
			return nil, 0, err
		}
		curTotal, err := s.DB.Collection(vi.GetString("collectionName")).Find(ctx, bson.M{})
		if err != nil {
			return nil, 0, err
		}
		totalSer = int64(curTotal.RemainingBatchLength())
		var st []models.Server
		for cur.Next(context.TODO()) {
			// create a value into which the single document can be decoded
			var elem models.Server
			er := cur.Decode(&elem)
			if er != nil {
				return nil, 0, err
			}
			st = append(st, elem)
			fmt.Println(elem)
		}
		fmt.Println("data list")
		fmt.Println(st)
		if st == nil {
			SetValueCache(s.MyRediscache, key, dt)
			SetValueCache(s.MyRediscache, keyTotal, 0)
			SetKeyToListKeyCache(key, index)
			SetKeyToListKeyCache(keyTotal, totalIndex)
			return dt, 0, nil
		}

		for _, v := range st {
			var status string
			var r map[string]interface{}
			var buf bytes.Buffer
			query := map[string]interface{}{
				"query": map[string]interface{}{
					"match": map[string]interface{}{
						"_id": v.ID.Hex(),
					},
				},
			}
			fmt.Println(v)
			fmt.Println(query)
			if err := json.NewEncoder(&buf).Encode(query); err != nil {
				return nil, 0, err
			}
			res, err := s.Elas.Search(
				s.Elas.Search.WithContext(context.Background()),
				s.Elas.Search.WithIndex(vi.GetString("indexName")),
				s.Elas.Search.WithBody(&buf),
				s.Elas.Search.WithTrackTotalHits(true),
				s.Elas.Search.WithPretty(),
			)
			fmt.Println(err)
			if err != nil {
				return nil, 0, err
			}
			defer res.Body.Close()
			if !res.IsError() {
				if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
					log.Fatalf("Error parsing the response body: %s", err)
				}
				if r != nil {
					var detailSV models.ListStatus
					for _, hit := range r["hits"].(map[string]interface{})["hits"].([]interface{}) {
						m, _ := json.Marshal(hit.(map[string]interface{})["_source"])
						err := json.Unmarshal(m, &detailSV)
						if err != nil {
							log.Fatalf("Error getting response: %s", err)
						}
					}
					if len(detailSV.ChangeStatus) > 0 {
						status = detailSV.ChangeStatus[len(detailSV.ChangeStatus)-1].Status
					}
				}
			}
			dt = append(dt, &models.Server{
				ID:          v.ID,
				Username:    v.Username,
				ServerName:  v.ServerName,
				Ip:          v.Ip,
				Password:    v.Password,
				Description: v.Description,
				Status:      status,
			})
		}
		SetValueCache(s.MyRediscache, key, dt)
		SetValueCache(s.MyRediscache, keyTotal, totalSer)
		SetKeyToListKeyCache(key, index)
		SetKeyToListKeyCache(keyTotal, totalIndex)
	}
	return dt, totalSer, nil
}
func (s *ServerRepositoryImpl) AddServer(sv *models.Server) (*models.Server, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	infoSv := &models.Server{
		ID:          [12]byte{},
		Username:    sv.Username,
		Password:    sv.Password,
		ServerName:  sv.ServerName,
		Ip:          sv.Ip,
		Description: sv.Description,
		Port:        sv.Port,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	insertResult, err := s.DB.Collection(vi.GetString("collectionName")).InsertOne(ctx, infoSv)
	if err != nil {
		return nil, err
	}
	str := fmt.Sprintf("%v", insertResult.InsertedID)
	idResp := strings.Split(str, "\"")

	var listStatus []models.StatusDetail
	listStatus = append(listStatus, models.StatusDetail{
		Status: "Off",
		Time:   time.Now(),
	},
	)
	dataConvert := &models.ListStatus{
		ChangeStatus: listStatus,
	}
	dataJSON, err := json.Marshal(dataConvert)
	res := esapi.IndexRequest{
		Index:      vi.GetString("indexName"),
		DocumentID: idResp[1],
		Body:       strings.NewReader(string(dataJSON)),
	}
	res.Do(context.Background(), s.Elas)
	idRespPrimitive, _ := primitive.ObjectIDFromHex(idResp[1])
	infoSvReponse := &models.Server{
		ID:          idRespPrimitive,
		Username:    sv.Username,
		Password:    sv.Password,
		ServerName:  sv.ServerName,
		Ip:          sv.Ip,
		Description: sv.Description,
		Port:        sv.Port,
		Status:      sv.Status,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	RemoveData(s.MyRediscache)
	return infoSvReponse, nil
}
func (s *ServerRepositoryImpl) UpdateServer(idServer string, sv *models.Server) (*models.Server, error) {
	id, _ := primitive.ObjectIDFromHex(idServer)
	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{
		"username":    sv.Username,
		"ip":          sv.Ip,
		"servername":  sv.ServerName,
		"description": sv.Description,
		"port":        sv.Port,
	}}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := s.DB.Collection(vi.GetString("collectionName")).UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, err
	}
	idRespPrimitive, _ := primitive.ObjectIDFromHex(idServer)
	infoSvReponse := &models.Server{
		ID:          idRespPrimitive,
		Username:    sv.Username,
		ServerName:  sv.ServerName,
		Ip:          sv.Ip,
		Description: sv.Description,
		Port:        sv.Port,
	}
	RemoveListKeyCache(s.MyRediscache, searchKey)
	RemoveListKeyCache(s.MyRediscache, totalSearchKey)
	RemoveListKeyCache(s.MyRediscache, index)
	RemoveListKeyCache(s.MyRediscache, totalIndex)
	return infoSvReponse, nil
}
func (s *ServerRepositoryImpl) DetailsServer(idServer string, timeIn string, timeOut string) (string, []*models.StatusDetail, error) {
	id, _ := primitive.ObjectIDFromHex(idServer)
	keyList := "detailstatusList_" + idServer + "_" + timeIn + timeOut
	keyStatus := "detailstatus_" + idServer
	var statusList []*models.StatusDetail
	var statusServer string
	var detailSV models.ListStatus
	dataList := GetValueCache(s.MyRediscache, keyList, &statusList)
	dataStatus := GetValueCache(s.MyRediscache, keyStatus, &statusServer)

	if dataList != nil && dataStatus != nil {
		//search
		var r map[string]interface{}
		var buf bytes.Buffer
		query := map[string]interface{}{
			"query": map[string]interface{}{
				"match": map[string]interface{}{
					"_id": id,
				},
			},
		}
		if err := json.NewEncoder(&buf).Encode(query); err != nil {
			return "", nil, err
		}
		res, err := s.Elas.Search(
			s.Elas.Search.WithContext(context.Background()),
			s.Elas.Search.WithIndex(vi.GetString("indexName")),
			s.Elas.Search.WithBody(&buf),
			s.Elas.Search.WithTrackTotalHits(true),
			s.Elas.Search.WithPretty(),
		)
		if err != nil {
			log.Fatalf("Error getting response: %s", err)
		}
		defer res.Body.Close()
		if !res.IsError() {
			if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
				log.Fatalf("Error parsing the response body: %s", err)
			}
			for _, hit := range r["hits"].(map[string]interface{})["hits"].([]interface{}) {
				m, _ := json.Marshal(hit.(map[string]interface{})["_source"])
				err := json.Unmarshal(m, &detailSV)
				if err != nil {
					log.Fatalf("Error getting response: %s", err)
				}
			}
			if len(detailSV.ChangeStatus) == 0 {
				return "", nil, errors.New("Idserver not found")
			}
			var start time.Time
			var end time.Time
			if timeIn == "" {
				start = detailSV.ChangeStatus[0].Time
			} else {
				startTT, err := time.Parse(time.RFC3339, timeIn+"+07:00")
				if err != nil {
					log.Fatalf("error start")
				}
				if startTT.Before(detailSV.ChangeStatus[0].Time) == true {
					start = detailSV.ChangeStatus[0].Time
				} else {
					start = startTT
				}
			}
			if timeOut == "" {
				end = time.Now()
			} else {
				endTT, err := time.Parse(time.RFC3339, timeOut+"+07:00")
				if err != nil {
					log.Fatalf("error end")
				}
				if endTT.After(time.Now()) == true {
					end = time.Now()
				} else {
					end = endTT
				}
			}

			for i := 0; i < len(detailSV.ChangeStatus); i++ {
				tmp := detailSV.ChangeStatus[i].Time
				if tmp.Before(detailSV.ChangeStatus[len(detailSV.ChangeStatus)-1].Time) {
					if tmp.Before(start) && detailSV.ChangeStatus[i+1].Time.After(start) {
						statusList = append(statusList, &models.StatusDetail{
							Status: detailSV.ChangeStatus[i].Status,
							Time:   start,
						})
					}
				}
				if tmp.After(start) && tmp.Before(end) || tmp == start || tmp == end {
					statusList = append(statusList, &models.StatusDetail{
						Status: detailSV.ChangeStatus[i].Status,
						Time:   detailSV.ChangeStatus[i].Time,
					})
				}
				if tmp.Before(detailSV.ChangeStatus[len(detailSV.ChangeStatus)-1].Time) {
					if tmp.Before(end) && detailSV.ChangeStatus[i+1].Time.After(end) {
						statusList = append(statusList, &models.StatusDetail{
							Status: detailSV.ChangeStatus[i].Status,
							Time:   end,
						})
					}
				}
			}
			if end.After(detailSV.ChangeStatus[len(detailSV.ChangeStatus)-1].Time) {
				statusList = append(statusList, &models.StatusDetail{
					Status: detailSV.ChangeStatus[len(detailSV.ChangeStatus)-1].Status,
					Time:   end,
				})
			}
			//set value cache

			if err := SetValueCache(s.MyRediscache, keyList, statusList); err != nil {
				panic(err)
			}
			statusServer = detailSV.ChangeStatus[len(detailSV.ChangeStatus)-1].Status
			if err := SetValueCache(s.MyRediscache, keyStatus, statusServer); err != nil {
				panic(err)
			}
			//set list key cache

			SetKeyToListKeyCache(keyList, detailStatusListKey+idServer)
			SetKeyToListKeyCache(keyStatus, statusServerKey+idServer)

		}

	}
	return statusServer, statusList, nil
}
func (s *ServerRepositoryImpl) DeleteServer(idServer string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	id, _ := primitive.ObjectIDFromHex(idServer)
	var sv models.Server
	err := s.DB.Collection(vi.GetString("collectionName")).FindOne(ctx, bson.M{"_id": id}).Decode(&sv)
	if err != nil {
		return errors.New("Id incorrect")
	}
	result, err := s.DB.Collection(vi.GetString("collectionName")).DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		log.Fatal(err)
	}

	if result.DeletedCount == 0 {
		return errors.New("Id incorrect")
	}
	res := esapi.DeleteRequest{
		Index:      vi.GetString("indexName"),
		DocumentID: idServer,
	}
	_, err = res.Do(context.Background(), s.Elas)
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}

	RemoveListKeyCache(s.MyRediscache, detailStatusListKey+idServer)
	RemoveListKeyCache(s.MyRediscache, statusServerKey+idServer)
	RemoveData(s.MyRediscache)
	RemoveValueCache(s.MyRediscache, checkServerName+sv.ServerName)
	return nil
}
func (s *ServerRepositoryImpl) ChangePassword(idServer string, pass string) error {
	id, _ := primitive.ObjectIDFromHex(idServer)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var sv models.Server
	err := s.DB.Collection(vi.GetString("collectionName")).FindOne(ctx, bson.M{"_id": id}).Decode(&sv)
	if err != nil {
		return errors.New("Id incorrect")
	}
	var r map[string]interface{}
	var buf bytes.Buffer
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match": map[string]interface{}{
				"_id": idServer,
			},
		},
	}

	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		log.Fatalf("Error encoding query: %s", err)
	}
	res, err := s.Elas.Search(
		s.Elas.Search.WithContext(context.Background()),
		s.Elas.Search.WithIndex(vi.GetString("indexName")),
		s.Elas.Search.WithBody(&buf),
		s.Elas.Search.WithTrackTotalHits(true),
		s.Elas.Search.WithPretty(),
	)
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
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
		if detailSV.ChangeStatus[len(detailSV.ChangeStatus)-1].Status == "Invalid" {

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
						"_id": idServer,
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
			res, err := s.Elas.UpdateByQuery(
				[]string{vi.GetString("indexName")},
				s.Elas.UpdateByQuery.WithBody(&buf),
				s.Elas.UpdateByQuery.WithContext(context.Background()),
				s.Elas.UpdateByQuery.WithPretty(),
			)
			if err != nil {
				log.Fatalf("Error update: %s", err)
			}
			defer res.Body.Close()
		}
	}
	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{
		"password":   pass,
		"updated_at": time.Now(),
	}}
	_, er := s.DB.Collection(vi.GetString("collectionName")).UpdateOne(ctx, filter, update)
	if er != nil {
		return errors.New("Id incorrect")
	}
	RemoveData(s.MyRediscache)
	return nil
}

func (s *ServerRepositoryImpl) CheckServerName(servername string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var elem models.Server
	key := checkServerName + servername
	var check bool
	result := GetValueCache(s.MyRediscache, key, &check)
	if result != nil {
		collection := s.DB.Collection(vi.GetString("collectionName"))
		err := collection.FindOne(ctx, bson.M{"servername": servername}).Decode(&elem)
		if err != nil {
			check = false
		} else {
			check = true
		}
		SetValueCache(s.MyRediscache, key, check)
	}
	return check
}
func (s *ServerRepositoryImpl) Search(key string, field string, limitpage int64, numberpage int64) ([]*models.Server, int64, error) {
	var limit int64 = limitpage
	var page int64 = numberpage
	var dt []*models.Server
	var totalSer int64
	var status string
	keySearch := "search_" + key + "_" + strconv.FormatInt(limit, 10) + "_" + strconv.FormatInt(page, 10)
	keyTotal := "search_" + key + "_total" + "_" + strconv.FormatInt(limit, 10) + "_" + strconv.FormatInt(page, 10)
	result := GetValueCache(s.MyRediscache, keySearch, &dt)
	total := GetValueCache(s.MyRediscache, keyTotal, &totalSer)

	if result == nil && total == nil {
		return dt, totalSer, nil
	} else {
		filter := bson.M{
			field: bson.M{
				"$regex": primitive.Regex{
					Pattern: key,
					Options: "i",
				},
			},
		}

		collection := s.DB.Collection(vi.GetString("collectionName"))
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		findOptions := options.Find()
		if page == 1 {
			findOptions.SetSkip(0)
			findOptions.SetLimit(limit)
		} else {
			findOptions.SetSkip((page - 1) * limit)
			findOptions.SetLimit(limit)
		}
		findOptions.SetSort(bson.M{"created_at": -1})
		cur, err := collection.Find(ctx, filter, findOptions)
		if err != nil {
			log.Fatal(err)
		}
		curTotal, err := collection.Find(ctx, filter)
		if err != nil {
			log.Fatal(err)
		}
		totalSer = int64(curTotal.RemainingBatchLength())
		var st []models.Server
		for cur.Next(context.TODO()) {
			// create a value into which the single document can be decoded
			var elem models.Server
			er := cur.Decode(&elem)
			if er != nil {
				log.Fatal(err)
			}
			st = append(st, elem)
		}
		if st == nil {
			SetValueCache(s.MyRediscache, keySearch, dt)
			SetValueCache(s.MyRediscache, keyTotal, 0)
			SetKeyToListKeyCache(keySearch, searchKey)
			SetKeyToListKeyCache(keyTotal, totalSearchKey)
			return dt, 0, nil
		}
		for _, v := range st {

			var r map[string]interface{}
			var buf bytes.Buffer
			query := map[string]interface{}{
				"query": map[string]interface{}{
					"match": map[string]interface{}{
						"_id": v.ID.Hex(),
					},
				},
			}
			if err := json.NewEncoder(&buf).Encode(query); err != nil {
				log.Fatalf("Error encoding query: %s", err)
			}
			res, err := s.Elas.Search(
				s.Elas.Search.WithContext(context.Background()),
				s.Elas.Search.WithIndex(vi.GetString("indexName")),
				s.Elas.Search.WithBody(&buf),
				s.Elas.Search.WithTrackTotalHits(true),
				s.Elas.Search.WithPretty(),
			)
			if err != nil {
				log.Fatalf("Error getting response: %s", err)
			}
			defer res.Body.Close()
			if !res.IsError() {
				if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
					log.Fatalf("Error parsing the response body: %s", err)
				}
				if r != nil {
					var detailSV models.ListStatus
					for _, hit := range r["hits"].(map[string]interface{})["hits"].([]interface{}) {
						m, _ := json.Marshal(hit.(map[string]interface{})["_source"])
						err := json.Unmarshal(m, &detailSV)
						if err != nil {
							log.Fatalf("Error getting response: %s", err)
						}
					}
					if len(detailSV.ChangeStatus) > 0 {
						status = detailSV.ChangeStatus[len(detailSV.ChangeStatus)-1].Status
					}
				}
			}
			dt = append(dt, &models.Server{
				ID:          v.ID,
				Username:    v.Username,
				ServerName:  v.ServerName,
				Ip:          v.Ip,
				Port:        v.Port,
				Status:      status,
				Password:    v.Password,
				Description: v.Description,
			})
		}
		SetValueCache(s.MyRediscache, keySearch, dt)
		SetValueCache(s.MyRediscache, keyTotal, totalSer)
		SetKeyToListKeyCache(keySearch, searchKey)
		SetKeyToListKeyCache(keyTotal, totalSearchKey)
	}

	return dt, totalSer, nil
}
func RemoveData(MyRediscache *cache.Cache) {
	RemoveListKeyCache(MyRediscache, searchKey)
	RemoveListKeyCache(MyRediscache, totalSearchKey)
	RemoveListKeyCache(MyRediscache, index)
	RemoveListKeyCache(MyRediscache, totalIndex)
}
func (s *ServerRepositoryImpl) UpdateStatus() {
	for {
		collection := s.DB.Collection(vi.GetString("collectionName"))
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		cur, err := collection.Find(ctx, bson.M{})
		if err != nil {
			log.Fatal(err)
		}
		if cur != nil {
			for cur.Next(context.TODO()) {
				// create a value into which the single document can be decoded
				var elem models.Server
				er := cur.Decode(&elem)
				if er != nil {
					log.Fatal(err)
				}
				dayChange := time.Now().Sub(elem.UpdatedAt).Hours() / 24
				if dayChange > 60 {
					//get list status
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
					res, err := s.Elas.Search(
						s.Elas.Search.WithContext(context.Background()),
						s.Elas.Search.WithIndex(vi.GetString("indexName")),
						s.Elas.Search.WithBody(&buf),
						s.Elas.Search.WithTrackTotalHits(true),
						s.Elas.Search.WithPretty(),
					)
					if err != nil {
						log.Fatalf("Error getting response: %s", err)
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
						if detailSV.ChangeStatus[len(detailSV.ChangeStatus)-1].Status != "Invalid" {
							detailSV.ChangeStatus = append(detailSV.ChangeStatus, models.StatusDetail{
								Status: "Invalid",
								Time:   time.Now(),
							},
							)
							info := &models.ListStatus{
								ChangeStatus: detailSV.ChangeStatus,
							}
							var inInterface map[string]interface{}
							inrec, _ := json.Marshal(info)
							json.Unmarshal(inrec, &inInterface)
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
							res, err := s.Elas.UpdateByQuery(
								[]string{vi.GetString("indexName")},
								s.Elas.UpdateByQuery.WithBody(&buf),
								s.Elas.UpdateByQuery.WithContext(context.Background()),
								s.Elas.UpdateByQuery.WithPretty(),
							)
							if err != nil {
								log.Fatalf("Error update: %s", err)
							}
							defer res.Body.Close()
						}
					}
				}
			}
		}
		time.Sleep(1 * time.Hour)
	}
}
