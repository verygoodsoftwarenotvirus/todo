import typing
from json import JSONEncoder


class Base(JSONEncoder):
    def default(self, o):
        return o.__dict__


class Item(Base):
    def __init__(
        self,
        identifier: int = 0,
        name: str = '',
        details: str = '',
        created_on: int = 0,
        updated_on: typing.Optional[int] = None,
        completed_on: typing.Optional[int] = None,
        belongs_to: int = 0,
        **kwargs,
    ):
        super(Item, self).__init__()
        self.id: int = identifier or kwargs.get('id', 0)
        self.name: str = name
        self.details: str = details
        self.created_on: int = created_on
        self.updated_on: typing.Optional[int] = updated_on
        self.completed_on: typing.Optional[int] = completed_on
        self.belongs_to: int = belongs_to


class User(Base):
    def __init__(
        self,
        identifier: int = 0,
        username: str = '',
        two_factor_secret: str = '',
        is_admin: bool = False,
        created_on: int = 0,
        password_last_changed_on: typing.Optional[int] = None,
        updated_on: typing.Optional[int] = None,
        archived_on: typing.Optional[int] = None,
        belongs_to: int = 0,
        **kwargs,
    ):
        super(User, self).__init__()
        self.id: int = identifier or kwargs.get('id', 0)
        self.username: str = username
        self.two_factor_secret: str = two_factor_secret
        self.is_admin: bool = is_admin
        self.created_on: int = created_on
        self.password_last_changed_on: typing.Optional[int] = password_last_changed_on
        self.updated_on: typing.Optional[int] = updated_on
        self.archived_on: typing.Optional[int] = archived_on
        self.belongs_to: int = belongs_to


class OAuth2Client(Base):
    def __init__(
        self,
        identifier: int = 0,
        client_id: str = '',
        client_secret: str = '',
        redirect_uri: str = '',
        scopes: typing.List[str] = None,
        implicit_allowed: bool = False,
        created_on: int = 0,
        updated_on: typing.Optional[int] = None,
        archived_on: typing.Optional[int] = None,
        belongs_to: int = 0,
        **kwargs,
    ):
        super(OAuth2Client, self).__init__()
        self.id: int = identifier or kwargs.get('id', 0)
        self.client_id: str = client_id
        self.client_secret: str = client_secret
        self.redirect_uri: str = redirect_uri
        self.scopes: typing.List[str] = scopes or []
        self.implicit_allowed: bool = implicit_allowed
        self.created_on: int = created_on
        self.updated_on: typing.Optional[int] = updated_on
        self.archived_on: typing.Optional[int] = archived_on
        self.belongs_to: int = belongs_to
