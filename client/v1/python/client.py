import typing
import requests

from .models import (
    Item,
    ItemCreationInput,
    OAuth2Client,
    OAuth2ClientCreationInput,
    User,
    UserCreationInput,
)
from .http_client import HTTPClient


class TodoClient(HTTPClient):
    def __init__(
        self,
        base_url: str,
        oauth2_client_id: str,
        oauth2_client_secret: str,
        scope: str,
    ):
        super(TodoClient, self).__init__(
            base_url=base_url,
            oauth2_client_id=oauth2_client_id,
            oauth2_client_secret=oauth2_client_secret,
            scope=scope,
        )

    def build_unauthenticated_url(self, parts: typing.List[str]) -> str:
        return f"{self.base_url}/{'/'.join(parts)}"

    def build_api_url(self, parts: typing.List[str]) -> str:
        url_parts = ["api", "v1"]
        url_parts.extend(parts)
        return f"{self.base_url}/{'/'.join(url_parts)}"

    def health_check(self):
        self.get(f"{self.base_url}/_meta_/health")

    # Item functions

    def create_item(self, creation_input: ItemCreationInput) -> Item:
        url = self.build_api_url(parts=["items"])
        res = self.post(url=url, data=creation_input.__dict__)

        return Item(**res)

    def get_item(self, identifier: str) -> Item:
        url = self.build_api_url(parts=["items", identifier])
        res = self.get(url=url)
        return Item(**res)

    def get_items(self) -> typing.List[Item]:
        url = self.build_api_url(parts=["items"])
        res = self.get(url=url)

        output = []
        raw_items = res.get("items", [])
        for raw_item in raw_items:
            item = Item(**raw_item)
            output.append(item)

        return output

    def update_item(self, item: Item):
        url = self.build_api_url(parts=["items", str(item.id)])
        res = self.put(url=url, data=item)
        return Item(**res)

    def delete_item(self, identifier: str):
        url = self.build_api_url(parts=["items", identifier])
        self.delete(url=url)

    # OAuth2 client functions

    def create_oauth2_client(
        self, creation_input: OAuth2ClientCreationInput
    ) -> OAuth2Client:
        url = self.build_api_url(parts=["items"])
        res = self.post(url=url, data=creation_input.__dict__)

        return OAuth2Client(**res)

    def get_oauth2_client(self, identifier: str) -> OAuth2Client:
        url = self.build_api_url(parts=["oauth2", "clients", identifier])
        res = self.get(url=url)
        return OAuth2Client(**res)

    def get_oauth2_clients(self) -> typing.List[OAuth2Client]:
        url = self.build_api_url(parts=["oauth2", "clients"])
        res = self.get(url=url)

        output = []
        raw_oauth2_clients = res.get("items", [])
        for raw_oauth2_client in raw_oauth2_clients:
            oauth2_client = OAuth2Client(**raw_oauth2_client)
            output.append(oauth2_client)

        return output

    def update_oauth2_client(self, client: OAuth2Client):
        url = self.build_api_url(parts=["oauth2", "clients", client.client_id])
        res = self.put(url=url, data=client)
        return Item(**res)

    def delete_oauth2_client(self, identifier: str):
        url = self.build_api_url(parts=["oauth2", "clients", identifier])
        self.delete(url=url)

    # User functions

    def create_user(self, creation_input: UserCreationInput) -> User:
        url = self.build_unauthenticated_url(parts=["users"])
        res = self.post(url=url, data=creation_input.__dict__)

        return User(**res)

    def get_user(self, identifier: str) -> User:
        url = self.build_api_url(parts=["users", identifier])
        res = self.get(url=url)
        user = User(**res)
        return user

    def get_users(self) -> typing.List[User]:
        url = self.build_api_url(parts=["users"])
        res = self.get(url=url)

        output = []
        raw_user = res.get("items", [])
        for raw_user in raw_user:
            user = User(**raw_user)
            output.append(user)

        return output

    def update_user(self, user: User):
        url = self.build_api_url(parts=["users", str(user.id)])
        res = self.put(url=url, data=user)
        return Item(**res)

    def delete_user(self, identifier: str):
        url = self.build_api_url(parts=["users", identifier])
        self.delete(url=url)

