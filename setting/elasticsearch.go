package setting

import (
	"chuanyun.io/esmeralda/util"
	"github.com/sirupsen/logrus"
	elastic "gopkg.in/olivere/elastic.v5"
)

type ElasticsearchSettings struct {
	Hosts    []string
	Debug    bool
	Poolsize int
	Username string
	Password string

	Client *elastic.Client
}

func InitializeElasticClient() error {
	var elasticsearchOptions []elastic.ClientOptionFunc
	elasticsearchOptions = append(elasticsearchOptions, elastic.SetURL(Settings.Elasticsearch.Hosts...))
	if Settings.Elasticsearch.Username != "" && Settings.Elasticsearch.Password != "" {
		elasticsearchOptions = append(elasticsearchOptions, elastic.SetBasicAuth(Settings.Elasticsearch.Username, Settings.Elasticsearch.Password))
	}
	var err error
	Settings.Elasticsearch.Client, err = elastic.NewClient(elasticsearchOptions...)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Error(util.Message("elastic client init error"))
	}
	return err
}
