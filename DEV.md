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
        data: [
            {
                name: 'product_view',
                filters: [
                    { type: "eq", key: "event_tag", value: "product_view" },
                    { type: "all", key: "user_id" },
                    { type: "all", key: "item_colour_id" },
                ],
                unique: ['user_id']
            },
            {
                name: 'add_to_cart',
                filters: [
                    { type: "eq", key: "event_tag", value: "add_to_cart" },
                    { type: "all", key: "user_id" },
                    { type: "all", key: "item_colour_id" },
                ],
                unique: ['user_id']
            }
        ],
        funnel: [
            type: 'exact_order',
            order: [
                'product_view',
                'add_to_cart',
            ]
            match: ['user_id', 'item_colour_id'],
        ]


data: [{
    name: 'product_view'
    keys: ['event_tag', 'user_id', 'item_colour_id],
    filters: [{ type: "eq", key: "event_tag", value: "product_view" }],
    operations: [{ type: "unique", key: "user_id" }]
},{
    name: 'add_to_cart'
    keys: ['event_tag', 'user_id', 'item_colour_id],
    filters: [{ type: "eq", key: "event_tag", value: "add_to_cart" }],
    operations: [{ type: "unique", key: "user_id" }]
}],
funnel: {
    type: 'exact_order',
    data: [
        'product_view',
        'add_to_cart',
    ],
    match: ['user_id', 'item_colour_id']
}



inc event_ids along with timestamp, user_id and item_colour_id for all product_views   <--- count unique on user_id

inc event_ids along with timestamp, user_id and item_colour_id for all add_to_carts

iterate over add_to_cart event_ids. Lookup map by user_id and item_colour_id. If not exist, no conversion. If exist, if event_id after any of event id's, its a conversion.







            { type: "eq", key: "path", value: "/foo" },
            { type: "regex", key: "status_code", value: "5.." }
        ]
    }`


## Performance

RangeKeys is fast and creates a new list of keys using append. There may be some performance
gain by making this list via Stream instead of serially (to be investigated).

The result (slice) from RangeKeys should not be iterated over and copied again.
1. When creating the result, we should decode the key there and then. This can also be done in Stream.
2. Saving the value should be done there and then, so we have a list of structs.
3. Creating the map for intersecting should be done there and then

Intersect - look to implement or use library to use fast intersect via sorting.
Benchmark different methods based on real data.

Multiple filters (where we are then doing an intersect) should be done in goroutines.
Intersection happens once data is back.

Keep around different methods so that we can benchmark them later with real data.

## TODO

Python integration tests:
Use something other than bash for integration testing. We want to write a lot of integration
tests and it should be easy to do so. The integration test script could start the server and
then call out to pytest tests to run the tests (sometimes in parallel).
Can prefix keys with test_name so we don't have to reset the db during the test run.

JSON schema for queries + docs for queries
Use JSON schema to validate queries and document what queries should look like.

Basic UI for querying and visualising results:
- Table view for viewing data from data queries
- How should meta data look like (counts etc)
- Basic graphs for visualising data
- Expose event keys for autocomplete exploration
- Basic statistics showing # of events, ingestion rate and eventually system status

Query to get all keys. e.g.
Get all keys with prefix => event_tag, user_id, etc
Get all keys for event_tag='product_view' =>  Store (key, value)


## DB

Key-value stored with values sorted by key.
