package util

import (
	"encoding/json"
	"errors"
	"fmt"
	elastic "gopkg.in/olivere/elastic.v5"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Response struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type ResponseDebug struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Debug   interface{} `json:"debug"`
}

var (
	Status = map[int]string{0: "normal", 8: "warning", 9: "danger", 10: "miss"}
)

//获取状态
func GetStatus(level int) string {
	if v, ok := Status[level]; ok {
		return v
	}
	return Status[0]
}

// 输出JSON
func JSON(w http.ResponseWriter, Response interface{}) {
	rs, err := json.Marshal(Response)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprint(w, string(rs))
}

func CalcIdxsNew(prefix string, fromTime time.Time, toTime time.Time) []string {
	return []string{prefix}
}

func CalcIdxs(prefix string, fromTime time.Time, toTime time.Time) []string {
	idxs := []string{}
	t := time.Date(fromTime.Year(), fromTime.Month(), fromTime.Day(),
		0, 0, 0, 0, fromTime.Location())

	for toTime.After(t) {
		idx := t.Format("20060102")
		idxs = append(idxs, prefix+idx)
		t = t.AddDate(0, 0, 1)
	}
	return idxs
}

func GetAggsSumValI(rlt *elastic.SearchResult, term string) int {
	v, ok := rlt.Aggregations.Sum(term)
	if ok && v.Value != nil {
		return int(*v.Value)
	} else {
		return -1
	}
}

func GetAggsSumValF(rlt *elastic.SearchResult, term string) float64 {
	v, ok := rlt.Aggregations.Sum(term)
	if ok && v.Value != nil {
		return *v.Value
	} else {
		return -1.0
	}
}

// 校验时间参数的范围
func VerifyParamTime(resp *ResponseDebug, from, to int64) (int64, int64, error) {
	if from <= 0 {
		from = time.Now().Add(time.Second * -3600).Unix()
	}
	if to <= 0 {
		to = time.Now().Unix()
	}
	if from > to {
		resp.Message = "time error"
		return from, to, errors.New(resp.Message)
	}

	// 最大跨度查询3天
	if to > from+60*60*24*3 {
		resp.Message = "time not support"
		return from, to, errors.New(resp.Message)
	}
	return from, to, nil
}

// 根据 Timestamp 返回字符串和time.Time 形式的变量
func CalcTimeRange(from, to int64) (fromStr, toStr string, fromTime, toTime time.Time) {
	fromTime = time.Unix(from, 0)
	toTime = time.Unix(to, 0)
	fromStr = fromTime.Format("2006-01-02 15:04:05")
	toStr = toTime.Format("2006-01-02 15:04:05")
	return
}

func FormatInt64Index(t int64) string {
	return time.Unix(t/1000000, t*1000%1e9).Format("20060102")
}

func FormatInt64TimeNsec(t int64) string {
	nsec := t * 1000 % 1e9
	string := strconv.FormatInt(nsec, 10)
	return time.Unix(t/1000000, nsec).Format("2006-01-02 15:04:05") + "." + string
}

func MaxInt64(num int64, args ...int64) int64 {
	for _, v := range args {
		if num < v {
			num = v
		}
	}
	return num
}
