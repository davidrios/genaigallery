#!/bin/bash

echo Compiling backend server...

if test "$COPY_GO_LIB" == "1"; then
	# copy files to local to be able to use LSPs
	(while true; do
	watchmedo \
		auto-restart \
		-d /workdir/backend \
		--patterns="go.mod;go.sum" \
		--recursive \
		-D \
		--kill-after 1 \
		--interval 0.3 \
		--no-restart-on-command-exit \
		sh -- -c '
			mkdir -p ../gomods
			rsync -a /go/ ../gomods/go/
			rsync -a /usr/local/go/ ../gomods/usr-local-go/
			echo go lib copied
			sleep 1
		'
	sleep 1
	done) &
fi

while true; do
	watchmedo \
		auto-restart \
		-d /workdir/backend \
		--patterns="*.go" \
		--recursive \
		-D \
		--kill-after 1 \
		--interval 0.3 \
		sh -- -c 'go run --tags "fts5" cmd/server/main.go || pkill -f "/root/.cache/go-build" && sleep 1'
	sleep 1
done

