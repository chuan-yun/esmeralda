package collector

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/julienschmidt/httprouter"
)

var traceLog = `
[
    {
        "timestamp": 1511936207702000,
        "id": "9187827891033104209",
        "duration": 8000,
        "name": "get",
        "traceId": "7795430183876274723",
        "annotations": [
            {
                "timestamp": 1511936207702000,
                "value": "sr",
                "endpoint": {
                    "port": 80,
                    "ipv4": "10.213.1.64",
                    "serviceName": "coupon"
                }
            },
            {
                "timestamp": 1511936207710000,
                "value": "ss",
                "endpoint": {
                    "port": 80,
                    "ipv4": "10.213.1.64",
                    "serviceName": "coupon"
                }
            }
        ],
        "binaryAnnotations": [
            {
                "value": "http://coupon.intra.ffan.com/v1/members/14080714091903416/coupons?__v=v1&__trace_id=10.209.226.12-1511936207.7-76930-2848&_realip=10.213.64.103&orderNo=51132500413416",
                "endpoint": {
                    "port": 80,
                    "ipv4": "10.213.1.64",
                    "serviceName": "coupon"
                },
                "key": "http.url"
            },
            {
                "value": "get",
                "endpoint": {
                    "port": 80,
                    "ipv4": "10.213.1.64",
                    "serviceName": "coupon"
                },
                "key": "http.method"
            }
        ],
        "version": "java-1"
    },
    {
        "timestamp": 1511936207704000,
        "id": "8609849928096015910",
        "parentId": "9187827891033104209",
        "duration": 3000,
        "name": "execute",
        "traceId": "7795430183876274723",
        "annotations": [
            {
                "timestamp": 1511936207704000,
                "value": "cs",
                "endpoint": {
                    "port": 80,
                    "ipv4": "10.213.1.64",
                    "serviceName": "coupon"
                }
            },
            {
                "timestamp": 1511936207707000,
                "value": "cr",
                "endpoint": {
                    "port": 80,
                    "ipv4": "10.213.1.64",
                    "serviceName": "coupon"
                }
            }
        ],
        "binaryAnnotations": [
            {
                "value": "SELECT count(*) FROM ec12 c\n\t\t \n\t\t  \n\t\t \n\t\twhere 0#=1#\n\t\t  AND c.member_id = ?  \n\t\t \n\t\t \n\t\t \n\t\t \n\n\t\t \n\t\t \n\n\t\t \n\t\t \n\t\t \t\n\t\t \n\t\t \n\t\t \n\t\t \n\t\t \n\t\t \n\t\t \n\t\t  AND c.order_no=?",
                "endpoint": {
                    "port": 80,
                    "ipv4": "10.213.1.64",
                    "serviceName": "coupon"
                },
                "key": "db.statement"
            },
            {
                "value": "mysql",
                "endpoint": {
                    "port": 80,
                    "ipv4": "10.213.1.64",
                    "serviceName": "coupon"
                },
                "key": "db.type"
            },
            {
                "value": "coupon",
                "endpoint": {
                    "port": 80,
                    "ipv4": "10.213.1.64",
                    "serviceName": "coupon"
                },
                "key": "db.instance"
            },
            {
                "value": "true",
                "endpoint": {
                    "port": 3316,
                    "ipv4": "snew3316.wdds.mysqldb.com",
                    "serviceName": "coupon"
                },
                "key": "sa"
            }
        ],
        "version": "java-1"
    },
    {
        "timestamp": 1511936207708000,
        "id": "481671506603443567",
        "parentId": "9187827891033104209",
        "duration": 1000,
        "name": "execute",
        "traceId": "7795430183876274723",
        "annotations": [
            {
                "timestamp": 1511936207708000,
                "value": "cs",
                "endpoint": {
                    "port": 80,
                    "ipv4": "10.213.1.64",
                    "serviceName": "coupon"
                }
            },
            {
                "timestamp": 1511936207709000,
                "value": "cr",
                "endpoint": {
                    "port": 80,
                    "ipv4": "10.213.1.64",
                    "serviceName": "coupon"
                }
            }
        ],
        "binaryAnnotations": [
            {
                "value": "SELECT c.* FROM ec12 c\n\t\t \n\t\t  \n\t\t \n\t\twhere 0#=1#\n\t\t  AND c.member_id = ?  \n\t\t \n\t\t \n\t\t \n\t\t \n\n\t\t \n\t\t \n\n\t\t \n\t\t \n\t\t \t\n\t\t \n\t\t \n\t\t \n\t\t \n\t\t \n\t\t \n\t\t \n\t\t  AND c.order_no=? \n\n\t\t \n\n\t \n\t\t \n\t\t \n\t\t\torder by buy_time desc\n\t\t \n\t\tlimit ?, ?",
                "endpoint": {
                    "port": 80,
                    "ipv4": "10.213.1.64",
                    "serviceName": "coupon"
                },
                "key": "db.statement"
            },
            {
                "value": "mysql",
                "endpoint": {
                    "port": 80,
                    "ipv4": "10.213.1.64",
                    "serviceName": "coupon"
                },
                "key": "db.type"
            },
            {
                "value": "coupon",
                "endpoint": {
                    "port": 80,
                    "ipv4": "10.213.1.64",
                    "serviceName": "coupon"
                },
                "key": "db.instance"
            },
            {
                "value": "true",
                "endpoint": {
                    "port": 3316,
                    "ipv4": "s3new3316.wdds.mysqldb.com",
                    "serviceName": "coupon"
                },
                "key": "sa"
            }
        ],
        "version": "java-1"
    },
    {
        "timestamp": 1511936207709000,
        "id": "7488192887814003437",
        "parentId": "9187827891033104209",
        "duration": 1000,
        "name": "mget",
        "traceId": "7795430183876274723",
        "annotations": [
            {
                "timestamp": 1511936207709000,
                "value": "cs",
                "endpoint": {
                    "port": 80,
                    "ipv4": "10.213.1.64",
                    "serviceName": "coupon"
                }
            },
            {
                "timestamp": 1511936207710000,
                "value": "cr",
                "endpoint": {
                    "port": 80,
                    "ipv4": "10.213.1.64",
                    "serviceName": "coupon"
                }
            }
        ],
        "binaryAnnotations": [
            {
                "value": "[Ljava.lang.String;@5f0f1cdc,",
                "endpoint": {
                    "port": 80,
                    "ipv4": "10.213.1.64",
                    "serviceName": "coupon"
                },
                "key": "db.statement"
            },
            {
                "value": "redis",
                "endpoint": {
                    "port": 80,
                    "ipv4": "10.213.1.64",
                    "serviceName": "coupon"
                },
                "key": "db.type"
            },
            {
                "value": "true",
                "endpoint": {
                    "port": 10475,
                    "ipv4": "r10475.wdds.redis.com",
                    "serviceName": "redis"
                },
                "key": "sa"
            }
        ],
        "version": "java-1"
    }
]`

func TestHTTPCollector(t *testing.T) {
	path := "/collector/log"

	router := httprouter.New()
	router.POST(path, HTTPCollector)
	ts := httptest.NewServer(router)
	defer ts.Close()

	reader := bytes.NewReader([]byte(traceLog))
	t.Log(ts.URL)
	res, err := http.Post(ts.URL+path, "application/json", reader)
	if err != nil {
		log.Fatal(err)
	}
	greeting, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}

	t.Log(fmt.Sprintf("%s", greeting))
}
