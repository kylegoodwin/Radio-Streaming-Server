const Channel = require('../models/streamModel.js');
const Message = require('../models/messageModel.js');
const message_controller = require('./message_controller.js')
const rabbit = require('./rabbit_handler')
const express = require('express');


exports.specificChannel = function (req, res, next) {
    var channelID = req.params.channelID;

    //find channel from the url
    Channel.findOne({channelID: channelID}, function(err, channel){
        if (err) {
            res.status = 400;
            res.send("Error: channel not found");
            return
        }
        var members = [];
        
        if (channel){
            members = channel.activeListeners;
        }else{
            res.status = 400;
            res.send("Error: channel not found");
            return
        }

        //check if the current user is a member
        /*
        if(listeners.includes(currentUser) == false){
            res.status(403)
            res.send("Forbidden")
            return
        */

        if (req.method == "GET"){
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
            message_controller.newMessage(req, res)
            return
        } 
    }); 
}
