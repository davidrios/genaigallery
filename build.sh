#!/bin/bash
set -e

cd frontend
npm run build
cd dist
7z a -tzip ../../frontend.zip .
cd ../../backend
go build -tags fts5 -ldflags "-s -w" -o ../backend-server cmd/server/main.go
cd ..
cat backend-server frontend.zip > genaigallery
chmod +x genaigallery
