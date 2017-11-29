package trace

import (
	"time"

	"chuanyun.io/esmeralda/util"
	elastic "gopkg.in/olivere/elastic.v5"
)

func getTraceTable(fromTime time.Time, toTime time.Time) []string {
	return util.CalcIdxs("chuanyun-", fromTime, toTime)
}

func getWaterTable(index string) []string {
	return []string{"chuanyun-" + index}
}

func createBoolMustTerm(key string, value string) *elastic.BoolQuery {
	query := elastic.NewBoolQuery()
	query = query.Must(elastic.NewTermQuery(key, value))
	return query
}

func createComponentQuery(ctype string) *elastic.BoolQuery {
	query := elastic.NewBoolQuery()
	query = query.Must(elastic.NewTermQuery("binaryAnnotations.key", "error"))
	query = query.Must(elastic.NewTermQuery("binaryAnnotations.key", "db.type"))
	query = query.Must(elastic.NewTermQuery("binaryAnnotations.value", ctype))
	return query
}

func createHTTPStatusQuery() *elastic.BoolQuery {
	query := elastic.NewBoolQuery()
	query = query.Must(elastic.NewTermQuery("binaryAnnotations.key", "http.status_code"))
	subQuery := elastic.NewBoolQuery()
	subQuery = subQuery.MustNot(elastic.NewWildcardQuery("binaryAnnotations.value", "1**"))
	subQuery = subQuery.MustNot(elastic.NewWildcardQuery("binaryAnnotations.value", "2**"))
	subQuery = subQuery.MustNot(elastic.NewWildcardQuery("binaryAnnotations.value", "3**"))
	query = query.Must(subQuery)
	return query
}
