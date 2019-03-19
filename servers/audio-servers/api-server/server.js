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


//Authentication check
app.all('/v1/channels*', function (req, res, next) {

  let xUserValue = req.get("X-User")
  if (xUserValue == undefined) {
    const err = new Error('User Not Authenticated');
    err.status = 401;
    res.send("User Not Authenticated");
    return
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

app.get('/v1/audio/iosclient/*', function (req, res) {
  res.sendFile(__dirname + '/ios-client.html');
});

app.patch("/v1/channels/:streamID", function (req, res) {

  let id = req.params.streamID;
  let newName = req.body.displayName;
  let newDescription = req.body.description;

  if( !newName || ! newDescription){
    res.status(400);
    res.send("New displayName and description required")
    return
  }

  Stream.findOneAndUpdate({ channelID: id }, {displayName: newName, description: newDescription}, { new: true }, function (err, response) {
    if (err) {
      res.status(500);
      res.send("Error updating stream")
    } else {
      res.json(response);
    }
  });
});


app.get("/v1/channels/:streamID", function (req, res) {

  let id = req.params.streamID;

  Stream.findOne({ channelID: id }, function (err, response) {
    if (err) {
      res.status(500);
      res.send("Error getting channel")
    } else {
      res.status(200);
      res.json(response);
    }
  });

});

app.all("/v1/channels/:channelID/listeners", function (req, res) {

  let xUserValue = req.get("X-User")

  let id = req.params.channelID;
  let authUserID = JSON.parse(xUserValue).id;

  if (req.method == "POST") {

    Stream.findOneAndUpdate({ channelID: id }, { $addToSet: { activeListeners: authUserID } }, { new: true }, function (err, response) {
      if (err) {
        res.status(500);
        res.send("Error adding active listener to database")
      } else {
        res.send(response);
      }

    });

  } else if (req.method == "DELETE") {

    Stream.findOneAndUpdate({ channelID: id }, { $pullAll: { activeListeners: [authUserID] } }, { new: true }, function (err, response) {
      if (err) {
        res.status(500);
        res.send("Error deleting active listener from database")
      } else {
        res.send(response);
      }

    });
  } else {
    res.status(405)
    res.send("Method Not Allowed")
  }
});

//Post a new channel
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
    let givenDescription = req.body.description;
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
        res.json(broadcast);

      }

    });

  } else {
    //The request body wasnt right
    res.status(400).send("Channel requires a channelID");
  }



});



//Handlers for username queri

//channels?live=bool
//channels?username=""
//channels?

app.get('/v1/channels', function (req, res) {


  let username = req.query.username;
  let live = req.query.live;

  let conditions = {};

  if (username) {
    conditions['creator.userName'] = username;
  }

  if (live === "true") {
    conditions.active = true
  }

  Stream.find(conditions, function (err, response) {

    if (err) {
      res.status(500);
      res.send("Error retriving user streams");
    } else {
      res.json(response)
    }
  });

});

//Gives teh channesl that the currently authenticated user is subscribed to
app.get('/v1/channels/followed', function (req, res) {

  currentUser = JSON.parse(req.header("X-User"));
  currentUserID = currentUser.id;

  Stream.find({ followers: currentUserID}, function (err, response) {

    if (err) {
      res.status(500);
      res.send("Error retriving user streams");
    } else {
      res.json(response)
    }
  });

});

//Allow users to follow a stream 
app.post('/v1/channels/:channelID/followers', function (req, res){
  let currentUser = JSON.parse(req.header("X-User"));
  let currentUserID = currentUser.id;
  let channelID = req.params.channelID;

  Stream.findOneAndUpdate({ channelID: channelID}, {$push : { followers: currentUserID}}, {new: true}, function (err, response) {

    if (err) {
      res.status(500);
      res.send("Error posting new follower");
    } else {
      res.json(response)
    }
  });

});


//Test client things
app.get('/v1/audio/rtcjs', function (req, res) {
  res.sendFile(__dirname + "/dist/RTCMultiConnection.min.js")
});

app.get('/v1/audio/adapter', function (req, res) {
  res.sendFile(__dirname + "/node_modules/webrtc-adapter/out/adapter.js")
});

app.get('/v1/audio/socket', function (req, res) {
  res.sendFile(__dirname + "/node_modules/socket.io-client/dist/socket.io.js")
});

http.listen(port, function () {
  console.log('listening on *:' + port);
});
