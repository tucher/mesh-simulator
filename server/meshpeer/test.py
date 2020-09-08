import requests
url = "http://localhost:8088"
# url = "http://burevestnik.means.live:8887"
script = """
log('started');
var api = new MeshAPI();
var myId = api.getMyID();
log('my ID:', myId);
api.sendMessage('bla', "some message");

api.registerMessageHandler(function(id, data) {
    //log('onMessage: ', id, data);
});

api.registerPeerAppearedHandler(function(id) {
    log('onPeerAppeared: ', id);
});

api.registerPeerDisappearedHandler(function(id) {
    log('onPeerDisappeared: ', id);
});
var currentTS = 0;
api.registerTimeTickHandler(function(ts) {
    currentTS = ts;
    //log('onTimeTick: ', ts);
    api.setDebugMessage(JSON.stringify({MyID: api.getMyID(), MyTS:  currentTS}));
});

"""
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

