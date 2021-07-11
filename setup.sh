#!/bin/sh

echo "setup called"
go run database/migrations/runner/main.go
echo "migration done"
go run database/seeds/runner/main.go
echo "seeding done"
./main