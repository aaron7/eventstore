-- example HTTP POST script which demonstrates setting the
-- HTTP method, body, and adding a header

wrk.method = "POST"
wrk.body   = '{"events": [{"ts": 1,"samplerate": 1,"data": {"dim1": "foo","dim2": "bar"}}]}'
wrk.headers["Content-Type"] = "application/json"
