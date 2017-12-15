package controller

import (
	"net/http"
	"strconv"
	"time"

	traceModel "chuanyun.io/esmeralda/model/trace"
	"chuanyun.io/esmeralda/util"
	"github.com/julienschmidt/httprouter"
)

func Lists(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	params := &traceModel.ListParams{}

	// default query params
	params.Limit = 10
	timeStr := time.Now().Format("2006-01-02")
	t, _ := time.ParseInLocation("2006-01-02", timeStr, time.Local)
	params.From = t.Unix()

	// response
	resp := &util.ResponseDebug{}
	resp.Status = 5001
	resp.Message = ""
	resp.Data = &struct{}{}

	m := util.GetParameters(r)

	if m["duration"] != "" {
		duration, err := strconv.Atoi(m["duration"])
		if err != nil {
			resp.Message = "耗时为整数"
			util.JSON(w, resp)
			return
		}
		params.Duration = duration * 1000
	}

	if m["limit"] != "" {
		limit, err := strconv.Atoi(m["limit"])
		if err != nil {
			resp.Message = "最大显示条数为整数"
			util.JSON(w, resp)
			return
		}

		if limit > 1000 {
			resp.Message = "最多支持1000条查询结果"
			util.JSON(w, resp)
			return
		}
		params.Limit = limit
	}

	if m["errorType"] != "" {
		limit, err := strconv.Atoi(m["limit"])
		if err != nil {
			resp.Message = "最大显示条数为整数"
			util.JSON(w, resp)
			return
		}
		params.Limit = limit
	}

	if m["serviceName"] != "" {
		params.ServiceName = m["serviceName"]
	}

	if m["ipv4"] != "" {
		params.Ipv4 = m["ipv4"]
	}

	if m["value"] != "" {
		params.Value = m["value"]
	}

	if m["from"] != "" {
		from, err := strconv.ParseInt(m["from"], 10, 64)
		if err != nil {
			resp.Message = "结束时间不正确"
			util.JSON(w, resp)
			return
		}
		params.From = from
	}

	if m["to"] != "" {
		to, err := strconv.ParseInt(m["to"], 10, 64)
		if err != nil {
			resp.Message = "结束时间不正确"
			util.JSON(w, resp)
			return
		}
		params.To = to
	}

	if (params.From > 0 && params.To > 0) && (params.From > params.To) {
		resp.Message = "开始时间不能大于结束时间"
		util.JSON(w, resp)
		return
	}

	var err error
	params.From, params.To, err = util.VerifyParamTime(resp, params.From, params.To)
	if err != nil {
		resp.Message = err.Error()
		util.JSON(w, resp)
		return
	}

	util.JSON(w, traceModel.Lists(params))
}

func Waterfall(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	resp := &util.Response{}
	resp.Status = 5001
	resp.Message = ""
	resp.Data = &struct{}{}

	params := &traceModel.WaterfallParams{}
	index := ps.ByName("index")
	traceID := ps.ByName("id")

	if index == "" || traceID == "" {
		resp.Message = "缺少 Index 或 TraceID 字段"
		util.JSON(w, resp)
		return
	}
	params.Index = index
	params.TraceID = traceID
	util.JSON(w, traceModel.Waterfall(params))
}
