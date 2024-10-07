{
    "files_index": {
        "mappings": {
            "properties": {
                "fileName": {
                    "type": "keyword"
                },
                "lines": {
                    "type": "nested",
                    "properties": {
                        "line": {
                            "type": "text",
                            "analyzer": "ngram_analyzer"
                        },
                        "lineNumber": {
                            "type": "integer"
                        }
                    }
                }
            }
        }
    }
}


func main() {

    client, err := elasticsearch.NewDefaultClient()
    if err != nil {
        log.Fatal(err)
    }

    index := "products"
    mapping := `
    {
      "settings": {
        "number_of_shards": 1
      },
      "mappings": {
        "properties": {
          "field1": {
            "type": "text"
          }
        }
      }
    }`

    res, err := client.Indices.Create(
        index,
        client.Indices.Create.WithBody(strings.NewReader(mapping)),
    )
    if err != nil {
        log.Fatal(err)
    }

    log.Println(res)
}

{
    "files_index": {
        "mappings": {
            "properties": {
                "fileName": {
                    "type": "keyword"
                },
                "line": {
                    "type": "wildcard"
                },
                "lineNumber": {
                    "type": "integer"
                }
            }
        }
    }
}


package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/davecgh/go-spew/spew"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/krisch/crm-backend/internal/configs"
	"github.com/krisch/crm-backend/internal/emails"
	"github.com/krisch/crm-backend/internal/helpers"
	"github.com/krisch/crm-backend/internal/logs"

	"github.com/sirupsen/logrus"
)

var (
	version   = "0.0.0"
	buildTime = "0.0.0"
)

func main() {
	fmt.Println("------------------")

	logrus.Info("version: ", version)
	logrus.Info("build time: ", buildTime)
	logrus.Info("session id: ", helpers.RandomBigNumber())
	logrus.Info("now: ", helpers.DateNow())

	opt := configs.NewConfigsFromEnv()

	logs.InitLog(opt.LOG_FORMAT)

	opt.Debug()

	GetESClient()

	return

	MS, err := emails.NewFromCreds(opt.SMTP_CREDS)
	if err != nil {
		logrus.Fatal(err)
	}

	message, err := emails.NewConfirmationMessage("asdasdasdasd")
	if err != nil {
		logrus.Fatal(err)
	}

	err = MS.SendEmail([]string{"info@oviovi.site"}, message)

	if err != nil {
		logrus.Fatal(err)
	}
}

func GetESClient() {
	ctx := context.Background()

	es, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{"http://elasticsearch:9200"},
	})
	if err != nil {
		log.Fatalf("Error creating the client: %s", err)
	}

	res, err := es.Info()
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}

	fmt.Println(res)

	// for i := 1000; i < 10000; i++ {
	// 	Index(ctx, es, string(i), helpers.FakeSentence(100))
	// 	print(".")
	// }

	println("--------------------------------------------------")
	//

	res3, err := Search(ctx, es, "out lean muster")
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}

	for _, v := range res3 {
		fmt.Println(v.Id)
		fmt.Println("  " + v.Name)
	}

}

func Search(ctx context.Context, es *elasticsearch.Client, description string) ([]Student, error) {

	should := make([]interface{}, 0, 3)

	should = append(should, map[string]interface{}{
		"match": map[string]interface{}{
			"name": map[string]interface{}{
				"query":     description,
				"fuzziness": "AUTO",
			},
		},
	})

	var query map[string]interface{}

	if len(should) > 1 {
		query = map[string]interface{}{
			"query": map[string]interface{}{
				"bool": map[string]interface{}{
					"should": should,
				},
			},
		}
	} else {
		query = map[string]interface{}{
			"query": should[0],
		}
	}

	var buf bytes.Buffer

	_ = json.NewEncoder(&buf).Encode(query)
 
	req := esapi.SearchRequest{
		Index: []string{"students"},
		Body:  &buf,
	}

	resp, err := req.Do(ctx, es)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	println(resp.String())

	var hits struct {
		Hits struct {
			Hits []struct {
				Source Student `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	_ = json.NewDecoder(resp.Body).Decode(&hits) // XXX: error omitted

	res := make([]Student, len(hits.Hits.Hits))

	for i, hit := range hits.Hits.Hits {
		res[i].Id = hit.Source.Id
		res[i].Name = hit.Source.Name
	}

	return res, nil
}

func Index(ctx context.Context, es *elasticsearch.Client, id string, name string) error {
	// XXX: Excluding OpenTelemetry and error checking for simplicity

	// indexSettings := map[string]interface{}{
	// 	"settings": map[string]interface{}{
	// 		"number_of_shards":   1,
	// 		"number_of_replicas": 0,
	// 	},
	// 	"mappings": map[string]interface{}{
	// 		"properties": map[string]interface{}{
	// 			"title": map[string]interface{}{
	// 				"type": "text",
	// 			},
	// 			"author": map[string]interface{}{
	// 				"type": "keyword",
	// 			},
	// 		},
	// 	},
	// }

	body := Student{
		Id:   name,
		Name: name,
	}

	var buf bytes.Buffer

	err := json.NewEncoder(&buf).Encode(body)
	if err != nil {
		return err
	}

	req := esapi.IndexRequest{
		Index:      "students",
		Body:       &buf,
		DocumentID: body.Id,
		Refresh:    "true",
	}

	resp, err := req.Do(ctx, es)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

type SearchQuery struct {
	QueryString  string
	SearchFields []string
	SortField    string
	SortOrder    string
}

type Student struct {
	Id           string  `json:"id"`
	Name         string  `json:"name"`
	Age          int64   `json:"age"`
	AverageScore float64 `json:"average_score"`
}
