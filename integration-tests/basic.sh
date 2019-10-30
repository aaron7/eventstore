#!/bin/bash -e

root=$(dirname "$0")/..

# Delete db
rm -rf "${root}/integration-tests/.db"

# Run eventstore
trap 'pkill -P $$' EXIT
"${root}/build/eventstore" --db "badger://${root}/integration-tests/.db" &

# Wait for eventstore to start listening
while ! nc -z localhost 8000; do
  sleep 0.1
done

# Post some events
response=$(curl -s -X POST \
  http://localhost:8000/events \
  -w '%{http_code}\n' \
  -H 'Content-Type: application/json' \
  -d '{
  "events": [{
    "ts": 1001,
    "samplerate": 1,
    "data": {
      "dim1": "foo",
      "dim2": "bar2"
    }
  }, {
    "ts": 1002,
    "samplerate": 1,
    "data": {
      "dim1": "foo",
      "dim2": "bar2",
      "dim3": "oof"
    }
  }]
}')
if [[ $response != "200" ]]; then
  echo "Error posting events."
  exit 1
fi

# Query events

no_matches=$(curl -s -X POST \
  http://localhost:8000/query \
  -H 'Content-Type: application/json' \
  -d '{
  "data": [
    {
      "name": "test",
      "keys": ["dim1", "dim2"],
      "filters":[
        {
          "type": "eq",
          "key": "dim1",
          "value": "foo"
        },{
          "type": "eq",
          "key": "dim2",
          "value": "foo"
        }
      ],
      "operations": []
    }
  ]
}')
expected='{"data":[{"name":"test","result":[]}]}'
if [[ "$no_matches" != "$expected" ]]; then
  echo "Error querying no_matches. Got: ${no_matches}, expected: ${expected}"
  exit 1
fi

match_all=$(curl -s -X POST \
  http://localhost:8000/query \
  -H 'Content-Type: application/json' \
  -d '{
  "data": [
    {
      "name": "test",
      "keys": ["dim1", "dim2", "dim3"],
      "filters":[
        {
          "type": "eq",
          "key": "dim1",
          "value": "foo"
        }
      ],
      "operations": []
    }
  ]
}')
expected='{"data":[{"name":"test","result":[{"dim1":"foo","dim2":"bar2","id":0},{"dim1":"foo","dim2":"bar2","dim3":"oof","id":1}]}]}'
if [[ "$match_all" != "$expected" ]]; then
  echo "Error querying match_all. Got: ${match_all}, expected: ${expected}"
  exit 1
fi

match_one=$(curl -s -X POST \
  http://localhost:8000/query \
  -H 'Content-Type: application/json' \
  -d '{
  "data": [
    {
      "name": "test",
      "keys": ["dim1", "dim2"],
      "filters":[
        {
          "type": "eq",
          "key": "dim1",
          "value": "foo"
        },
        {
          "type": "eq",
          "key": "dim3",
          "value": "oof"
        }
      ],
      "operations": []
    }
  ]
}')
expected='{"data":[{"name":"test","result":[{"dim1":"foo","dim2":"bar2","dim3":"oof","id":1}]}]}'
if [[ "$match_one" != "$expected" ]]; then
  echo "Error querying match_one. Got: ${match_one}, expected: ${expected}"
  exit 1
fi

echo "Tests pass."
exit 0
