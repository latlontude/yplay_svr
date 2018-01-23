GOPATH=$(shell pwd)
DATE=$(shell date)
GO=go

all: yplay_svr

yplay_svr:
	@go version
	#@GOPATH=$(GOPATH) $(GO) build -o ypaly_svr -v main
	@GOPATH=$(GOPATH) $(GO) install -v main
	@GOPATH=$(GOPATH) $(GO) install -v calc2Degree
	@GOPATH=$(GOPATH) $(GO) install -v frozenMonitor
	@GOPATH=$(GOPATH) $(GO) install -v mysqldeadlock
	@GOPATH=$(GOPATH) $(GO) install -v phoneParser
	@GOPATH=$(GOPATH) $(GO) install -v preGeneQIds
	mv bin/main bin/yplay_svr
	#cp bin/yplay_svr /home/work/yplay_svr/bin/yplay_svr

clean:
	@GOPATH=$(GOPATH) $(GO) clean -i main
	rm -rf bin/yplay_svr
