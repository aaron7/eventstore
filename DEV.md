# Event Store

## Terms

- `event_id` - unique ID for every ingested event
- `timestamp` - timestamp in ms for an event

## API

### eventstore

- POST `/events`

    `{events: [
        {ts: "", samplerate: "", data: { dimension1: "value1", dimension2: "value2" }},
        {ts: "", samplerate: "", data: { dimension1: "value1", dimension2: "value2" }}
    ]}`

    => 200 OK

- POST `/query`

    `{
        start: "...",
        end: "...",
        filters: [
            { type: "eq", key: "path", value: "/foo" },
            { type: "regex", key: "status_code", value: "5.." }
        ]
    }`

## DB

Key-value stored with values sorted by key.
