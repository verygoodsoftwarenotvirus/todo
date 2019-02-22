# /usr/bin/python3

import os
import time
from urllib.parse import urlparse
from http import cookiejar
from typing import List

import requests
import pyotp
from oauthlib.oauth2 import BackendApplicationClient
from requests_oauthlib import OAuth2Session

CLIENT_ID = "HEREISACLIENTIDWHICHIVEMADEUPBECAUSEIWANNATESTRELIABLY"
CLIENT_SECRET = "HEREISASECRETWHICHIVEMADEUPBECAUSEIWANNATESTRELIABLY"
TWO_FACTOR_SECRET = \
    'OAQ56A2GAOZSOTXUWQJJIPRCGWAI4CUYSSASSGFNU36ZLUQVM24KGM6R6GZYOFBOBVOZZPATX6SRNKL55W3PD3QJE5YBUMJNSLJBVJY'

USERNAME = "username"
PASSWORD = "password"


def create_user(username: str, password: str) -> dict:
    res = requests.post(
        url="http://localhost/users",
        json={'username': username, 'password': password},
    )
    return res.json()


def login_user(username: str, password: str, totp_secret: str) -> cookiejar.CookieJar:
    totp = pyotp.TOTP(totp_secret)

    res = requests.post(url="http://localhost/users/login", json={
        'username': username,
        'password': password,
        'totp_token': totp.now(),
    })
    return res.cookies


def build_init_oauth2_client(username: str, password: str, totp_secret: str, cookies: cookiejar.CookieJar) -> dict:
    res = requests.post(
        url="http://localhost/oauth2/init_client",
        cookies=cookies,
        json={
            'username': username,
            'password': password,
            'totp_token': pyotp.TOTP(totp_secret).now(),
        },
    )
    return res.json()


class OAuth2Client:
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
        self._token = None

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

    def update_token(self, new_token: str):
        self._token = new_token

    def get(self, url: str, data: dict = None) -> requests.Response:
        self.sess.token = self.token
        res = self.sess.get(
            url=url,
            json=data,
            headers={
                'Accept': 'application/json',
                'Content-type': 'application/json',
            },
            client_id=self.client_id,
            client_secret=self.client_secret,
        )

        return res

    @property
    def token(self) -> dict:
        if not self._token or (self._token.get('expires_at', 0) - time.time() <= 0):
            t = self.sess.fetch_token(
                token_url=f"{self.base_url}/oauth2/token",
                client_id=self.client_id,
                client_secret=self.client_secret,
                include_client_id=True,
            )
            self._token = t

        return self._token


class TodoClient(OAuth2Client):
    def build_api_url(self, parts: List[str]) -> str:
        url_parts = ['api', 'v1']
        url_parts.extend(parts)
        return f"{self.base_url}/{'/'.join(url_parts)}"

    def get_item(self, identifier: str) -> dict:
        url = self.build_api_url(parts=['item', identifier])
        res = self.get(url=url)
        j = res.json()
        return j

    def get_items(self) -> dict:
        url = self.build_api_url(parts=['items'])
        res = self.get(url=url)
        j = res.json()
        return j


if __name__ == "__main__":
    os.putenv("OAUTHLIB_INSECURE_TRANSPORT", "true")  # SECUREME: removing this would require using HTTPS

    # user = create_user(USERNAME, PASSWORD)
    # two_factor_secret = user.get('two_factor_secret')
    cookie_jar = login_user(USERNAME, PASSWORD, TWO_FACTOR_SECRET)
    oauth2_client = build_init_oauth2_client(USERNAME, PASSWORD, TWO_FACTOR_SECRET, cookie_jar)

    client_id, client_secret = oauth2_client.get("client_id"), oauth2_client.get("client_secret")

    todo_client = TodoClient(
        base_url='http://localhost',
        oauth2_client_id=client_id,
        oauth2_client_secret=client_secret,
        scope='*',
    )

    items = todo_client.get_items()
    print()

