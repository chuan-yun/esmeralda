package controller

import (
	traceModel "chuanyun.io/esmeralda/model/trace"
	"chuanyun.io/esmeralda/util"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"strconv"
)

func Lists(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	params := &traceModel.ListParams{}

	resp := &util.Response{}
	resp.Status = 5001
	resp.Message = ""
	resp.Data = &struct{}{}

	// 耗时条件
	if ps.ByName("duration") != "" {
		duration, err := strconv.Atoi(ps.ByName("duration"))
		if err != nil {
			resp.Message = "耗时为整数"
			util.JSON(w, resp)
			return
		}
		params.Duration = duration
	}

	// 查询条数
	if ps.ByName("limit") != "" {
		limit, err := strconv.Atoi(ps.ByName("limit"))
		if err != nil {
			resp.Message = "最大显示条数为整数"
			util.JSON(w, resp)
			return
		}
		params.Limit = limit
	}

	// erroType
	if ps.ByName("errorType") != "" {
		params.ErrorType = ps.ByName("errorType")
	}

	// serviceName
	if ps.ByName("serviceName") != "" {
		params.ServiceName = ps.ByName("serviceName")
	}

	// erroType
	if ps.ByName("ipv4") != "" {
		params.Ipv4 = ps.ByName("ipv4")
	}

	// api
	if ps.ByName("value") != "" {
		params.Value = ps.ByName("value")
	}

	// 耗时条件：开始时间
	if ps.ByName("from") != "" {
		from, err := strconv.ParseInt(ps.ByName("from"), 10, 64)
		if err != nil {
			resp.Message = "开始时间不正确"
			util.JSON(w, resp)
			return
		}
		params.From = from
	}

	// 耗时条件：结束时间
	if ps.ByName("to") != "" {
		to, err := strconv.ParseInt(ps.ByName("to"), 10, 64)
		if err != nil {
			resp.Message = "结束时间不正确"
			util.JSON(w, resp)
			return
		}
		params.To = to
	}

	util.JSON(w, traceModel.Lists(params))
}

// trace 瀑布图
func Waterfall(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	resp := &util.Response{}
	resp.Status = 5001
	resp.Message = ""
	resp.Data = &struct{}{}

	params := &traceModel.WaterfallParams{}
	index := ps.ByName("index")
	traceId := ps.ByName("id")

	if index == "" || traceId == "" {
		resp.Message = "缺少 Index 或 TraceId 字段"
		util.JSON(w, resp)
		return
	}
	params.Index = index
	params.TraceId = traceId
	util.JSON(w, traceModel.Waterfall(params))
}
