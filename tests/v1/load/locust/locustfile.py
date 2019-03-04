from http import cookiejar

import requests
import pyotp
from faker import Faker
from faker.providers import internet, misc
from locust import HttpLocust, TaskSet, task

import client.v1.python as todoclient


INSTANCE_URL = "http://todo-server"


class UserTasks(TaskSet):
    def __init__(self, *args, **kwargs):
        super(UserTasks, self).__init__(*args, **kwargs)

    def on_start(self):
        fake = Faker()
        fake.add_provider(misc)
        fake.add_provider(internet)

        username = fake.user_name()
        password = fake.password(
            length=32, special_chars=True, digits=True, upper_case=True, lower_case=True
        )

        res = self.client.post(
            url="/users", json={"username": username, "password": password}
        )
        user = res.json()

        self.client.post(
            "/users/login",
            {
                "username": username,
                "password": password,
                "totp_token": pyotp.TOTP(user.get("two_factor_secret", "")).now(),
            },
        )

    def create_user(
        self, base_url: str, username: str, password: str
    ) -> todoclient.User:
        res: requests.Response = self.client.post(
            url=f"{base_url}/users", json={"username": username, "password": password}
        )

        user = todoclient.User(**res.json())
        return user

    def login_user(
        self, base_url: str, username: str, password: str, totp_secret: str
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

    @task(1)
    def health(self):
        self.client.get("/_meta_/health")

    # @task(10)
    # def create_item(self):
    #     self.todo_client.health_check()


class TodoAPIServerLocust(HttpLocust):
    task_set = UserTasks
    min_wait = 1000
    max_wait = 5000


###
