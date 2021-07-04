import sys
import requests

host = sys.argv[1]
endpoint = "/new"

requests.post(host+endpoint, json={"hostname": "https://googlexxxxxxxxxxxxxxxxxxxxxxxx.com", "onFailWebhook": "http://localhost:3000/webhook", "waitTime": 3})