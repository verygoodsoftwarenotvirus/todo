import typing
import random
from typing import List

import pyotp
from mimesis import Generic
from locust import HttpLocust, TaskSet, task

INSTANCE_URL = "http://todo-server"
API_URL_PREFIX = "/api/v1"

ITEMS_URL_PREFIX = f"{API_URL_PREFIX}/items"
WEBHOOKS_URL_PREFIX = f"{API_URL_PREFIX}/webhooks"
OAUTH2_CLIENTS_URL_PREFIX = f"{API_URL_PREFIX}/oauth2/clients"


class UserTasks(TaskSet):
    def __init__(self, *args, **kwargs):
        self._token: str = ""
        self._oauth2_authorized: bool = False
        self.created_item_ids: List[int] = []
        self.created_webhook_ids: List[int] = []
        self.created_oauth2_client_ids: List[str] = []
        self.username: str = ""
        self.password: str = ""
        self.client_id: str = ""
        self.client_secret: str = ""
        self.two_factor_secret: str = ""
        self.fake = Generic()

        super(UserTasks, self).__init__(*args, **kwargs)

    def on_start(self):
        self.username = self.fake.person.username()
        self.password = self.fake.person.password(length=32)

        res = self.client.post(
            url="/users", json={"username": self.username, "password": self.password}
        )
        user = res.json()
        self.two_factor_secret = user.get("two_factor_secret")
        if not self.two_factor_secret:
            raise Exception("no two factor secret present in response")

        # we make this request because Locust keeps track of the cookie
        self.client.post(
            url="/users/login",
            json={
                "username": self.username,
                "password": self.password,
                "totp_token": pyotp.TOTP(self.two_factor_secret).now(),
            },
        )

        client = self.client.post(
            url="/oauth2/client",
            json={
                "username": self.username,
                "password": self.password,
                "totp_token": pyotp.TOTP(self.two_factor_secret).now(),
                "redirect_uri": "http://localhost:8080",
            },
        ).json()
        self.client_id = client.get("client_id")
        self.client_secret = client.get("client_secret")

    @property
    def token(self):
        # FIXME: :(
        if self._token != "" and self.client_id != "" and self.client_secret != "":
            res = self.client.post(
                url="/oauth2/token",  # authorize
                data={"grant_type": "client_credentials"},
                verify=False,
                allow_redirects=False,
                auth=(self.client_id, self.client_secret),
            )

            print(""" yayayayayayayay authorize hit yayayayayayayay """)
            print(res.content)
            print(res.json())

    @property
    def auth_headers(self) -> typing.Dict:
        # return {"Authorization": f"Bearer {self.token}"}
        return {}

    @task(weight=1)
    def change_password(self):
        new_password = self.fake.person.password(length=32)
        res = self.client.post(
            url="/users/password/new",
            json={
                "new_password": new_password,
                "current_password": self.password,
                "totp_token": pyotp.TOTP(self.two_factor_secret).now(),
            },
        )
        if res.status_code == 200:
            self.password = new_password

    @task(weight=1)
    def change_two_factor_secret(self):
        res = self.client.post(
            url="/users/password/new",
            json={
                "current_password": self.password,
                "totp_token": pyotp.TOTP(self.two_factor_secret).now(),
            },
        )
        if res.status_code == 200:
            try:
                self.two_factor_secret = res.json().get("two_factor_secret")
            except:
                pass

    @task(weight=5)
    def health(self):
        self.client.get("/_meta_/health")

    # Item things

    def random_item_id(self) -> int:
        number_of_items = len(self.created_item_ids)
        if 0 < number_of_items:
            return random.choice(self.created_item_ids)
        else:
            return -1

    @task(weight=100)
    def create_item(self):
        item_creation_input = {
            "name": self.fake.text.word(),
            "details": self.fake.text.sentence(),
        }

        res = self.client.post(
            url=ITEMS_URL_PREFIX, json=item_creation_input, headers=self.auth_headers
        )
        item_id = res.json().get("id")
        self.created_item_ids.append(item_id)

    @task(weight=100)
    def read_item(self):
        item_id = self.random_item_id()
        if item_id > 0:
            self.client.get(
                url=f"{ITEMS_URL_PREFIX}/{item_id}",
                name=f"{ITEMS_URL_PREFIX}/[item_id]",
                headers=self.auth_headers,
            )

    @task(weight=10)
    def read_invalid_item(self):
        u: str = f"{ITEMS_URL_PREFIX}/999999999"
        with self.client.get(url=u, catch_response=True) as response:
            if response.status_code == 404:
                response.success()

    @task(weight=75)
    def update_item(self):
        item_id = self.random_item_id()
        if item_id > 0:
            new_name = self.fake.text.word()
            response = self.client.put(
                url=f"{ITEMS_URL_PREFIX}/{item_id}",
                name=f"{ITEMS_URL_PREFIX}/[item_id]",
                json={"name": new_name},
                headers=self.auth_headers,
            )

            try:
                body = response.json()
                if body.get("name") != new_name:
                    print(body)
                    response.failure("service returned an unchanged item")
            except:
                pass

    @task(weight=10)
    def update_invalid_item(self):
        new_name = self.fake.text.word()
        with self.client.put(
            url=f"{ITEMS_URL_PREFIX}/999999999",
            json={"name": new_name},
            catch_response=True,
        ) as response:
            if response.status_code == 404:
                response.success()

    @task(weight=100)
    def delete_item(self):
        number_of_items = len(self.created_item_ids)
        if number_of_items > 0:
            unlucky_item = self.created_item_ids.pop(random.randrange(number_of_items))
            self.client.delete(
                url=f"{ITEMS_URL_PREFIX}/{unlucky_item}",
                name=f"{ITEMS_URL_PREFIX}/[item_id]",
                headers=self.auth_headers,
            )

    @task(weight=10)
    def delete_invalid_item(self):
        u: str = f"{ITEMS_URL_PREFIX}/999999999"
        with self.client.get(url=u, catch_response=True) as response:
            if response.status_code == 404:
                response.success()

    @task(weight=50)
    def list_items(self):
        self.client.get(
            url=ITEMS_URL_PREFIX, name=ITEMS_URL_PREFIX, headers=self.auth_headers
        )

    @task(weight=5)
    def request_high_offset_items(self):
        self.client.get(
            url=f"{ITEMS_URL_PREFIX}?page=999999&limit=500",
            name=ITEMS_URL_PREFIX,
            headers=self.auth_headers,
        )

    # Webhook things

    def random_webhook_id(self) -> int:
        number_of_webhooks = len(self.created_webhook_ids)
        if 0 < number_of_webhooks:
            return random.choice(self.created_webhook_ids)
        else:
            return -1

    @task(weight=100)
    def create_webhook(self):
        webhook_creation_input = {
            "name": self.fake.text.word(),
            "url": self.fake.internet.home_page(),
            "method": "POST",
        }

        res = self.client.post(
            url=WEBHOOKS_URL_PREFIX,
            json=webhook_creation_input,
            headers=self.auth_headers,
        )

        webhook_id = res.json().get("id")
        self.created_webhook_ids.append(webhook_id)

    @task(weight=100)
    def read_webhook(self):
        webhook_id = self.random_webhook_id()
        if webhook_id > 0:
            self.client.get(
                url=f"{WEBHOOKS_URL_PREFIX}/{webhook_id}",
                name=f"{WEBHOOKS_URL_PREFIX}/[webhook_id]",
                headers=self.auth_headers,
            )

    @task(weight=10)
    def read_invalid_webhook(self):
        u: str = f"{WEBHOOKS_URL_PREFIX}/999999999"
        with self.client.get(url=u, catch_response=True) as response:
            if response.status_code == 404:
                response.success()

    @task(weight=75)
    def update_webhook(self):
        webhook_id = self.random_webhook_id()
        if webhook_id > 0:
            new_name = self.fake.text.word()
            response = self.client.put(
                url=f"{WEBHOOKS_URL_PREFIX}/{webhook_id}",
                name=f"{WEBHOOKS_URL_PREFIX}/[webhook_id]",
                json={"name": new_name},
                headers=self.auth_headers,
            )

            try:
                body = response.json()
                if body.get("name") != new_name:
                    print(body)
                    response.failure("service returned an unchanged webhook")
            except:
                pass

    @task(weight=10)
    def update_invalid_webhook(self):
        new_name = self.fake.text.word()
        with self.client.put(
            url=f"{WEBHOOKS_URL_PREFIX}/999999999",
            json={"name": new_name},
            headers=self.auth_headers,
        ) as response:
            if response.status_code == 404:
                response.success()

    @task(weight=100)
    def delete_webhook(self):
        number_of_webhooks = len(self.created_webhook_ids)
        if number_of_webhooks > 0:
            unlucky_webhook = self.created_webhook_ids.pop(
                random.randrange(number_of_webhooks)
            )
            self.client.delete(
                url=f"{WEBHOOKS_URL_PREFIX}/{unlucky_webhook}",
                name=f"{WEBHOOKS_URL_PREFIX}/[webhook_id]",
                headers=self.auth_headers,
            )

    @task(weight=10)
    def delete_invalid_webhook(self):
        u: str = f"{WEBHOOKS_URL_PREFIX}/999999999"
        with self.client.delete(url=u, catch_response=True) as response:
            if response.status_code == 404:
                response.success()

    @task(weight=50)
    def list_webhooks(self):
        self.client.get(
            url=WEBHOOKS_URL_PREFIX, name=WEBHOOKS_URL_PREFIX, headers=self.auth_headers
        )

    @task(weight=5)
    def request_high_offset_webhooks(self):
        self.client.get(
            url=f"{WEBHOOKS_URL_PREFIX}?page=999999&limit=500",
            name=WEBHOOKS_URL_PREFIX,
            headers=self.auth_headers,
        )

    # OAuth2 client things

    def random_oauth2_client_id(self) -> int:
        number_of_oauth2_clients = len(self.created_oauth2_client_ids)
        if 0 < number_of_oauth2_clients:
            return random.choice(self.created_oauth2_client_ids)
        else:
            return -1

    @task(weight=10)
    def get_invalid_oauth2_client(self):
        u: str = f"{OAUTH2_CLIENTS_URL_PREFIX}/999999999"
        with self.client.get(
            catch_response=True, url=u, headers=self.auth_headers
        ) as response:
            if response.status_code == 404:
                response.success()

    @task(weight=50)
    def create_oauth2_client(self):
        oauth2_client_creation_input = {
            "username": self.username,
            "password": self.password,
            "totp_token": pyotp.TOTP(self.two_factor_secret).now(),
        }

        res = self.client.post(
            url="/oauth2/client",
            json=oauth2_client_creation_input,
            headers=self.auth_headers,
        )
        oauth2_client_db_id = res.json().get("id")
        self.created_oauth2_client_ids.append(oauth2_client_db_id)

    @task(weight=100)
    def read_oauth2_client(self):
        oauth2_client_db_id = self.random_oauth2_client_id()
        if oauth2_client_db_id > 0:
            self.client.get(
                url=f"{OAUTH2_CLIENTS_URL_PREFIX}/{oauth2_client_db_id}",
                name=f"{OAUTH2_CLIENTS_URL_PREFIX}/[oauth2_client_db_id]",
                headers=self.auth_headers,
            )

    @task(weight=50)
    def delete_oauth2_client(self):
        number_of_oauth2_clients = len(self.created_oauth2_client_ids)
        if number_of_oauth2_clients > 0:
            unlucky_oauth2_client = self.created_oauth2_client_ids.pop(
                random.randrange(number_of_oauth2_clients)
            )
            self.client.delete(
                url=f"{OAUTH2_CLIENTS_URL_PREFIX}/{unlucky_oauth2_client}",
                name=f"{OAUTH2_CLIENTS_URL_PREFIX}/[oauth2_client_db_id]",
                headers=self.auth_headers,
            )

    @task(weight=20)
    def list_oauth2_clients(self):
        self.client.get(
            url=OAUTH2_CLIENTS_URL_PREFIX,
            name=OAUTH2_CLIENTS_URL_PREFIX,
            headers=self.auth_headers,
        )

    @task(weight=5)
    def request_high_offset_oauth2_clients(self):
        self.client.get(
            url=f"{OAUTH2_CLIENTS_URL_PREFIX}?page=999999&limit=500",
            name=OAUTH2_CLIENTS_URL_PREFIX,
            headers=self.auth_headers,
        )


class TodoAPIServerLocust(HttpLocust):
    task_set = UserTasks
    min_wait = 1000
    max_wait = 5000
