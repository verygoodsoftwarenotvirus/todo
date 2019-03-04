import unittest

from client.v1.python import TodoClient

TEST_BASE_URL = "https://todo-server"


class ItemsTests(unittest.TestCase):
    def __init__(self, methodName="runTest"):
        self.todo_client = TodoClient(
            base_url=TEST_BASE_URL,
            oauth2_client_id="obligatory_client_id",
            oauth2_client_secret="obligatory_client_secret",
            scope="*",
        )
        super(ItemsTests, self).__init__(methodName)

    def test_item_retrieval(self):
        pass
        # self.todo_client.create


if __name__ == "__main__":
    unittest.main()
