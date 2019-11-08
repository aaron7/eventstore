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
    "tag": "tag1",
    "ts": 1001,
    "samplerate": 1,
    "data": {
      "dim1": "foo",
      "dim2": "bar2"
    }
  }, {
    "tag": "tag1",
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
      "tag": "tag1",
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
expected='{"data":[{"name":"test","result":[],"meta":{}}]}'
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
      "tag": "tag1",
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
expected='{"data":[{"name":"test","result":[{"eventID":0,"tag":"tag1","data":[{"key":"dim1","value":"foo"},{"key":"dim2","value":"bar2"}]},{"eventID":1,"tag":"tag1","data":[{"key":"dim1","value":"foo"},{"key":"dim2","value":"bar2"},{"key":"dim3","value":"oof"}]}],"meta":{}}]}'
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
      "tag": "tag1",
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
expected='{"data":[{"name":"test","result":[{"eventID":1,"tag":"tag1","data":[{"key":"dim1","value":"foo"},{"key":"dim3","value":"oof"},{"key":"dim2","value":"bar2"}]}],"meta":{}}]}'
if [[ "$match_one" != "$expected" ]]; then
  echo "Error querying match_one. Got: ${match_one}, expected: ${expected}"
  exit 1
fi

# Test count operation = 2
match_2_count=$(curl -s -X POST \
  http://localhost:8000/query \
  -H 'Content-Type: application/json' \
  -d '{
  "data": [
    {
      "name": "test",
      "tag": "tag1",
      "keys": ["dim1", "dim2", "dim3"],
      "filters":[
        {
          "type": "eq",
          "key": "dim1",
          "value": "foo"
        }
      ],
      "operations": [{
        "type": "count"
      }],
      "hideData": true
    }
  ]
}')
expected='{"data":[{"name":"test","result":[],"meta":{"count":2}}]}'
if [[ "$match_2_count" != "$expected" ]]; then
  echo "Error querying match_2_count. Got: ${match_2_count}, expected: ${expected}"
  exit 1
fi

# Test count operation = 1
match_1_count=$(curl -s -X POST \
  http://localhost:8000/query \
  -H 'Content-Type: application/json' \
  -d '{
  "data": [
    {
      "name": "test",
      "tag": "tag1",
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
      "operations": [{
        "type": "count"
      }],
      "hideData": true
    }
  ]
}')
expected='{"data":[{"name":"test","result":[],"meta":{"count":1}}]}'
if [[ "$match_1_count" != "$expected" ]]; then
  echo "Error querying match_1_count. Got: ${match_1_count}, expected: ${expected}"
  exit 1
fi

# Test unique count
match_unique_count=$(curl -s -X POST \
  http://localhost:8000/query \
  -H 'Content-Type: application/json' \
  -d '{
  "data": [
    {
      "name": "test",
      "tag": "tag1",
      "keys": ["dim1", "dim2", "dim3"],
      "filters":[
        {
          "type": "eq",
          "key": "dim1",
          "value": "foo"
        }
      ],
      "operations": [{
        "type": "uniqueCount",
        "key": "dim1"
      }],
      "hideData": true
    }
  ]
}')
expected='{"data":[{"name":"test","result":[],"meta":{"uniqueCount":1}}]}'
if [[ "$match_unique_count" != "$expected" ]]; then
  echo "Error querying match_unique_count. Got: ${match_unique_count}, expected: ${expected}"
  exit 1
fi

echo "Tests pass."
exit 0
