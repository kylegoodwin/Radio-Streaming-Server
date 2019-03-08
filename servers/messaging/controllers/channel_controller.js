const Channel = require('../models/channelModel.js');
const Message = require('../models/messageModel.js');
const message_controller = require('./message_controller.js')
const rabbit = require('./rabbit_handler')
const express = require('express');

 exports.handleGeneral = function() {
    Channel.find({name:"general"}, function(err, data) {
        if (err) {
            res.status = 500;
            res.send("Channel error");
            return
        }
        if (data.length != 0) {
            return
        } else {
            var c = Channel({
                name: "general",
                description: "Channel for general discussion",
                private: false,
                createdAt: Date.now()
            })
            c.save(function(err) {
                if (err) {
                    res.status = 500;
                    res.send("Error saving channel");
                    return
                }
                return
            })
        }
    })
}

exports.allChannels = function (req, res) {
    if (req.method == "GET"){
        Channel.find({}, function (err, channels) {
            if (err) {
                res.status = 400;
                res.send("Error finding channel");
                return
            }
            res.json(channels);
        });
    } else if (req.method == "POST"){

        //get user and add to members
        var channelCreator = req.get("X-User")
        var newChannelMembers = [channelCreator]

        var desc = ""
        if (req.body.description) {
            desc = req.body.description;
        }

        //create new channel
        var newChannel = Channel({
            name: req.body.name,
            description: desc,
            private: req.body.private,
            creator: channelCreator,
            members: newChannelMembers,
            createdAt: Date.now()
        })
        newChannel.save(function(err) {
            if (err) {
                res.status = 500
                res.send("Error saving channel")
            }
            res.status = 201
            res.setHeader('Content-Type', 'application/json')
            rabbit.sendMessage(newChannel.id, "channel-new", newChannel, "channel", newChannel.members)
            res.send(newChannel)
            return
        })
    };
}

exports.specificChannel = function (req, res, next) {
        console.log("specificChannels")
    console.log(res)
    var urlSplit = req.originalUrl.split("/")
    var channelID = parseInt(urlSplit[urlSplit.length-1], 10)
    var currentUser = parseInt(req.get("X-User"), 10)
    console.log(channelID)

    //find channel from the url
    Channel.findOne({id:channelID}, function(err, channel){
        if (err) {
            res.status = 400;
            res.send("Error: channel not found");
            return
        }
        var members = channel.members

        //check if the current user is a member
        if(members.includes(currentUser) == false && channel.private == true){
            res.status(403)
            res.send("Forbidden")
            return

        } else if (req.method == "GET"){
            //check for before query parameter
            var beforeMessageID = req.query.before
            if(beforeMessageID){
                Message.find({channelID:channelID, messageid: {$lt: beforeMessageID}})
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
                Message.find({channelID:channelID})
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
            message_controller.newMessage(req, res, channelID, currentUser)
            return

        } else if (req.method == "PATCH"){

            //check if user is channel creator
            if(currentUser != channel.creator){
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
                
                //find by channelID and update
                Channel.findOneAndUpdate({id:channelID}, update, {new:true}, function(err, doc){
                    if (err) {
                        res.status = 400
                        res.send("Channel not found")
                    }
                    res.setHeader('Content-Type', 'application/json');
                    rabbit.sendMessage(channelID, "channel-update", channel, "channel", channel.members)
                    res.send(doc);
                    return

                });
            };

            return

        } else if(req.method == "DELETE"){

            //check if user is channel creator
            if(currentUser != channel.creator){
                res.status(403);
                res.send("Forbidden");
                return

            } else {
                
                //delete channel and messages
                Channel.findOneAndDelete({id:channelID}, function(err, doc){
                    if (err) {
                        res.status = 400
                        res.send("Channel not found")
                    }
                    Message.deleteMany({channelID:channelID}, function(err){
                        if (err){
                            res.status = 400
                            res.send("Error deleting channe;")
                        }
                        rabbit.sendMessage(channelID, "channel-delete", channelID, "channelID")
                        res.send("Channel succesfully deleted");
                        next();
                        return
                    })
                })
            }
            return
        }
    }); 
}

exports.channelMembers = function(req, res, next){
    var urlSplit = req.originalUrl.split("/")
    var channelID = parseInt(urlSplit[urlSplit.length-2], 10)
    console.log(channelID)
    var currentUser = parseInt(req.get("X-User"), 10)

    //find channel from the url
    Channel.findOne({id:channelID}, function(err, channel){
        if(currentUser != channel.creator){
            res.status(403);
            res.send("Forbidden");
            return
        } else if(req.method == "POST"){
            Channel.findOneAndUpdate({id:channelID}, { $push: {members:req.body.user} }, {new:true}, function(err, doc){
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
            var newMembers = channel.members
            var index = newMembers.indexOf(req.body.user);
            if (index > -1) {
                newMembers.splice(req.body.user, 1);
            }
            Channel.findOneAndUpdate({id:channelID}, ({members:newMembers}), {new:true}, function(err, doc){
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