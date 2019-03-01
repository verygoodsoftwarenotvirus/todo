import typing
import time
from urllib.parse import urlparse
from json import JSONEncoder

import requests

from oauthlib.oauth2 import BackendApplicationClient
from requests_oauthlib import OAuth2Session

DEFAULT_HEADERS = {
    "Accept": "application/json",
    "Content-type": "application/json",
}

def raise_for_status(func):
    def invoke_request(*args, **kwargs) -> typing.Dict:
        # Invoke the wrapped function first
        res: requests.Response = func(*args, **kwargs)
        res.raise_for_status()
        return res.json()
    return invoke_request


class HTTPClient:
    def __init__(
        self,
        base_url: str,
        oauth2_client_id: str,
        oauth2_client_secret: str,
        scope: str,
    ):
        u = urlparse(base_url, allow_fragments=False)

        self.base_url = u.geturl()
        self.client_id = oauth2_client_id
        self.client_secret = oauth2_client_secret
        self._token: dict = {}

        self.oauth2_client = BackendApplicationClient(
            client_id=self.client_id,
            client_secret=self.client_secret,
            scope=scope,
        )

        self.sess = OAuth2Session(
            client_id=self.client_id,
            client=self.oauth2_client,
            token_updater=self.update_token,
            auto_refresh_url=f"{self.base_url}/oauth2/token",
        )

    def update_token(self, new_token: typing.Dict):
        self._token = new_token

    @raise_for_status
    def get(self, url: str) -> requests.Response:
        self.sess.token = self.token
        res: requests.Response = self.sess.get(
            url=url,
            headers=DEFAULT_HEADERS,
            client_id=self.client_id,
            client_secret=self.client_secret,
        )
        return res

    @raise_for_status
    def put(self, url: str, data: JSONEncoder = None) -> requests.Response:
        self.sess.token = self.token
        res: requests.Response = self.sess.put(
            url=url,
            json=data,
            headers=DEFAULT_HEADERS,
            client_id=self.client_id,
            client_secret=self.client_secret,
        )
        return res

    @raise_for_status
    def post(self, url: str, data: JSONEncoder = None) -> requests.Response:
        self.sess.token = self.token
        res: requests.Response = self.sess.post(
            url=url,
            json=data,
            headers=DEFAULT_HEADERS,
            client_id=self.client_id,
            client_secret=self.client_secret,
        )
        return res

    @raise_for_status
    def delete(self, url: str) -> requests.Response:
        self.sess.token = self.token
        res: requests.Response = self.sess.delete(
            url=url,
            headers=DEFAULT_HEADERS,
            client_id=self.client_id,
            client_secret=self.client_secret,
        )
        return res

    @property
    def token(self) -> typing.Dict:
        if not self._token or (self._token.get("expires_at", 0) - time.time() <= 0):
            self._token: typing.Dict = self.sess.fetch_token(
                token_url=f"{self.base_url}/oauth2/token",
                client_id=self.client_id,
                client_secret=self.client_secret,
                include_client_id=True,
            )
        return self._token

