#!/bin/bash -e

root=$(dirname "$0")/..

# Delete db
rm -rf "${root}/integration-tests/.db"

# Run eventstore
trap 'pkill -P $$' EXIT
"${root}/build/eventstore" --db "badger://${root}/integration-tests/.db" --debug &

# Wait for eventstore to start listening
while ! nc -z localhost 8000; do
  sleep 0.1
done

# Run python tests
pytest "${root}/integration-tests/python"
