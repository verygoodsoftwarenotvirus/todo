import time
import typing
from urllib.parse import urlparse
from http import cookiejar
from json import JSONEncoder

from .models import Item, OAuth2Client, User
from .http_client import HTTPClient

import requests
import pyotp
from oauthlib.oauth2 import BackendApplicationClient
from requests_oauthlib import OAuth2Session


class TodoClient(HTTPClient):

    def __init__(self, base_url: str, oauth2_client_id: str, oauth2_client_secret: str, scope: str):
        super(TodoClient, self).__init__(
            base_url=base_url,
            oauth2_client_id=oauth2_client_id,
            oauth2_client_secret=oauth2_client_secret,
            scope=scope,
        )

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

        output = []
        raw_items = res.json().get('items', [])
        for raw_item in raw_items:
            item = Item(**raw_item)
            output.append(item)

        return output

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

        output = []
        raw_oauth2_clients = res.json().get('items', [])
        for raw_oauth2_client in raw_oauth2_clients:
            oauth2_client = Item(**raw_oauth2_client)
            output.append(oauth2_client)

        return output

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
        user = User(**res.json())
        return user

    def get_users(self) -> typing.List[User]:
        url = self.build_api_url(parts=['users'])
        res = self.get(url=url)

        output = []
        raw_user = res.json().get('items', [])
        for raw_user in raw_user:
            user = User(**raw_user)
            output.append(user)

        return output

    def update_user(self, user: User):
        url = self.build_api_url(parts=['users', user.id])
        res = self.put(url=url, data=user)
        return Item(**res.json())

    def delete_user(self, identifier: str):
        url = self.build_api_url(parts=['users', identifier])
        self.delete(url=url)

    @staticmethod
    def create_user(self, username: str, password: str) -> User:
        res: requests.Response = requests.post(
            url="http://localhost/users",
            json={"username": username, "password": password}
        )

        user = User(**res.json())
        return user

    @staticmethod
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

