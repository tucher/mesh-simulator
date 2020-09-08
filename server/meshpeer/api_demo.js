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
api.registerUserDataUpdateHandler(function(userDataObj){
    //handle updated user data. This will be called from frontend
});
