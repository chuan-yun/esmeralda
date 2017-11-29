package setting

import (
	"chuanyun.io/esmeralda/util"
	"github.com/sirupsen/logrus"
	elastic "gopkg.in/olivere/elastic.v5"
)

type ElasticsearchSettings struct {
	Hosts    []string
	Username string
	Password string
	Sniff    bool

	Client *elastic.Client
}

func InitializeElasticClient() {
	var elasticsearchOptions []elastic.ClientOptionFunc
	elasticsearchOptions = append(elasticsearchOptions, elastic.SetURL(Settings.Elasticsearch.Hosts...))
	if Settings.Elasticsearch.Username != "" && Settings.Elasticsearch.Password != "" {
		elasticsearchOptions = append(elasticsearchOptions, elastic.SetBasicAuth(Settings.Elasticsearch.Username, Settings.Elasticsearch.Password))
	}

	elasticsearchOptions = append(elasticsearchOptions, elastic.SetHealthcheck(true))
	elasticsearchOptions = append(elasticsearchOptions, elastic.SetSniff(Settings.Elasticsearch.Sniff))
	elasticsearchOptions = append(elasticsearchOptions, elastic.SetScheme("http"))

	var err error
	Settings.Elasticsearch.Client, err = elastic.NewClient(elasticsearchOptions...)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Fatal(util.Message("Initialize elasticsearch client connections failed"))
	}
}
