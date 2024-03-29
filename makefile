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
	@GOPATH=$(GOPATH) $(GO) install -v geneSnapSession
	@GOPATH=$(GOPATH) $(GO) install -v wxpublic_svr
	@GOPATH=$(GOPATH) $(GO) install -v dizhidaxuePush
	@GOPATH=$(GOPATH) $(GO) install -v ddactivity_svr
	@GOPATH=$(GOPATH) $(GO) install -v ddsinger_svr
	@GOPATH=$(GOPATH) $(GO) install -v chengyuan_svr
	@GOPATH=$(GOPATH) $(GO) install -v om_svr
	mv bin/main bin/yplay_svr
	#cp bin/yplay_svr /home/work/yplay_svr/bin/yplay_svr

clean:
	@GOPATH=$(GOPATH) $(GO) clean -i main
	rm -rf bin/yplay_svr
