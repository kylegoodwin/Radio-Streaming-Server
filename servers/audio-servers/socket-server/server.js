var app = require('express')();
var fs = require('fs')
var fullchain = process.env.TLSCERT;
var mdport = process.env.MDPORT
var privkey = process.env.TLSKEY; 
var http = require('https').createServer({key: fs.readFileSync(privkey), cert: fs.readFileSync(fullchain)}, app)
var io = require('socket.io')(http);
var port = process.env.PORT || 3001;

const RTCMultiConnectionServer = require('./custom-rtc-server');

/*
var mongoose = require("mongoose");

var db;

var mongourl = "mongodb://" + mdport

mongoose.connect(mongourl)

// Get Mongoose to use the global promise library
mongoose.Promise = global.Promise;
//Get the default connection
var db = mongoose.connection;

//Bind connection to error event (to get notification of connection errors)
db.on('error', console.error.bind(console, 'MongoDB connection error:'));
*/


//var stream = require("./custom-rtc-server/node_scripts/stream-schema")

const jsonPath = {
    config: 'config.json',
    logs: 'logs.json'
};


io.on('connection', function(socket){

    console.log("here")
    RTCMultiConnectionServer.addSocket(socket);

    // ----------------------
    // below code is optional
    console.log("socker ")
    //console.log(socket._events.get-public-rooms);
    const params = socket.handshake.query;
    console.log(params.socketCustomEvent)
    if (!params.socketCustomEvent) {
        
        params.socketCustomEvent = 'custom-message';
    }

    socket.on(params.socketCustomEvent, function (message) {
        console.log("HAPPENING")
        socket.broadcast.emit(params.socketCustomEvent, message);
    });
});

http.listen(port, function(){
  console.log('listening on *:' + port);
});

