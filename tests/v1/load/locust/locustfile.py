import random
from http import cookiejar

import requests
import pyotp
from mimesis import Generic
from locust import HttpLocust, TaskSet, task

import client.v1.python as todoclient


INSTANCE_URL = "http://todo-server"


class UserTasks(TaskSet):
    def __init__(self, *args, **kwargs):
        self.created_item_ids = []

        super(UserTasks, self).__init__(*args, **kwargs)

    def on_start(self):
        fake = Generic()
        username = fake.person.username()
        password = fake.person.password(length=32)

        res = self.client.post(
            url="/users", json={"username": username, "password": password}
        )
        user = res.json()

        self.client.post(
            url="/users/login",
            json={
                "username": username,
                "password": password,
                "totp_token": pyotp.TOTP(user.get("two_factor_secret", "")).now(),
            },
        )

    @task(1)
    def health(self):
        self.client.get("/_meta_/health")

    @task(1)
    def get_invalid_item(self):
        with self.client.get(
            url="/api/v1/items/999999999", catch_response=True
        ) as response:
            if response.status_code != 404:
                response.failure("service returned item that is irrelevant")

    @task(5)
    def delete_item(self):
        number_of_items = len(self.created_item_ids)
        if number_of_items > 0:
            unlucky_item = self.created_item_ids.pop(random.randrange(number_of_items))
            self.client.delete(url=f"/api/v1/items/{unlucky_item}")

    @task(10)
    def create_item(self):
        fake = Generic()

        item_creation_input = {
            "name": fake.text.word(),
            "details": fake.text.sentence(),
        }

        res = self.client.post(url="/api/v1/items", json=item_creation_input)
        item_id = res.json().get("id")
        self.created_item_ids.append(item_id)

    @task(3)
    def list_items(self):
        self.client.get(url="/api/v1/items")


class TodoAPIServerLocust(HttpLocust):
    task_set = UserTasks
    min_wait = 1000
    max_wait = 5000
