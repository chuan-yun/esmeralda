#!/bin/bash
#
# /etc/rc.d/init.d/esmeralda
#
# chuanyun.io esmeralda
#
#  chkconfig: 2345 20 80 Read
#  description: chuanyun.io esmeralda
#  processname: esmeralda

# Source function library.
. /etc/rc.d/init.d/functions

PROGNAME=esmeralda

PROGHOME=/home/licunchang/gopath/src/chuanyun.io/${PROGNAME}/target
PROG=${PROGHOME}/bin/${PROGNAME}
USER=licunchang

LOG_LEVEL=info
LOG_FILE=$PROGHOME/logs/${PROGNAME}-$(date +'%Y%m%d').log

LOCKFILE=/var/lock/subsys/$PROGNAME
pidfile=/var/run/$PROGNAME.pid

EXPORTER_PORT=10301

ELASTICSEARCH_HOSTS=http://10.209.26.199:11520,http://10.209.26.171:11520,http://10.209.26.172:11520,http://10.209.26.198:11520
KAFKA_GROUP_ID=licunchang
GATEWAY_URL=http://chuanyun.sit.ffan.biz/api/api/search
KAFKA_TOPICS=full_stack_tracing
ZOOKEEPER_ADDR=zk2181a.wdds.zk.com:2181,zk2181b.wdds.zk.com:2181,zk2181c.wdds.zk.com:2181
ZOOKEEPER_PATH=/ffan/kafka/ffan_service/svc_hippo
MODULE_ENABLE=false
MODULE_THRESHOLD=5
MYSQL_DSN="chuanyun:TianShang1Ge**@tcp(10.213.58.181:13306)/chuanyun?charset=utf8mb4"
BULK_SIZE=4000
BUFFER_SIZE=256

start() {
    echo -n $"Starting $PROGNAME: "

    daemon --pidfile=${pidfile} "$PROG -kafka.buffer=$BUFFER_SIZE -module.threshold=$MODULE_THRESHOLD -mysql.dsn='$MYSQL_DSN' -module.enable=$MODULE_ENABLE -log.level=$LOG_LEVEL -exporter.port=$EXPORTER_PORT -elasticsearch.hosts=$ELASTICSEARCH_HOSTS -kafka.group.id=$KAFKA_GROUP_ID -gateway.url=$GATEWAY_URL -kafka.topics=$KAFKA_TOPICS -zookeeper.addr=$ZOOKEEPER_ADDR -zookeeper.path=$ZOOKEEPER_PATH -elasticsearch.bulk.size=$BULK_SIZE >> $LOG_FILE 2>&1 &"
    RETVAL=$?
    echo
    [ $RETVAL = 0 ] && touch ${LOCKFILE}
    return $RETVAL
}

stop() {
    echo -n $"Stopping ${PROGNAME}: "
    killproc -p ${pidfile} ${PROGNAME}
    RETVAL=$?
    echo
    [ $RETVAL = 0 ] && rm -f ${LOCKFILE} ${pidfile}
}

case "$1" in
    start)
        start
        ;;
    stop)
        stop
        ;;
    restart)
        stop
        start
        ;;
    *)
        echo "Usage: service ${PROGNAME} {start|stop|restart}"
        exit 1
    ;;
esac