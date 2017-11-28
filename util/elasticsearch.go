package util

import (
	elastic "gopkg.in/olivere/elastic.v5"
	"log"
	"os"
	"time"
)

type Pool struct {
	maxIdle int
	connCh  chan *elastic.Client
	addrs   []string
	debug   bool
}

func NewPool(addrs []string, maxIdle int, debug bool) *Pool {
	ep := Pool{}
	ep.maxIdle = maxIdle
	ep.addrs = addrs
	ep.debug = debug
	ep.connCh = make(chan *elastic.Client, maxIdle)
	return &ep
}

func (this *Pool) Close() {
	// 此时没有free放回到chan里的conn，会泄漏
	for {
		select {
		default:
			return
		case conn := <-this.connCh:
			if conn != nil {
				conn.Stop()
			}
		}
	}
}

//
func (this *Pool) Alloc() *elastic.Client {

	select {
	default:
		if conn, err := ConnEs(this.addrs, this.debug); err != nil {
			log.Printf("ConnEs error:%v %s", err, this.addrs)
			return nil
		} else {
			return conn
		}
	case conn := <-this.connCh:
		return conn
	}
}

func (this *Pool) Free(conn *elastic.Client) {
	select {
	default:
		conn.Stop()
	case this.connCh <- conn:
	}

}

func ConnEs(addrs []string, debug bool) (*elastic.Client, error) {
	opt := []elastic.ClientOptionFunc{
		elastic.SetURL(addrs...),
		elastic.SetHealthcheck(true),
		elastic.SetHealthcheckInterval(time.Second * 30),
		elastic.SetSniff(true),
		elastic.SetSnifferTimeout(time.Second * 5),
		elastic.SetSnifferInterval(time.Minute * 5),
		elastic.SetErrorLog(log.New(os.Stdout, "[ES_ERR]", log.LstdFlags)),
	}

	if debug {
		opt = append(opt, elastic.SetTraceLog(log.New(os.Stdout, "[ES_TRA]", log.LstdFlags)))
	}

	conn, err := elastic.NewClient(opt...)
	if err != nil {
		log.Printf("es new client error: %v ", err)
		return nil, err
	}

	return conn, nil
}
