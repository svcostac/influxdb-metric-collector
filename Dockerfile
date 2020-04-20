FROM golang:latest

ENV BASE    /root/go

RUN mkdir -p $BASE/src    \
    && mkdir -p $BASE/pkg \
    && mkdir -p $BASE/bin 

ENV GOPATH      $BASE
ENV PATH        $PATH:$GOPATH/bin

RUN go get github.com/influxdata/influxdb1-client/v2

COPY ./src/collector/bin/default.conf /default.conf
COPY ./src/collector $BASE/src/collector

RUN cd $BASE/src/collector && go build -o collector .
RUN cp $BASE/src/collector/collector $BASE/bin/collector

CMD $BASE/bin/collector simulate -host $PODNAME -server $SERVER -sendertype $SENDERTYPE -conf $CONF_FILE -maxproc $MAXCLIENTS      
