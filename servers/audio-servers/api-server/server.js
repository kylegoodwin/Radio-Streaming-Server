var express = require('express');
var app = express();
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

app.use(express.json());

app.get('/v1/audio', function (req, res) {
  res.sendFile(__dirname + '/index.html');
});


app.patch("/v1/audio/channels/:streamID", function (req, res) {

  let id = req.params.streamID;
  let newName = req.body.name;

  Stream.findOneAndUpdate({ channelID: id }, { displayName: newName }, { new: true }, function (err, doc, response) {

    if (err) {
      console.log(err);
    } else {
      res.json(response);
    }


  });



});



app.post("/v1/audio/channel", function (req, res) {

  //Get the user sending the request
  var currentUser = 0;
  if (req.header("X-User")) {
    currentUser = parseInt(req.header("X-User"), 10);
  }
  //currentUser = parseInt(req.header("X-User"),10);
  console.log("channnelidtest");
  console.log(req.body);
  console.log(req.body.channelID);
  if (req.body.channelID) {

    let givenChannelID = req.body.channelID;
    let givenDisplayName = req.body.channelID;
    let givenDescription = "";
    let givenGenre = "Any";
    let creator = currentUser;


    if (req.body.displayName) {
      givenDisplayName = req.body.displayName;
    }
    if (req.body.description) {
      givenDescription = req.body.discription;
    }
    if (req.body.genre) {
      givenGenre = req.body.genre;
    }





    Stream.findOne({ channelID: req.body.channelID }, function (err, response) {

      //The channel doesnt exist, we can make a new one
      if (!response) {

        let broadcast = new Stream({
          channelID: givenChannelID,
          displayName: givenDisplayName,
          discription: givenDescription,
          genre: givenGenre,
          createdAt: Date.now(),
          creator: creator,
          followers: [],
          active: false
        });

      broadcast.save();

      res.json(broadcast);

      }

    });

  } else {
    //The request body wasnt right
    res.status(400).send("Channel requires a channelID");
  }



});

app.get('/v1/audio/rtcjs', function (req, res) {
  res.sendFile(__dirname + "/dist/RTCMultiConnection.min.js")
});

app.get('/v1/audio/adapter', function (req, res) {
  res.sendFile(__dirname + "/node_modules/webrtc-adapter/out/adapter.js")
});

app.get('/v1/audio/socket', function (req, res) {
  res.sendFile(__dirname + "/node_modules/socket.io-client/dist/socket.io.js")
});

app.get('/v1/audio/channels/all', function (req, res) {

  Stream.find({}, function (err, response) {

    if (err) {
      console.log(err)
    }

    res.json(response);

  });



});

app.get('/v1/audio/channels/live', function (req, res) {

  Stream.find({ active: true }, function (err, response) {

    if (err) {
      console.log(err)
    }

    res.json(response);

  });



});


http.listen(port, function () {
  console.log('listening on *:' + port);
});
