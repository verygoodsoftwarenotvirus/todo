# /usr/bin/python3

import time
import typing
from urllib.parse import urlparse
from http import cookiejar
from json import JSONEncoder

from .models import Item, OAuth2Client, User

import requests
import pyotp
from oauthlib.oauth2 import BackendApplicationClient
from requests_oauthlib import OAuth2Session


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
            headers={
                "Accept": "application/json",
                "Content-type": "application/json",
            },
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
            headers={
                "Accept": "application/json",
                "Content-type": "application/json",
            },
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
            headers={
                "Accept": "application/json",
                "Content-type": "application/json",
            },
            client_id=self.client_id,
            client_secret=self.client_secret,
        )
        return res

    @raise_for_status
    def delete(self, url: str) -> requests.Response:
        self.sess.token = self.token
        res: requests.Response = self.sess.delete(
            url=url,
            headers={
                "Accept": "application/json",
                "Content-type": "application/json",
            },
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


class TodoClient(HTTPClient):
    def build_api_url(self, parts: typing.List[str]) -> str:
        url_parts = ['api', 'v1']
        url_parts.extend(parts)
        return f"{self.base_url}/{'/'.join(url_parts)}"

    # Item functions

    def get_item(self, identifier: str) -> Item:
        url = self.build_api_url(parts=['items', identifier])
        res = self.get(url=url)
        return Item(**res.json())

    def get_items(self) -> typing.List[Item]:
        url = self.build_api_url(parts=['items'])
        res = self.get(url=url)
        return [Item(**x) for x in res.json().get('items', [])]

    def update_item(self, item: Item):
        url = self.build_api_url(parts=['items', str(item.id)])
        res = self.put(url=url, data=item)
        return Item(**res.json())

    def delete_item(self, identifier: str):
        url = self.build_api_url(parts=['items', identifier])
        self.delete(url=url)

    # OAuth2 client functions

    def get_oauth2_client(self, identifier: str) -> OAuth2Client:
        url = self.build_api_url(parts=['oauth2', 'clients', identifier])
        res = self.get(url=url)
        return OAuth2Client(**res.json())

    def get_oauth2_clients(self) -> typing.List[OAuth2Client]:
        url = self.build_api_url(parts=['oauth2', 'clients'])
        res = self.get(url=url)
        return [OAuth2Client(**x) for x in res.json().get('oauth2', 'clients', [])]

    def update_oauth2_client(self, client: OAuth2Client):
        url = self.build_api_url(parts=['oauth2', 'clients', client.client_id])
        res = self.put(url=url, data=client)
        return Item(**res.json())

    def delete_oauth2_client(self, identifier: str):
        url = self.build_api_url(parts=['oauth2', 'clients', identifier])
        self.delete(url=url)

    # User functions

    def get_user(self, identifier: str) -> User:
        url = self.build_api_url(parts=['users', identifier])
        res = self.get(url=url)
        return User(**res.json())

    def get_users(self) -> typing.List[User]:
        url = self.build_api_url(parts=['users'])
        res = self.get(url=url)
        return [User(**x) for x in res.json().get('users', [])]

    def update_user(self, user: User):
        url = self.build_api_url(parts=['users', user.id])
        res = self.put(url=url, data=user)
        return Item(**res.json())

    def delete_user(self, identifier: str):
        url = self.build_api_url(parts=['users', identifier])
        self.delete(url=url)

# misc


def create_user(username: str, password: str) -> typing.Dict:
    res: requests.Response = requests.post(
        url="http://localhost/users",
        json={"username": username, "password": password}
    )
    return res.json()


def login_user(username: str, password: str, totp_secret: str) -> cookiejar.CookieJar:
    totp = pyotp.TOTP(totp_secret)

    res: requests.Response = requests.post(
        url="http://localhost/users/login",
        json={
            "username": username,
            "password": password,
            "totp_token": totp.now(),
        },
    )
    return res.cookies


def build_init_oauth2_client(
    username: str,
    password: str,
    totp_secret: str,
    cookies: cookiejar.CookieJar,
) -> dict:
    res: requests.Response = requests.post(
        url="http://localhost/oauth2/init_client",
        cookies=cookies,
        json={
            "username": username,
            "password": password,
            "totp_token": pyotp.TOTP(totp_secret).now(),
        },
    )
    return res.json()


# if __name__ == "__main__":
#     os.putenv(
#         "OAUTHLIB_INSECURE_TRANSPORT", "true"
#     )  # SECUREME: removing this would require using HTTPS

#     # user = create_user(USERNAME, PASSWORD)
#     # two_factor_secret = user.get('two_factor_secret')
#     cookie_jar = login_user(USERNAME, PASSWORD, TWO_FACTOR_SECRET)
#     oauth2_client = build_init_oauth2_client(
#         USERNAME, PASSWORD, TWO_FACTOR_SECRET, cookie_jar
#     )

#     client_id, client_secret = (
#         oauth2_client.get("client_id"),
#         oauth2_client.get("client_secret"),
#     )

#     todo_client = TodoClient(
#         base_url="http://localhost",
#         oauth2_client_id=client_id,
#         oauth2_client_secret=client_secret,
#         scope="*",
#     )

#     items = todo_client.get_items()
#     print()

