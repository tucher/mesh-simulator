import requests
url = "http://localhost:8088"
# url = "http://burevestnik.means.live:8887"

script = open("./peer.js", 'r').read()

while True:
    input()
    res = requests.post(url + "/create_peer", json={
        "StartCoord": [53.904153, 27.556925],
        "Script": script,
        "Meta": {"color": "white", "label": "I am JS peer :)"}
        }).json()
    print(res)

    id = res["id"]
    print("New id: ", id)

    input()
    print(requests.post(url + "/delete_peer", json={"ID": id}).json())

