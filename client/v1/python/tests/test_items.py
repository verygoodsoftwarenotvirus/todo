import unittest
import json

import responses

from client.v1.python import TodoClient, Item

TEST_BASE_URL = "https://todo-server.com"


class ItemsTests(unittest.TestCase):
    def __init__(self, methodName="runTest"):
        self.todo_client = TodoClient(
            base_url=TEST_BASE_URL,
            oauth2_client_id="obligatory_client_id",
            oauth2_client_secret="obligatory_client_secret",
            scope="*",
        )
        super(ItemsTests, self).__init__(methodName)

    def test_arbitrary(self):
        self.assertTrue(True)

    @responses.activate
    def test_fetch_item(self):
        expected = Item(identifier=1, name="whatever", details="who_cares")
        example_url = self.todo_client.build_api_url(["items", str(expected.id)])

        responses.add(
            method=responses.POST,
            url=f"{TEST_BASE_URL}/oauth2/token",
            status=200,
            body=json.dumps({"access_token": "access_token"}),
        )

        responses.add(
            method=responses.GET,
            url=example_url,
            status=200,
            body=json.dumps(expected.__dict__),
        )

        actual = self.todo_client.get_item(str(expected.id))

        for k, v in actual.__dict__.items():
            self.assertEqual(expected.__dict__[k], v)


if __name__ == "__main__":
    unittest.main()
