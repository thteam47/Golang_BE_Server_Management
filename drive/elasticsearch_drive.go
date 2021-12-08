package drive

import (
	"log"

	"github.com/elastic/go-elasticsearch/v7"
)

type ElasticsearchDB struct {
	Elas *elasticsearch.Client
}

var Elasticsearch = &ElasticsearchDB{}

func ConnectElasticsearch(dbelasticurl string) *ElasticsearchDB {
	cfg := elasticsearch.Config{
		Addresses: []string{
			dbelasticurl,
		},
	}
	clientEL, err := elasticsearch.NewClient(cfg)
	if err != nil {
		log.Fatalf("Error creating the client: %s", err)
	}
	Elasticsearch.Elas = clientEL
	return Elasticsearch
}
