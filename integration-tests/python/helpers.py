import requests

EVENT_TS_COUNTER = 1

def create_event(tag=None, ts=None, samplerate=1, data=None):
    if tag is None:
        tag = ''.join(random.choices(string.ascii_lowercase, k=10))

    if ts is None:
        ts=EVENT_TS_COUNTER
        ts += 1

    if data is None:
        data = {}

    return {
        "tag": tag,
        "ts": ts,
        "samplerate": samplerate,
        "data": data,
    }


def send_events(events):
    response = requests.post('http://localhost:8000/events', json={"events": events})
    assert response.status_code == 200
    return response


def execute_query(query):
    response = requests.post('http://localhost:8000/query', json=query)
    assert response.status_code == 200
    return response.json()


def wipe_database():
    response = requests.post('http://localhost:8000/debug', params = {'wipe': True})
    assert response.status_code == 200
