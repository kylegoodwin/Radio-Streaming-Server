// Muaz Khan      - www.MuazKhan.com
// MIT License    - www.WebRTC-Experiment.com/licence
// Documentation  - github.com/muaz-khan/RTCMultiConnection

// pushLogs is used to write error logs into logs.json
var pushLogs = require('./support/pushLogs.js');

var users = {};
var broadcasts = [];
const Stream = require("./stream-schema")
const Request = require("./../../node_modules/request")

var mdport = process.env.MDPORT;

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

module.exports = exports = function (config, socket, maxRelayLimitPerUser) {
    try {
        maxRelayLimitPerUser = parseInt(maxRelayLimitPerUser) || 2;
    } catch (e) {
        maxRelayLimitPerUser = 2;
    }

    socket.on('join-broadcast', function (user) {
        console.log("rawuser")
        console.log(user)

        var headers = {
            'Authorization': user.gatewayAuth
        }

        // Configure the request
        var options = {
            url: 'https://audio-api.kjgoodwin.me/v1/users/me',
            method: 'GET',
            headers: headers
        }

        let authUserID = "";

        Request(options, function (error, response, body) {
            if (!error && response.statusCode == 200) {
                // Print out the response body
                console.log(" WE MADE IT INSIDE OF THE AUTH CHECK")
                console.log(body)
                authUserID = JSON.parse(body).id
                console.log(authUserID)
            }

            try {
                if (!users[user.userid]) {
                    socket.userid = user.userid;
                    socket.isScalableBroadcastSocket = true;

                    users[user.userid] = {
                        userid: user.userid,
                        broadcastId: user.broadcastId,
                        isBroadcastInitiator: false,
                        maxRelayLimitPerUser: maxRelayLimitPerUser,
                        relayReceivers: [],
                        receivingFrom: null,
                        canRelay: false,
                        typeOfStreams: user.typeOfStreams || {
                            audio: true,
                            video: true
                        },
                        socket: socket
                    };

                    notifyBroadcasterAboutNumberOfViewers(user.broadcastId);
                }

                var relayUser = getFirstAvailableBroadcaster(user.broadcastId, maxRelayLimitPerUser);

                if (relayUser === 'ask-him-rejoin') {
                    socket.emit('rejoin-broadcast', user.broadcastId);
                    return;
                }

                if (relayUser && user.userid !== user.broadcastId) {
                    var hintsToJoinBroadcast = {
                        typeOfStreams: relayUser.typeOfStreams,
                        userid: relayUser.userid,
                        broadcastId: relayUser.broadcastId,
                        authUserID: authUserID
                    };

                    users[user.userid].receivingFrom = relayUser.userid;
                    users[relayUser.userid].relayReceivers.push(
                        users[user.userid]
                    );
                    users[user.broadcastId].lastRelayuserid = relayUser.userid;
                    console.log("auth toekn print")
                    console.log(user.authToken)

                    /*
                    if( authUserID != ""){
                        Stream.findOneAndUpdate( { channelID: user.broadcastId}, { $addToSet: {activeListeners: authUserID }}, function(err,response){
                            if( err){
                                console.log("err updating activelsitner " + err)
                            }
                            console.log("added active listener")
                        });


                    }
                    */

                    socket.emit('join-broadcaster', hintsToJoinBroadcast);

                    // logs for current socket
                    socket.emit('logs', 'You <' + user.userid + '> are getting data/stream from <' + relayUser.userid + '>');

                    // logs for target relaying user
                    relayUser.socket.emit('logs', 'You <' + relayUser.userid + '>' + ' are now relaying/forwarding data/stream to <' + user.userid + '>');




                } else {
                    // This is when the user begins a broadcast
                    console.log("user:")
                    console.log(user)

                    // broadcasts.push(user.broadcastId);
                    users[user.userid].isBroadcastInitiator = true;


                    if (authUserID != "") {

                        Stream.findOneAndUpdate({ channelID: user.broadcastId, "creator.id": authUserID }, { active: true, goLiveTime: Date.now() }, function (err, response) {

                            if (err) {
                                console.log("error updating the current socket, you are not the owner of the stream!");
                                socket.emit('failed-broadcast-start', 'you do not own the broadcast');
                            } else if (response) {
                                socket.emit('start-broadcasting', users[user.userid].typeOfStreams);
                                // logs to tell he is now broadcast initiator
                                socket.emit('logs', 'You <' + user.userid + '> are now serving the broadcast.');
                            } else {
                                socket.emit('broadcast-doesnt-exist', 'broadcast not found in database');
                            }

                            console.log("stream successfully posted as active");

                        });
                    } else {
                        socket.emit('failed-broadcast-start', 'you do not own the broadcast');
                    }



                }
            } catch (e) {
                pushLogs(config, 'join-broadcast', e);
            }

        });
    });

    socket.on('scalable-broadcast-message', function (message) {
        socket.broadcast.emit('scalable-broadcast-message', message);
    });

    socket.on('can-relay-broadcast', function () {
        if (users[socket.userid]) {
            users[socket.userid].canRelay = true;
        }
    });

    socket.on('can-not-relay-broadcast', function () {
        if (users[socket.userid]) {
            users[socket.userid].canRelay = false;
        }
    });

    socket.on('check-broadcast-presence', function (userid, callback) {
        // we can pass number of viewers as well
        try {
            callback(!!users[userid] && users[userid].isBroadcastInitiator === true);
        } catch (e) {
            pushLogs(config, 'check-broadcast-presence', e);
        }
    });

    socket.on('get-number-of-users-in-specific-broadcast', function (broadcastId, callback) {
        try {
            if (!broadcastId || !callback) return;

            if (!users[broadcastId]) {
                callback(0);
                return;
            }

            callback(getNumberOfBroadcastViewers(broadcastId));
        } catch (e) { }
    });

    function getNumberOfBroadcastViewers(broadcastId) {
        try {
            var numberOfUsers = 0;
            Object.keys(users).forEach(function (uid) {
                var user = users[uid];
                if (user.broadcastId === broadcastId) {
                    numberOfUsers++;
                }
            });
            return numberOfUsers - 1;
        } catch (e) {
            return 0;
        }
    }

    function notifyBroadcasterAboutNumberOfViewers(broadcastId, userLeft) {
        try {
            if (!broadcastId || !users[broadcastId] || !users[broadcastId].socket) return;
            var numberOfBroadcastViewers = getNumberOfBroadcastViewers(broadcastId);

            if (userLeft === true) {
                numberOfBroadcastViewers--;
            }

            users[broadcastId].socket.emit('number-of-broadcast-viewers-updated', {
                numberOfBroadcastViewers: numberOfBroadcastViewers,
                broadcastId: broadcastId
            });
        } catch (e) { }
    }

    // this even is called from "signaling-server.js"
    socket.ondisconnect = function () {

        console.log("THE DISCCONET IS ALSO HAPPENING HERE")
        try {
            if (!socket.isScalableBroadcastSocket) return;

            var user = users[socket.userid];

            if (!user) return;

            if (user.isBroadcastInitiator === false) {
                notifyBroadcasterAboutNumberOfViewers(user.broadcastId, true);
            } else {

                var index = broadcasts.indexOf(user.broadcastId);
                if (index > -1) {
                    broadcasts.splice(index, 1);
                }

                /*
                Stream.deleteOne({channelID: user.broadcastId}, function(err){

                    if(err){
                        console.log("error, stream was not successfully deleted");
                    }

                    console.log("stream successfully deleted");
                });
                */


                Stream.findOneAndUpdate({ channelID: user.broadcastId }, { active: false }, function (err) {

                    if (err) {
                        console.log("error, stream was not successfully deleted");
                    }

                    console.log("stream successfully went down");
                });




            }

            if (user.isBroadcastInitiator === true) {
                // need to stop entire broadcast?
                for (var n in users) {
                    var _user = users[n];

                    if (_user.broadcastId === user.broadcastId) {
                        _user.socket.emit('broadcast-stopped', user.broadcastId);
                    }
                }

                delete users[socket.userid];
                return;
            }


            if (user.receivingFrom || user.isBroadcastInitiator === true) {

                //This is the list of
                var parentUser = users[user.receivingFrom];

                if (parentUser) {
                    var newArray = [];
                    parentUser.relayReceivers.forEach(function (n) {
                        if (n.userid !== user.userid) {
                            newArray.push(n);
                        }
                    });
                    users[user.receivingFrom].relayReceivers = newArray;
                }
            }

            if (user.relayReceivers.length && user.isBroadcastInitiator === false) {
                askNestedUsersToRejoin(user.relayReceivers);
            }

            delete users[socket.userid];
        } catch (e) {
            pushLogs(config, 'scalable-broadcast-disconnect', e);
        }
    };

    return {
        getUsers: function () {
            try {
                var list = [];
                Object.keys(users).forEach(function (uid) {
                    var user = users[uid];
                    if (!user) return;

                    try {
                        var relayReceivers = [];
                        user.relayReceivers.forEach(function (s) {
                            relayReceivers.push(s.userid);
                        });

                        list.push({
                            userid: user.userid,
                            broadcastId: user.broadcastId,
                            isBroadcastInitiator: user.isBroadcastInitiator,
                            maxRelayLimitPerUser: user.maxRelayLimitPerUser,
                            relayReceivers: relayReceivers,
                            receivingFrom: user.receivingFrom,
                            canRelay: user.canRelay,
                            typeOfStreams: user.typeOfStreams
                        });
                    }
                    catch (e) {
                        pushLogs('getUsers', e);
                    }
                });
                return list;
            }
            catch (e) {
                pushLogs('getUsers', e);
            }
        }
    };
};

function askNestedUsersToRejoin(relayReceivers) {
    try {
        var usersToAskRejoin = [];

        relayReceivers.forEach(function (receiver) {
            if (!!users[receiver.userid]) {
                users[receiver.userid].canRelay = false;
                users[receiver.userid].receivingFrom = null;
                receiver.socket.emit('rejoin-broadcast', receiver.broadcastId);
            }

        });
    } catch (e) {
        pushLogs(config, 'askNestedUsersToRejoin', e);
    }
}

function getFirstAvailableBroadcaster(broadcastId, maxRelayLimitPerUser) {
    try {
        var broadcastInitiator = users[broadcastId];

        // if initiator is capable to receive users
        if (broadcastInitiator && broadcastInitiator.relayReceivers.length < maxRelayLimitPerUser) {
            return broadcastInitiator;
        }

        // otherwise if initiator knows who is current relaying user
        if (broadcastInitiator && broadcastInitiator.lastRelayuserid) {
            var lastRelayUser = users[broadcastInitiator.lastRelayuserid];
            if (lastRelayUser && lastRelayUser.relayReceivers.length < maxRelayLimitPerUser) {
                return lastRelayUser;
            }
        }

        // otherwise, search for a user who not relayed anything yet
        // todo: why we're using "for-loop" here? it is not safe.
        var userFound;
        for (var n in users) {
            var user = users[n];

            if (userFound) {
                continue;
            } else if (user.broadcastId === broadcastId) {
                // if (!user.relayReceivers.length && user.canRelay === true) {
                if (user.relayReceivers.length < maxRelayLimitPerUser && user.canRelay === true) {
                    userFound = user;
                }
            }
        }

        if (userFound) {
            return userFound;
        }

        // need to increase "maxRelayLimitPerUser" in this situation
        // so that each relaying user can distribute the bandwidth
        return broadcastInitiator;
    } catch (e) {
        pushLogs(config, 'getFirstAvailableBroadcaster', e);
    }
}



