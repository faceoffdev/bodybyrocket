PORT=2346
DLV=dlv --listen=:$(PORT) --headless=true --api-version=2 --accept-multiclient exec

build:
	go build -buildvcs=false -gcflags "all=-N -l" -o /app/bbr

upload: build
	/app/bbr upload

parse: build
	/app/bbr parse

dlv-upload: build
	$(DLV) /app/bbr upload

dlv-parse: build
	$(DLV) /app/bbr parse
