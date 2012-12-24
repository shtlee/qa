ROOT = ~/.qbox.me1
all:
	cd base; go install -v ./...

install: all
	mkdir -p $(ROOT)
	cd testing; go build -o bin/qboxtestcase
	rm $(ROOT)/conf.d
	ln -s $(QBOXROOT)/qa/testcase/testing/conf.d $(ROOT)/conf.d

clean:
	rm $(ROOT)/conf.d -rf
