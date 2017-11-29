package trace

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"chuanyun.io/esmeralda/setting"
	"chuanyun.io/esmeralda/util"
	elastic "gopkg.in/olivere/elastic.v5"
)

type ErrorResult struct {
	Spans []ErrorSpans `json:"spans"`
	Meta  ErrorMeta
}

type ErrorMeta struct {
	Total int64 `json:"total"`
}

type ErrorSpans struct {
	ErrorMessage string `json:"errorMessage"`
	ErrorType    string `json:"errorType"`
	TraceId      string `json:"traceId"`
	Duration     int64  `json:"duration"`
	Time         int64  `json:"time"`
	Index        string `json:"index"`
}

type ListParams struct {
	Duration    int
	Limit       int
	ErrorType   string
	Value       string
	ServiceName string
	Ipv4        string
	From        int64
	To          int64
}

type WaterfallParams struct {
	Index   string
	TraceId string
}

type ErrorParams struct {
	Api  string
	From int64
	To   int64
}

func InitErrorResult() *ErrorResult {
	return &ErrorResult{
		Spans: []ErrorSpans{},
		Meta:  ErrorMeta{Total: 0},
	}
}

func (self *ErrorResult) DoingSpan(span Span) {
	errorSpans := ErrorSpans{
		Time:     span.Timestamp,
		Duration: span.Duration,
		Index:    util.FormatInt64Index(span.Timestamp),
		TraceId:  span.TraceId,
	}
	if len(span.BinaryAnnotations) > 0 {
		for _, bA := range span.BinaryAnnotations {
			if bA.Key == "" {
				continue
			}
			if bA.Key == "http.status_code" || bA.Key == "http.url" {
				errorSpans.ErrorType = "http"
				errorSpans.ErrorMessage = bA.Value
			}
			if bA.Key == "db.type" && bA.Value == "memcache" {
				errorSpans.ErrorType = "memcached"
			}
			if bA.Key == "db.type" && bA.Value == "mysql" {
				errorSpans.ErrorType = "mysql"
			}
			if bA.Key == "db.type" && bA.Value == "redis" {
				errorSpans.ErrorType = "redis"
			}
			if bA.Key == "error" {
				errorSpans.ErrorMessage = bA.Value
			}
		}
	}
	self.Spans = append(self.Spans, errorSpans)
}

func Lists(params *ListParams) *util.ResponseDebug {

	resp := &util.ResponseDebug{}
	resp.Status = 5001
	resp.Message = ""
	resp.Data = &struct{}{}

	if params.Limit > 1000 {
		resp.Message = "最多支持1000条查询结果"
		return resp
	}

	// 默认显示 10 条
	if params.Limit == 0 {
		params.Limit = 10
	}

	if params.Duration > 0 {
		params.Duration = params.Duration * 1000
	}

	// 默认设置当天0点时间为开始时间
	if params.From <= 0 {
		timeStr := time.Now().Format("2006-01-02")
		t, _ := time.ParseInLocation("2006-01-02", timeStr, time.Local)
		params.From = t.Unix()
	}

	var err error
	params.From, params.To, err = util.VerifyParamTime(resp, params.From, params.To)
	if err != nil {
		resp.Message = err.Error()
		return resp
	}

	resp.Status = http.StatusOK
	resp.Message = "OK"
	resp.Data, resp.Debug = GetTraceList(params)
	return resp
}

//trace 瀑布图
func Waterfall(params *WaterfallParams) *util.Response {
	resp := &util.Response{}
	resp.Status = http.StatusOK
	resp.Data = GetTraceWaterfall(params)
	return resp
}

func GetTraceList(params *ListParams) (map[string]*ListResult, interface{}) {
	traceIdList := []interface{}{}
	ListResultMap := map[string]*ListResult{}
	var dsl interface{}
	esClient := setting.Settings.Elasticsearch.Client

	//newTime := time.Now().Format("2006-01-02 15:04:05")
	_, _, fromTime, toTime := util.CalcTimeRange(params.From, params.To)
	esIndexes := getTraceTable(fromTime, toTime)

	//拼装es DSL
	query := elastic.NewBoolQuery()

	//条件 timestamp开始 和结束 >>start && << end
	query = query.Must(elastic.NewRangeQuery("timestamp").Gte(fromTime.UnixNano() / 1000).Lte(toTime.UnixNano() / 1000).
		IncludeLower(true).IncludeUpper(true))

	//条件 1.insertTime
	//query = query.Must(elastic.NewRangeQuery("insertTime").Lte(newTime))

	if len(params.Value) > 0 {
		//处理host
		queryShould := elastic.NewBoolQuery()
		//match_phrase
		queryShould = queryShould.Should(elastic.NewMatchPhraseQuery("binaryAnnotations.value", params.Value))
		queryShould = queryShould.Should(elastic.NewMatchPhraseQuery("relatedApi", params.Value))
		queryShould = queryShould.Should(elastic.NewMatchPhraseQuery("selfApi", params.Value))
		query = query.Must(queryShould)
	}

	//条件 duration 网络耗时
	if params.Duration > 0 {
		query = query.Must(elastic.NewRangeQuery("duration").Gte(params.Duration))
	}

	//条件 服务名
	if len(params.ServiceName) > 0 {
		query = query.Must(elastic.NewTermQuery("annotations.endpoint.serviceName", params.ServiceName))
	}

	if len(params.Ipv4) > 0 {
		query = query.Must(elastic.NewTermQuery("annotations.endpoint.ipv4", params.Ipv4))
	}

	//错误类型
	err, errTypes := parseErrorType(params.ErrorType)
	if err == nil {
		isAllErr := ""
		for _, errType := range errTypes {
			if errType == "all" {
				isAllErr = "all"
			}
		}
		if isAllErr == "all" {
			queryShould := elastic.NewBoolQuery()
			queryShould = queryShould.Should(createBoolMustTerm("binaryAnnotations.key", "error"))
			queryShould = queryShould.Should(createHttpStatusQuery())
			query = query.Must(queryShould)
		} else {
			queryShould := elastic.NewBoolQuery()
			for _, errType := range errTypes {
				if errType == "api" {
					queryShould = queryShould.Should(createHttpStatusQuery())
				} else {
					queryShould = queryShould.Should(createComponentQuery(errType))
				}
			}
			query = query.Must(queryShould)
		}
	}
	aggsTrace := elastic.NewTermsAggregation().Field("traceId").Size(params.Limit) //聚合

	dsl, _ = query.Source()

	tracesDSL := esClient.Search(esIndexes...).
		IgnoreUnavailable(true).
		FetchSource(false).
		Size(0).From(0).
		Sort("timestamp", false).
		Aggregation("traceId", aggsTrace).
		Query(query)

	if rlt, err := tracesDSL.Do(context.Background()); err != nil {
		fmt.Println("tracesDSL json err: ", err)
	} else {
		if terms, ok := rlt.Aggregations.Terms("traceId"); ok {
			for _, b := range terms.Buckets {
				traceIdList = append(traceIdList, b.Key.(string))
			}
		}
	}

	//获取多个traceId的数据
	traceQuery := elastic.NewBoolQuery().Must(elastic.NewTermsQuery("traceId", traceIdList...))
	tracelistDSL := esClient.Search(esIndexes...).
		IgnoreUnavailable(true).
		Size(1500).From(0).
		Query(traceQuery)

	if list, err := tracelistDSL.Do(context.Background()); err != nil {
		fmt.Println("tracelistDSL json err: ", err)
	} else {

		for _, hit := range list.Hits.Hits {
			s := Span{}
			if err := json.Unmarshal(*hit.Source, &s); err != nil {
				fmt.Println("tracelistDSL list json err: ", err)
			} else {
				if _, ok := ListResultMap[s.TraceId]; !ok {
					ListResultMap[s.TraceId] = InitResult(s.TraceId, s.Id)
				}
				if s.ParentId == "" {
					ListResultMap[s.TraceId].SetTimestamp(s.Timestamp)
					ListResultMap[s.TraceId].SetDuration(s.Duration)
					ListResultMap[s.TraceId].SetRoot(true)
				} else {
					if ListResultMap[s.TraceId].Root == false && s.Duration >= ListResultMap[s.TraceId].Duration {
						ListResultMap[s.TraceId].SetDuration(s.Duration)
					}
					if ListResultMap[s.TraceId].Timestamp == 0 {
						ListResultMap[s.TraceId].SetTimestamp(s.Timestamp)
					}
				}
				ListResultMap[s.TraceId].SpanPlus(s.Id) //span count++

				// @todo 什么情况下为空，以及如何处理
				if len(s.Annotations) == 0 || len(s.BinaryAnnotations) == 0 {
					fmt.Println("Annotations,BinaryAnnotations is empty")
					continue
				}

				//ServiceNameList
				serverName := s.Annotations[0].Endpoint.ServiceName
				if serverName != "" {
					ListResultMap[s.TraceId].SetServiceName(serverName, s.RelatedApi)
					ListResultMap[s.TraceId].ServiceNamePlus(serverName)
					ListResultMap[s.TraceId].ServiceNameDuration(serverName, s.Duration)
					ListResultMap[s.TraceId].ServiceNameUri(serverName, s.BinaryAnnotations)
				}
				ListResultMap[s.TraceId].setComponentInfo(s.BinaryAnnotations)
			}
		}
		//计算占比
		// for _, val := range ListResultMap {
		// 	val.TraceRatio()
		// }

	}

	return ListResultMap, dsl
}

func GetTraceWaterfall(params *WaterfallParams) *WaterResult {
	result := InitWaterResult()
	esIndexes := getWaterTable(params.Index)
	query := elastic.NewTermQuery("traceId", params.TraceId)
	esClient := setting.Settings.Elasticsearch.Client

	queryDSL := esClient.Search(esIndexes...).
		IgnoreUnavailable(true).
		FetchSource(true).
		Size(1500).From(0).
		Sort("timestamp", true).
		Query(query)
	if rlt, err := queryDSL.Do(context.Background()); err != nil {
		fmt.Println("GetTraceWaterfall search es error: ", err)
	} else {
		var span Span
		var spans []Span
		for _, val := range rlt.Hits.Hits {
			span = Span{}
			if err := json.Unmarshal(*val.Source, &span); err != nil {
				fmt.Println("GetTraceWaterfall Source json err: ", err)
			} else {
				result.SpanStat(span)
				spans = append(spans, span)
			}
		}
		result.SpanList(spans)
	}

	return result
}

func GetErrorDetail(params ErrorParams) *ErrorResult {
	result := InitErrorResult()
	newTime := time.Now().Format("2006-01-02 15:04:05")
	_, _, fromTime, toTime := util.CalcTimeRange(params.From, params.To)
	esIndexes := getTraceTable(fromTime, toTime)

	query := elastic.NewBoolQuery()
	//条件 timestamp开始 和结束 >>start && << end
	query = query.Must(elastic.NewRangeQuery("timestamp").Gte(fromTime.UnixNano() / 1000).Lte(toTime.UnixNano() / 1000).
		IncludeLower(true).IncludeUpper(true))
	query = query.Must(elastic.NewRangeQuery("insertTime").Lte(newTime))
	query = query.Must(elastic.NewTermQuery("relatedApi", params.Api))

	queryShould := elastic.NewBoolQuery()
	queryShould = queryShould.Should(createBoolMustTerm("binaryAnnotations.key", "error"))
	queryShould = queryShould.Should(createHttpStatusQuery())
	query = query.Must(queryShould)
	include := []string{"traceId", "binaryAnnotations", "timestamp"}
	fsc := elastic.NewFetchSourceContext(true).Include(include...)

	esClient := setting.Settings.Elasticsearch.Client
	errorDSL := esClient.Search(esIndexes...).
		IgnoreUnavailable(true).
		FetchSourceContext(fsc).
		Size(10).From(0).
		Sort("timestamp", false).
		Query(query)

	if rlt, err := errorDSL.Do(context.Background()); err != nil {
		fmt.Printf("GetTraceWaterfall search es error:%v", err)
	} else {
		var span Span
		result.Meta.Total = rlt.Hits.TotalHits
		for _, val := range rlt.Hits.Hits {
			span = Span{}
			if err := json.Unmarshal(*val.Source, &span); err != nil {
				fmt.Printf("Span json err: " + err.Error())
			} else {
				result.DoingSpan(span)
			}
		}
	}
	return result
}

func checkServerName(serverName string) bool {
	return serverName != "" && serverName != "mysql" && serverName != "redis" && serverName != "memcache"
}

//解析errorType
func parseErrorType(str string) (error, []string) {
	errTypes := []string{}
	if str == "" {
		return nil, errTypes
	}
	var err error
	if err = json.Unmarshal([]byte(str), &errTypes); err != nil {
		return err, errTypes
	}
	return nil, errTypes
}
