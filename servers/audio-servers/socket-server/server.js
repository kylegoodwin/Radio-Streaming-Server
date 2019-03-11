var app = require('express')();
var fs = require('fs')
var fullchain = process.env.TLSCERT;
var mdport = process.env.MDPORT
var privkey = process.env.TLSKEY; 
var http = require('https').createServer({key: fs.readFileSync(privkey), cert: fs.readFileSync(fullchain)}, app)
var io = require('socket.io')(http);
var port = process.env.PORT || 3001;

const RTCMultiConnectionServer = require('./custom-rtc-server');


const jsonPath = {
    config: 'config.json',
    logs: 'logs.json'
};


io.on('connection', function(socket){

    RTCMultiConnectionServer.addSocket(socket);

    // ----------------------
    // below code is optional

    const params = socket.handshake.query;
    console.log(params.socketCustomEvent)
    if (!params.socketCustomEvent) {
        params.socketCustomEvent = 'custom-message';
    }

    socket.on(params.socketCustomEvent, function (message) {
        console.log(message)
        socket.broadcast.emit(params.socketCustomEvent, message);
    });
});

http.listen(port, function(){
  console.log('listening on *:' + port);
});

