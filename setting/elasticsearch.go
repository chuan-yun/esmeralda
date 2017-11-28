package setting

import (
	elastic "gopkg.in/olivere/elastic.v5"
)

type ElasticsearchSettings struct {
	Hosts  []string
	Client *elastic.Client
}
