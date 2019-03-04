from http import cookiejar

import requests
import pyotp
from faker import Faker
from faker.providers import internet, misc

import client.v1.python as todoclient


def build_obligatory_client(base_url: str) -> todoclient.TodoClient:
    fake = Faker()
    fake.add_provider(internet)
    fake.add_provider(misc)

    pw = fake.password(
        length=32, special_chars=True, digits=True, upper_case=True, lower_case=True
    )
    user = create_user(base_url, fake.user_name(), pw)
    oac = create_oauth2_client(base_url, user.username, pw, user.two_factor_secret)

    c = todoclient.TodoClient(
        base_url=base_url,
        oauth2_client_id=oac.client_id,
        oauth2_client_secret=oac.client_secret,
        scope=",".join(oac.scopes),
    )
    return c


def create_user(base_url: str, username: str, password: str) -> todoclient.User:
    res: requests.Response = requests.post(
        url=f"{base_url}/users", json={"username": username, "password": password}
    )

    user = todoclient.User(**res.json())
    return user


def login_user(
    base_url: str, username: str, password: str, totp_secret: str
) -> cookiejar.CookieJar:
    res: requests.Response = requests.post(
        url=f"{base_url}/users/login",
        json={
            "username": username,
            "password": password,
            "totp_token": pyotp.TOTP(totp_secret).now(),
        },
    )
    return res.cookies


def create_oauth2_client(
    base_url: str, username: str, password: str, totp_secret: str
) -> todoclient.OAuth2Client:
    cookies = login_user(base_url, username, password, totp_secret)

    res: requests.Response = requests.post(
        url=f"{base_url}/oauth2/client",
        cookies=cookies,
        json={
            "username": username,
            "password": password,
            "totp_token": pyotp.TOTP(totp_secret).now(),
        },
    )
    oac = todoclient.OAuth2Client(**res.json())
    return oac
