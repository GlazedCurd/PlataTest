import pytest
import requests

# крайне упрощённая версия интеграционных тестов
# сброс базы и проверка начального состояния делается руками
# моков апи с валютами

@pytest.fixture(scope="session")
def api_base_url():
    return "http://localhost:8080"

def test_currency_request(api_base_url):
    pair = "EUR_USD"
    response = requests.post(f"{api_base_url}/quotes/{pair}/update", '{"idempotency_key":"abcdefghij9"}')
    assert response.status_code == 200
    update_id = response.json()["id"]
    response = requests.get(f"{api_base_url}/quotes/{pair}/update/{update_id}")
    assert response.status_code == 200

def test_request_without_last(api_base_url):
    pair = "EUR_MXN"
    response = requests.get(f"{api_base_url}/quotes/{pair}")
    assert response.status_code == 404

def test_idempotancy(api_base_url):
    pair = "USD_MXN"
    idempotency_key = "abcdefghi20"
    response = requests.post(f"{api_base_url}/quotes/{pair}/update", f'{{"idempotency_key":"{idempotency_key}"}}')
    assert response.status_code == 200
    response2 = requests.post(f"{api_base_url}/quotes/{pair}/update", f'{{"idempotency_key":"{idempotency_key}"}}')
    assert response2.status_code == 200
    assert response2.json() == response.json()


def test_idempotancy_conflict(api_base_url):
    pair = "EUR_MXN"
    idempotency_key = "abcdefghi21"
    response = requests.post(f"{api_base_url}/quotes/{pair}/update", f'{{"idempotency_key":"{idempotency_key}"}}')
    assert response.status_code == 200
    pair = "EUR_USD"
    response2 = requests.post(f"{api_base_url}/quotes/{pair}/update", f'{{"idempotency_key":"{idempotency_key}"}}')
    assert response2.status_code == 409
