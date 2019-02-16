# /usr/bin/python3

from oauthlib.oauth2 import BackendApplicationClient
from requests_oauthlib import OAuth2Session

CLIENT_ID = "HEREISACLIENTIDWHICHIVEMADEUPBECAUSEIWANNATESTRELIABLY"
CLIENT_SECRET = "HEREISASECRETWHICHIVEMADEUPBECAUSEIWANNATESTRELIABLY"


if __name__ == "__main__":
    client = BackendApplicationClient(client_id=CLIENT_ID)
    oauth = OAuth2Session(client_id=CLIENT_ID, client=client)
    token = oauth.fetch_token(
        token_url="https://provider.com/oauth2/token",
        client_id=CLIENT_ID,
        client_secret=CLIENT_SECRET,
    )
