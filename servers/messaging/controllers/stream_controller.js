const Stream = require('../models/streamModel.js');
const Message = require('../models/messageModel.js');
const message_controller = require('./message_controller.js')
const rabbit = require('./rabbit_handler')
const express = require('express');

//do not need a general stream in this program
//  exports.handleGeneral = function() {
//     Stream.find({name:"general"}, function(err, data) {
//         if (err) {
//             res.status = 500;
//             res.send("Stream error");
//             return
//         }
//         if (data.length != 0) {
//             return
//         } else {
//             var c = Stream({
//                 name: "general",
//                 description: "Stream for general discussion",
//                 private: false,
//                 createdAt: Date.now()
//             })
//             c.save(function(err) {
//                 if (err) {
//                     res.status = 500;
//                     res.send("Error saving stream");
//                     return
//                 }
//                 return
//             })
//         }
//     })
// }

// --> /streams
exports.allStreams = function (req, res) {
    if (req.method == "GET"){
        Stream.find({}, function (err, streams) {
            if (err) {
                res.status = 400;
                res.send("Error finding stream");
                return
            }
            res.json(streams);
        });
    } else if (req.method == "POST"){ 

        //get user and add to members
        var streamCreator = req.get("X-User")

        var desc = ""
        if (req.body.description) {
            desc = req.body.description;
        }

        //create new stream
        var newStream = Stream({
            name: req.body.name,
            description: desc,
            private: req.body.private,
            creator: streamCreator,
            listeners: [streamCreator.ID],
            createdAt: Date.now()
        })
        newStream.save(function(err) {
            if (err) {
                res.status = 500
                res.send("Error saving stream")
            }
            res.status = 201
            res.setHeader('Content-Type', 'application/json')
            rabbit.sendMessage(newStream.id, "stream-new", newStream, "stream", newStream.members)
            res.send(newStream)
            return
        })
    };
}

exports.specificStream = function (req, res, next) {
    console.log("specificStreams")
    console.log(res)
    var urlSplit = req.originalUrl.split("/")
    var streamID = parseInt(urlSplit[urlSplit.length-1], 10)
    var currentUser = parseInt(req.get("X-User"), 10)
    console.log(streamID)

    //find stream from the url
    Stream.findOne({id:streamID}, function(err, stream){
        if (err) {
            res.status = 400;
            res.send("Error: stream not found");
            return
        }
        var members = stream.members

        //check if the current user is a member
        if(members.includes(currentUser) == false && stream.private == true){
            res.status(403)
            res.send("Forbidden")
            return

        } else if (req.method == "GET"){
            //check for before query parameter
            var beforeMessageID = req.query.before
            if(beforeMessageID){
                Message.find({streamID:streamID, messageid: {$lt: beforeMessageID}})
                .sort({createdAt: -1})
                .limit(100)
                .exec(function(err, response){
                    if (err) {
                        res.status = 400;
                        res.send("Error finding message");
                        return
                    }
                    res.status = 201;
                    res.setHeader('Content-Type', 'application/json');
                    res.send(response);
                    return
                })
            } else {
                //get the first 100 messages
                Message.find({streamID:streamID})
                .sort({createdAt: -1})
                .limit(100)
                .exec(function(err, response){
                    if (err) {
                        res.status = 400;
                        res.send("Error finding message");
                        return
                    }
                    res.status = 201;
                    res.setHeader('Content-Type', 'application/json');
                    res.send(response);
                    return
                })
            }

        } else if (req.method == "POST"){

            //post a new message
            message_controller.newMessage(req, res, streamID, currentUser)
            return

        } else if (req.method == "PATCH"){

            //check if user is stream creator
            if(currentUser != stream.creator){
                res.status(403);
                res.send("Forbidden");
                return

            } else {

                //check for name a description 
                var update = {};
                if (req.body.name.length != 0){
                    update.name = req.body.name;
                };
                if(req.body.description.length != 0){
                    update.description = req.body.description;
                };
                
                //find by streamID and update
                Stream.findOneAndUpdate({id:streamID}, update, {new:true}, function(err, doc){
                    if (err) {
                        res.status = 400
                        res.send("Stream not found")
                    }
                    res.setHeader('Content-Type', 'application/json');
                    rabbit.sendMessage(streamID, "stream-update", stream, "stream", stream.members)
                    res.send(doc);
                    return

                });
            };

            return

        } else if(req.method == "DELETE"){

            //check if user is stream creator
            if(currentUser != stream.creator){
                res.status(403);
                res.send("Forbidden");
                return

            } else {
                
                //delete stream and messages
                Stream.findOneAndDelete({id:streamID}, function(err, doc){
                    if (err) {
                        res.status = 400
                        res.send("Stream not found")
                    }
                    Message.deleteMany({streamID:streamID}, function(err){
                        if (err){
                            res.status = 400
                            res.send("Error deleting channe;")
                        }
                        rabbit.sendMessage(streamID, "stream-delete", streamID, "streamID")
                        res.send("Stream succesfully deleted");
                        next();
                        return
                    })
                })
            }
            return
        }
    }); 
}

exports.streamMembers = function(req, res, next){
    var urlSplit = req.originalUrl.split("/")
    var streamID = parseInt(urlSplit[urlSplit.length-2], 10)
    console.log(streamID)
    var currentUser = parseInt(req.get("X-User"), 10)

    //find stream from the url
    Stream.findOne({id:streamID}, function(err, stream){
        if(currentUser != stream.creator){
            res.status(403);
            res.send("Forbidden");
            return
        } else if(req.method == "POST"){
            Stream.findOneAndUpdate({id:streamID}, { $push: {members:req.body.user} }, {new:true}, function(err, doc){
                if (err) {
                    res.status = 400
                    res.send("Error adding new user")
                    return
                }
                res.status = 201
                res.send("User successfully added");
                return
            });
        } else if(req.method == "DELETE"){
            //TODO: check if user is already a member
            //reafctor to make these the same else statement
            var newMembers = stream.members
            var index = newMembers.indexOf(req.body.user);
            if (index > -1) {
                newMembers.splice(req.body.user, 1);
            }
            Stream.findOneAndUpdate({id:streamID}, ({members:newMembers}), {new:true}, function(err, doc){
                if (err) {
                    res.status = 400
                    res.send("Error deleting user")
                }
                res.status = 200
                res.send("User successfully deleted");
                return
            });
        } else {
            res.status = 405
            res.send("Method Not Allowed")
        }
    });
}