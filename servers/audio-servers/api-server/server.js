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

app.all('/v1/channels*', function (req, res, next) {

  let xUserValue = req.get("X-User")
  if (xUserValue == undefined) {
    const err = new Error('User Not Authenticated');
    err.status = 401;
    next(err);
  } else {
    next()
  }

});

app.get('/v1/audio/client', function (req, res) {
  res.sendFile(__dirname + '/index.html');
});

app.get('/v1/audio/directclient/*', function (req, res) {
  res.sendFile(__dirname + '/test-client.html');
});

app.patch("/v1/channels/:streamID", function (req, res) {

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

app.all("/v1/channels/:channelID/listeners", function (req, res) {

  let xUserValue = req.get("X-User")

  let id = req.params.channelID;
  let authUserID = JSON.parse(xUserValue).id;

  if (req.method == "POST") {

    Stream.findOneAndUpdate({ channelID: id }, { $addToSet: { activeListeners: authUserID } },{new: true}, function (err, response) {
      if (err) {
        console.log("err updating activelsitner " + err)
        res.status(500);
        res.send("error updating activelistener")
      }else{
        console.log("added active listener")
        res.send(response);
      }

    });

  } else if (req.method == "DELETE") {

    Stream.findOneAndUpdate({ channelID: id }, { $pullAll: { activeListeners: [authUserID] }},{new: true} , function (err, response) {
      if (err) {
        console.log("err updating activelsitner " + err)
        res.status(500);
        res.send("error delting activeliseter")
      }else{
        console.log("deleted active listener successsss")
        res.send(response);
      }

    });


  } else {
    res.status(405)
    res.send("Method Not Allowed")
  }


});

app.post("/v1/channels", function (req, res) {

  //Get the user sending the request
  var currentUser = {};
  if (req.header("X-User")) {
    currentUser = JSON.parse(req.header("X-User"));
  }

  //currentUser = parseInt(req.header("X-User"),10);

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
          active: false,
          activeListeners: [0]
        });

        broadcast.save();
        console.log(broadcast);
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

app.get('/v1/channels/all', function (req, res) {

  Stream.find({}, function (err, response) {

    if (err) {
      console.log(err)
    }

    res.json(response);

  });



});

app.get('/v1/channels/live', function (req, res) {

  Stream.find({ active: true }, function (err, response) {

    if (err) {
      console.log(err)
    }

    res.json(response);

  });



});



//Handlers for username queri

//channels?live=bool
//channels?username=""
//channels?

app.get('/v1/channels', function(req,res){


  let username = req.query.username;
  let live = req.query.live;

  let conditions = {};

  if(username){
    conditions['creator.userName'] = username;
  }

  if( live === "true"){
    conditions.active = true
  }


  console.log(conditions);
  Stream.find(conditions, function (err, response) {

    if (err) {
      console.log(err)
      res.status(500);
      res.send("Error retriving user streams");
    }else{
      res.json(response)
    }

  });


});


http.listen(port, function () {
  console.log('listening on *:' + port);
});
