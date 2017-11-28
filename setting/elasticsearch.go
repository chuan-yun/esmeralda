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
	Sniff    bool

	Client *elastic.Client
}

func InitializeElasticClient() (err error) {
	var elasticsearchOptions []elastic.ClientOptionFunc
	elasticsearchOptions = append(elasticsearchOptions, elastic.SetURL(Settings.Elasticsearch.Hosts...))
	if Settings.Elasticsearch.Username != "" && Settings.Elasticsearch.Password != "" {
		elasticsearchOptions = append(elasticsearchOptions, elastic.SetBasicAuth(Settings.Elasticsearch.Username, Settings.Elasticsearch.Password))
	}

	elasticsearchOptions = append(elasticsearchOptions, elastic.SetHealthcheck(true))
	elasticsearchOptions = append(elasticsearchOptions, elastic.SetSniff(Settings.Elasticsearch.Sniff))
	elasticsearchOptions = append(elasticsearchOptions, elastic.SetScheme("http"))

	Settings.Elasticsearch.Client, err = elastic.NewClient(elasticsearchOptions...)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Error(util.Message("Initialize elasticsearch client connections failed"))
	} else {
		logrus.WithFields(logrus.Fields{
			"clients": Settings.Elasticsearch.Client,
		}).Info("Initialize elasticsearch client connections completed")
	}

	return err
}
