from unittest import mock

from . import helpers


class TestBasic:
    def setup_class(self):
        helpers.wipe_database()
        events = [
            helpers.create_event("tag1", ts=1001, data={"dim1": "foo", "dim2": "bar2"}),
            helpers.create_event(
                "tag1", ts=1002, data={"dim1": "foo", "dim2": "bar2", "dim3": "oof"}
            ),
        ]
        helpers.send_events(events)

    def test_match_all(self):
        result = helpers.execute_query(
            {
                "data": [
                    {
                        "name": "test",
                        "tag": "tag1",
                        "keys": ["dim1", "dim2", "dim3"],
                        "filters": [{"type": "eq", "key": "dim1", "value": "foo"}],
                        "operations": [],
                    }
                ]
            }
        )

        assert result["data"][0]["name"] == "test"
        assert result["data"][0]["meta"] == {}
        assert result["data"][0]["result"] == [
            {
                "id": mock.ANY,
                "ts": 1001,
                "tag": "tag1",
                "data": [
                    {"key": "dim1", "value": "foo"},
                    {"key": "dim2", "value": "bar2"},
                ],
            },
            {
                "id": mock.ANY,
                "ts": 1002,
                "tag": "tag1",
                "data": [
                    {"key": "dim1", "value": "foo"},
                    {"key": "dim2", "value": "bar2"},
                    {"key": "dim3", "value": "oof"},
                ],
            },
        ]

    def test_no_matches(self):
        result = helpers.execute_query(
            {
                "data": [
                    {
                        "name": "test",
                        "tag": "tag1",
                        "keys": ["dim1", "dim2"],
                        "filters": [
                            {"type": "eq", "key": "dim1", "value": "foo"},
                            {"type": "eq", "key": "dim2", "value": "foo"},
                        ],
                        "operations": [],
                    }
                ]
            }
        )
        assert len(result["data"][0]["result"]) == 0

    def test_match_one(self):
        result = helpers.execute_query(
            {
                "data": [
                    {
                        "name": "test",
                        "tag": "tag1",
                        "keys": ["dim1", "dim2"],
                        "filters": [
                            {"type": "eq", "key": "dim1", "value": "foo"},
                            {"type": "eq", "key": "dim3", "value": "oof"},
                        ],
                        "operations": [],
                    }
                ]
            }
        )

        assert result["data"][0]["meta"] == {}
        assert result["data"][0]["result"] == [
            {
                "id": mock.ANY,
                "ts": 1002,
                "tag": "tag1",
                "data": [
                    {"key": "dim1", "value": "foo"},
                    {"key": "dim3", "value": "oof"},
                    {"key": "dim2", "value": "bar2"},
                ],
            },
        ]

    def test_count_all(self):
        result = helpers.execute_query(
            {
                "data": [
                    {
                        "name": "test",
                        "tag": "tag1",
                        "keys": ["dim1", "dim2", "dim3"],
                        "filters": [{"type": "eq", "key": "dim1", "value": "foo"}],
                        "operations": [{"type": "count"}],
                        "hideData": True,
                    }
                ]
            }
        )

        assert result["data"][0]["meta"] == {"count": 2}

    def test_count_one(self):
        result = helpers.execute_query(
            {
                "data": [
                    {
                        "name": "test",
                        "tag": "tag1",
                        "keys": ["dim1", "dim2"],
                        "filters": [
                            {"type": "eq", "key": "dim1", "value": "foo"},
                            {"type": "eq", "key": "dim3", "value": "oof"},
                        ],
                        "operations": [{"type": "count"}],
                        "hideData": True,
                    }
                ]
            }
        )

        assert result["data"][0]["meta"] == {"count": 1}

    def test_unique_count(self):
        result = helpers.execute_query(
            {
                "data": [
                    {
                        "name": "test",
                        "tag": "tag1",
                        "keys": ["dim1", "dim2", "dim3"],
                        "filters": [{"type": "eq", "key": "dim1", "value": "foo"}],
                        "operations": [{"type": "uniqueCount", "key": "dim1"}],
                        "hideData": True,
                    }
                ]
            }
        )

        assert result["data"][0]["meta"] == {"uniqueCount": 1}


class TestRegex:
    def setup_class(self):
        helpers.wipe_database()
        events = [
            helpers.create_event("tag1", ts=1001, data={"a": "foo1",}),
            helpers.create_event("tag1", ts=1002, data={"a": "foobar"}),
            helpers.create_event("tag1", ts=1003, data={"a": "foo2"}),
        ]
        helpers.send_events(events)

    def test_basic_regex(self):
        result = helpers.execute_query(
            {
                "data": [
                    {
                        "name": "test",
                        "tag": "tag1",
                        "keys": ["dim1", "dim2"],
                        "filters": [{"type": "regex", "key": "a", "value": "foo\d"},],
                        "operations": [],
                    }
                ]
            }
        )

        assert result["data"][0]["result"] == [
            {
                "id": mock.ANY,
                "ts": 1001,
                "tag": "tag1",
                "data": [{"key": "a", "value": "foo1"},],
            },
            {
                "id": mock.ANY,
                "ts": 1003,
                "tag": "tag1",
                "data": [{"key": "a", "value": "foo2"},],
            },
        ]
