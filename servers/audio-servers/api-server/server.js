var app = require('express')();
var fs = require('fs')
var mdport = process.env.MDPORT;
var http = require('http').createServer(app);
var port = process.env.PORT || 80;

var mongoose = require("mongoose");

var db;

var mongourl = "mongodb://" + mdport;

mongoose.connect(mongourl)

// Get Mongoose to use the global promise library
mongoose.Promise = global.Promise;
//Get the default connection
var db = mongoose.connection;

var Stream = require("./stream-schema");

app.get('/v1/audio', function(req, res){
  res.sendFile(__dirname + '/index.html');
});


app.get('/v1/audio/rtcjs',function(req,res){
    res.sendFile(__dirname + "/dist/RTCMultiConnection.min.js")
});

app.get('/v1/audio/adapter',function(req,res){
    res.sendFile(__dirname + "/node_modules/webrtc-adapter/out/adapter.js")
});

app.get('/v1/audio/socket',function(req,res){
    res.sendFile(__dirname + "/node_modules/socket.io-client/dist/socket.io.js")
});

app.get('/v1/audio/streams', function(req,res){

  Stream.find({}, function(err, response){

    if(err){
      console.log(err)
    }

    res.json(response);

  });



});


http.listen(port, function(){
  console.log('listening on *:' + port);
});
