const http = require("http");
const amqp = require('amqplib/callback_api');
const bodyParser = require('body-parser');
const express = require('express');
var mongoose = require('mongoose');
var Channel = require('./models/channel.js');
var Message = require('./models/message.js');
const app = express();

//Get environment variables
let PORT = "";
if(!process.env.PORT){
    PORT = "80";
}else{
    PORT = process.env.PORT;
}

/*
const dbIP = process.env.DBIP;
if(isNull(dbIP)){
    dbIP = "localhost:1234";
}*/
//var dbIP = "localhost:27017";


//Connect to MongoDB
    //Make sure to create docker container for mongo instance with:
    //sudo docker run -d -p 27017:27017 mongo
    
// Retrieve mongoconnection
var MongoClient = require('mongodb').MongoClient;
//Creates the collections if they don't already exist

//old mongo connect address: "mongodb://localhost:27017/"
MongoClient.connect("mongodb://mongo:27017", function(err,dbconn){
    //Get the db from the connection
    var db = dbconn.db("messagingDB");
    db.createCollection('Channels', function(err, collection) {});
    db.createCollection('Messages', function(err, collection) {});
});

//Let mongoose connect too
mongoose.connect("mongodb://mongo:27017");

//Connect to RabbitMQ
let messagingChannel;
//var rabbit
amqp.connect('amqp://guest:guest@rabbitmq:5672', function(err, conn) {
    if (err) {
		console.log("Failed to connect to Rabbit instance from messaging.");
		process.exit(1);
	}
    conn.createChannel(function(err,ch){
        if (err) {
            console.log("Failed to create channel in messaging.");
            process.exit(1);
        }
        messagingChannel = ch;
        messagingChannel.assertQueue("MessagingQ", {durable: false});
    });
    //rabbit = conn;
});

function parseMemberIDs(channel) {
    //Collect the IDs of members in the selected channel
    var memberIDs = [];
    channel.members.forEach(member => {
        memberIDs = memberIDs.concat(member.id);
    });
    return memberIDs;
}

//Setting up the app tools
app.use(bodyParser.json());

//Channel Creation
app.post("/v1/channels", (req,res) => {
        //415
        if(req.header("Content-Type") == "application/json"){
            //401
            if(req.header("X-User") != null ){
                //Fetch the JSON body
                var givenChannel = req.body;   

                //Make sure that the name is on the channel (Minimum requirement for a channel) (400)
                if(givenChannel.name){
                    //Get the user sending the request
                    var currentUser = JSON.parse(req.header("X-User"));

                    console.log(currentUser.id);
                    //Put the creator on the new channel
                    givenChannel.creator = currentUser;

                    //Add the creator as the first memeber
                    givenChannel.members = [currentUser];

                    //Create an instance of a channel
                    var newChannel = new Channel({
                        name: givenChannel.name,
                        description: givenChannel.description,
                        private: givenChannel.private,
                        members: givenChannel.members,
                        creator: givenChannel.creator

                    });

                    //Attempt mongodb insert
                    newChannel.editedAt = Date.now();
                    newChannel.save(err => {
                        if (err){
                            if(err == "Error: Duplicate Channel Creation"){
                                
                                res.status(409).send("Channel of name "+newChannel.name+" already exists.");
                                return;
                            }
                            res.status(500).send("Error with DB during channel creation: "+err);
                            return;
                        } 
                        Channel.findOne({name: newChannel.name}, (err, found) =>{
                            if (err) throw err;

                            //Rabbit event creation
                            var channelCreateEvent;
                            if(found.private){
                                memberIDs = parseMemberIDs(found)
                                channelCreateEvent = {
                                    "type": "channel-new",
                                    "channel":found,
                                    "userIDs": memberIDs
                                };
                                console.log(channelCreateEvent)
                            }else{
                                channelCreateEvent = {
                                    "type": "channel-new",
                                    "channel":found
                                };
                            }
                            //I think this is how you send an event to a queue?
                            channelCreateEvent = JSON.stringify(channelCreateEvent);
                            messagingChannel.sendToQueue("MessagingQ", new Buffer(channelCreateEvent));

                            //This call sets the status and stringifies the json and sets the content header
                            res.status(201).json(found);
                            return;
                        });

                    });
                }else{
                    res.writeHead(400)
                    res.end()
                    return;
                }
            }else{
                res.writeHead(401);
                res.end();
                return;
            }
        }else{
            res.writeHead(415);
            res.end();
            return;
        }
    }
);

//Get all channels the current user has access to.
app.get("/v1/channels", (req,res) => {
    //401
    if(req.header("X-User") != null ){
        //Get the user sending the request
        var currentUser = JSON.parse(req.header("X-User"));
        var foundAll; 
        //Find both the non-private channels, and private channels where the user is a member
        Channel.find({private: false}, (err, found1) =>{
            if (err){
                res.status(500).send("Error with DB during channel fetching: "+err);
                return;
            }
            
            //second call to make sure it inclues all channels where the user is a memeber
            Channel.find({private: true, members:{$elemMatch: currentUser}}, (err, found2) =>{ // cant believe i found that flag on first google
                if (err){
                    res.status(500).send("Error with DB during channel fetching: "+err);
                    return;
                }
                foundAll = found1.concat(found2);
                
                //This call sets the status and stringifies the json and sets the content header
                res.status(200).json(foundAll);
                return;
            });                   
        });
    }else{
        res.writeHead(401);
        res.end();
        return;
    }
});

//Post a message to a specific channel as long as the the current user has access.
app.post("/v1/channels/:channelID", (req,res) => {
    //415
    if(req.header("Content-Type") == "application/json"){
        //401
        if(req.header("X-User") != null ){
            //Fetch the JSON body
            var givenMessage = req.body;
            //Get the user sending the request
            var currentUser = JSON.parse(req.header("X-User"));

            //Find the requested channel and check that they have access
            Channel.findOne({_id: req.params.channelID}, (err, requestedChannel) =>{
                if (err){
                    res.status(500).send("Error with DB during channel fetching: "+err);
                    return;
                }

                //Make sure that it was found
                if(!requestedChannel){
                    res.status(404).send("Channel with ID "+req.params.channelID+" was not found");
                    return;
                }
            
                // //Collect the IDs of members in the selected channel
                // var memberIDs = [];
                // requestedChannel.members.forEach(member => {
                //     memberIDs = memberIDs.concat(member.id);
                // });

                memberIDs = parseMemberIDs(requestedChannel);

                //verify that the current user's ID is in the private channel's members
                if(requestedChannel.private && !memberIDs.includes(currentUser.id)){
                    res.status(403).send("This channel is private and you are not a member");
                    return;
                }

                //Message creation
                var newMessage = new Message({
                    channelID: req.params.channelID,
                    body: givenMessage.body,
                    createdAt: Date.now(),
                    creator: currentUser,
                    editedAt: Date.now()
                });

                //Perform the save to DB
                newMessage.save((err, msg) =>{
                    if (err){
                        res.status(500).send("Error with DB during message creation: "+err);
                        return;
                    }

                    //Rabbit event creation
                    var messageCreateEvent;
                    if(requestedChannel.private){
                        messageCreateEvent = {
                            "type": "message-new",
                            "message":msg,
                            "userIDs": memberIDs
                        };
                    }else{
                        messageCreateEvent = {
                            "type": "message-new",
                            "message":msg
                        };
                    }
                    console.log(messageCreateEvent);

                    //I think this is how you send an event to an exchange?
                    messageCreateEvent = JSON.stringify(messageCreateEvent);
                    messagingChannel.sendToQueue("MessagingQ", new Buffer(messageCreateEvent));

                    //respond to client
                    res.status(201).json(msg);
                    return
                });                                       
            });
        }else{
            res.writeHead(401);
            res.end();
            return;
        }
    }else{
        res.writeHead(415);
        res.end();
        return;
    }
});

//return last 100 messages in this channel
app.get("/v1/channels/:channelID", (req,res) => {
    //401
    if(req.header("X-User") != null ){
        //Get the user sending the request
        var currentUser = JSON.parse(req.header("X-User"));

        //Find the requested channel and check that they have access
        Channel.findOne({_id: req.params.channelID}, (err, requestedChannel) =>{
            if (err){
                res.status(500).send("Error with DB during channel fetching: "+err);
                return;
            }

            //Make sure that it was found
            if(!requestedChannel){
                res.status(404).send("Channel with ID "+req.params.channelID+" was not found");
                return;
            }
        
            //Collect the IDs of members in the selected channel
            var memberIDs = [];
            requestedChannel.members.forEach(member => {
                memberIDs = memberIDs.concat(member.id);
            });

            //verify that the current user's ID is in the private channel's members
            if(requestedChannel.private && !memberIDs.includes(currentUser.id)){
                res.status(403).send("This channel is private and you are not a member");
                return;
            }

            //scrolling back support
            if(req.query.before){
                var msgID = req.query.before;
                //Retrieve most recent 100 messages before the given message ID from the requested channel
                Message.find({channelID: req.params.channelID, _id: {$lt: msgID}},(err, msg) =>{
                    if (err){
                        res.status(500).send("Error with DB during message creation: "+err);
                        return;
                    }
                    res.status(200).json(msg);
                    return;
                }).sort({createdAt: -1}).limit(100);
            }else{
                //Perform the standard messages request
                //Retrieve most recent 100 messages from the requested channel
                Message.find({channelID: req.params.channelID},(err, msg) =>{
                    if (err){
                        res.status(500).send("Error with DB during message creation: "+err);
                        return;
                    }
                    res.status(200).json(msg);
                    return;
                }).sort({createdAt: -1}).limit(100);
            }

        });
    }else{
        res.writeHead(401);
        res.end();
        return;
    }
});

//Edit channel name or description
app.patch('/v1/channels/:channelID', (req,res) => {
    //415
    if(req.header("Content-Type") == "application/json"){
        //401
        if(req.header("X-User") != null ){
            //Fetch the JSON body
            var channelUpdates = req.body;

            //Get the user sending the request
            var currentUser = JSON.parse(req.header("X-User"));
            
            //Verify the request
            if (!channelUpdates.name && !channelUpdates.description){
                res.status(400).send("Channel updates needs a name, description, or both.");
                return;
            }

            //Create an updates object based on what is present, disregardng updates that arent name or description
            var updates = {};
            if(channelUpdates.name){
                updates.name = channelUpdates.name;
            }
            if(channelUpdates.description){
                updates.description = channelUpdates.description;
            }

            //Find the requested channel and check that they have access
            Channel.findOne({_id: req.params.channelID}, (err, requestedChannel) =>{
                if (err){
                    res.status(500).send("Error with DB during channel fetching: "+err);
                    return;
                }

                //Make sure that it was found
                if(!requestedChannel){
                    res.status(404).send("Channel with ID "+req.params.channelID+" was not found");
                    return;
                }

                //verify that the current user's ID is the private channel's creator
                if(!(requestedChannel.creator.id == currentUser.id)){
                    res.status(403).send("You are not the creator of this Channel");
                    return;
                }

                //new option returns the modified document
                Channel.findOneAndUpdate({_id:req.params.channelID}, updates, {new:true}, (err, doc) =>{
                    if (err){
                        res.status(500).send("Error with DB during channel modification: "+err);
                        return;
                    }

                    //Rabbit event creation
                    var channelUpdateEvent;
                    if(doc.private){
                        memberIDs = parseMemberIDs(doc)
                        channelUpdateEvent = {
                            "type": "channel-update",
                            "channel":doc,
                            "userIDs": memberIDs
                        };
                    }else{
                        channelUpdateEvent = {
                            "type": "channel-update",
                            "channel":doc
                        };
                    }
                    console.log(channelUpdateEvent);
                    //I think this is how you send an event to an exchange?
                    channelUpdateEvent = JSON.stringify(channelUpdateEvent);
                    messagingChannel.sendToQueue("MessagingQ", new Buffer(channelUpdateEvent));

                    res.status(200).json(doc);
                    return;
                });
                
                
            });
        }else{
            res.writeHead(401);
            res.end();
            return;
        }
    }else{
        res.writeHead(415);
        res.end();
        return;
    }
});

//Delete channel and its messages
app.delete("/v1/channels/:channelID", (req,res) => {
    //401
    if(req.header("X-User") != null ){
        //Get the user sending the request
        var currentUser = JSON.parse(req.header("X-User"));

        //Find the requested channel and check that they have access
        Channel.findOne({_id: req.params.channelID}, (err, requestedChannel) =>{
            if (err){
                res.status(500).send("Error with DB during channel fetching: "+err);
                return;
            }
            
            //Make sure that it was found
            if(!requestedChannel){
                res.status(404).send("Channel with ID "+req.params.channelID+" was not found");
                return;
            }

            //verify that the current user's ID is the private channel's creator
            if(!(requestedChannel.creator.id == currentUser.id)){
                res.status(403).send("You are not the creator of this Channel");
                return;
            }

            //Begin deletion
            Channel.findOneAndDelete({_id: req.params.channelID}, (err,result) => {
                if (err){
                    res.status(500).send("Error with DB during channel fetching: "+err);
                    return;
                }
                console.log("channel deletion results: "+result);
                
                //Delete associated messages
                Message.deleteMany({channelID:req.params.channelID}, (err, result2) => {
                    if (err){
                        res.status(500).send("Error with DB during channel fetching: "+err);
                        return;
                    }
                    console.log("Message deletion results: "+result2);

                    //Rabbit event creation
                    var channelDeleteEvent;
                    if(requestedChannel.private){
                        memberIDs = parseMemberIDs(requestedChannel)
                        channelDeleteEvent = {
                            "type": "channel-delete",
                            "channelID":requestedChannel._id,
                            "userIDs": memberIDs
                        };
                    }else{
                        channelDeleteEvent = {
                            "type": "channel-delete",
                            "channelID":requestedChannel._id
                        };
                    }
                    console.log(channelDeleteEvent);
                    //I think this is how you send an event to an exchange?
                    channelDeleteEvent = JSON.stringify(channelDeleteEvent);
                    messagingChannel.sendToQueue("MessagingQ", new Buffer(channelDeleteEvent));

                    res.status(200).send("Deleted channel of ID "+req.params.channelID+" and all of its messages");
                    return;
                });
            });
        });
    }else{
        res.writeHead(401);
        res.end();
        return;
    }

});

//Add member to private channel
app.post("/v1/channels/:channelID/members", (req,res) => {
    //415 I'm going to code my client to always send an object instead of fetching from user store within this app
    if(req.header("Content-Type") == "application/json"){
        //401
        if(req.header("X-User") != null ){
            //Fetch the JSON body
            var userToAdd = req.body;
            //Get the user sending the request
            var currentUser = JSON.parse(req.header("X-User"));

            //Find the requested channel and check that they have access
            Channel.findOne({_id: req.params.channelID}, (err, requestedChannel) =>{
                if (err){
                    res.status(500).send("Error with DB during channel fetching: "+err);
                    return;
                }
                
                //Make sure that it was found
                if(!requestedChannel){
                    res.status(404).send("Channel with ID "+req.params.channelID+" was not found");
                    return;
                }

                if(!requestedChannel.private){
                    res.status(400).send("This channel is not private");
                    return;
                }

                //verify that the current user's ID is the private channel's creator
                if(!(requestedChannel.creator.id == currentUser.id)){
                    res.status(403).send("You are not the creator of this Channel and therefore you do not have the authority to add new members to this private channel");
                    return;
                }

                //Add the sent user to the channel's members
                Channel.updateOne({_id: req.params.channelID}, {$push: {members: userToAdd}}, err => {
                    if (err){
                        res.status(500).send("Error with DB during member adding: "+err);
                        return;
                    }
                    res.status(200).send("User was added to this channel");
                    return;
                });
            });
        }else{
            res.writeHead(401);
            res.end();
            return;
        }
    }else{
        res.writeHead(415);
        res.end();
        return;
    }

});

//Remove a member from a private channel
app.delete("/v1/channels/:channelID/members", (req,res) => {
    //415 I'm going to code my client to always send an object instead of fetching from user store within this app
    if(req.header("Content-Type") == "application/json"){
        //401
        if(req.header("X-User") != null ){
            //Fetch the JSON body
            var userToRemove = req.body;
            //Get the user sending the request
            var currentUser = JSON.parse(req.header("X-User"));

            //Find the requested channel and check that they have access
            Channel.findOne({_id: req.params.channelID}, (err, requestedChannel) =>{
                if (err){
                    res.status(500).send("Error with DB during channel fetching: "+err);
                    return;
                }
                
                //Make sure that it was found
                if(!requestedChannel){
                    res.status(404).send("Channel with ID "+req.params.channelID+" was not found");
                    return;
                }

                if(!requestedChannel.private){
                    res.status(400).send("This channel is not private");
                    return;
                }

                //verify that the current user's ID is the private channel's creator
                if(!(requestedChannel.creator.id == currentUser.id)){
                    res.status(403).send("You are not the creator of this Channel and therefore you do not have the authority to remove members from this private channel");
                    return;
                }

                //Add the sent user to the channel's members
                Channel.updateOne({_id: req.params.channelID}, {$pull: {members: userToRemove}}, err => {
                    if (err){
                        res.status(500).send("Error with DB during member adding: "+err);
                        return;
                    }
                    res.status(200).send("User was removed from this channel");
                    return;
                });
            });
        }else{
            res.writeHead(401);
            res.end();
            return;
        }
    }else{
        res.writeHead(415);
        res.end();
        return;
    }

});

//Edit message body
app.patch('/v1/messages/:messageID', (req,res) => {
    //415
    if(req.header("Content-Type") == "application/json"){
        //401
        if(req.header("X-User") != null ){
            //Fetch the JSON body
            var messageUpdates = req.body;

            //Get the user sending the request
            var currentUser = JSON.parse(req.header("X-User"));
            
            //Verify the request
            if (!messageUpdates.body){
                res.status(400).send("Message updates requires a body.");
                return;
            }

            //Create an updates object populated with only the given body, disregardng updates that arent name or description
            var updates = {body: messageUpdates.body};

            //Find the requested message and check that they have access
            Message.findOne({_id: req.params.messageID}, (err, requestedMessage) =>{
                if (err){
                    res.status(500).send("Error with DB during message fetching: "+err);
                    return;
                }

                //Make sure that it was found
                if(!requestedMessage){
                    res.status(404).send("Message with ID "+req.params.messageID+" was not found");
                    return;
                }

                //verify that the current user's ID is the private channel's creator
                if(!(requestedMessage.creator.id == currentUser.id)){
                    res.status(403).send("You are not the creator of this message");
                    return;
                }

                //new option returns the modified document
                Message.findOneAndUpdate({_id:req.params.messageID}, updates, {new:true}, (err, doc) =>{
                    if (err){
                        res.status(500).send("Error with DB during message modification: "+err);
                        return;
                    }

                    Channel.findOne({_id:doc.channelID}, (err, found) =>{
                        if (err){
                            res.status(500).send("Error with DB during channel fetching: "+err);
                            return;
                        }

                        //Rabbit event creation
                        var messageUpdateEvent;
                        if(found.private){
                            memberIDs = parseMemberIDs(found);
                            messageUpdateEvent = {
                                "type": "message-update",
                                "message":msg,
                                "userIDs": memberIDs
                            };
                        }else{
                            messageUpdateEvent = {
                                "type": "message-update",
                                "message":msg
                            };
                        }
                        console.log(messageUpdateEvent);
                        //I think this is how you send an event to an exchange?
                        messageUpdateEvent = JSON.stringify(messageUpdateEvent);
                        messagingChannel.sendToQueue("MessagingQ", new Buffer(messageUpdateEvent));

                        //respond to client
                        res.status(200).json(doc);
                        return;
                    });
                });
                
                
            });
        }else{
            res.writeHead(401);
            res.end();
            return;
        }
    }else{
        res.writeHead(415);
        res.end();
        return;
    }
});

//Delete Message
//Delete channel and its messages
app.delete("/v1/messages/:messageID", (req,res) => {
    //401
    if(req.header("X-User") != null ){
        //Get the user sending the request
        var currentUser = JSON.parse(req.header("X-User"));

        //Find the requested channel and check that they have access
        Message.findOne({_id: req.params.messageID}, (err, requestedMessage) =>{
            if (err){
                res.status(500).send("Error with DB during message fetching: "+err);
                return;
            }
            
            //Make sure that it was found
            if(!requestedMessage){
                res.status(404).send("Message with ID "+req.params.messageID+" was not found");
                return;
            }

            //verify that the current user's ID is the private channel's creator
            if(!(requestedMessage.creator.id == currentUser.id)){
                res.status(403).send("You are not the creator of this message");
                return;
            }

            //Begin deletion
            Message.findOneAndDelete({_id: req.params.messageID}, (err,result) => {
                if (err){
                    res.status(500).send("Error with DB during message fetching: "+err);
                    return;
                }
                Channel.findOne({_id:result.channelID}, (err, found) =>{
                    if (err){
                        res.status(500).send("Error with DB during channel fetching: "+err);
                        return;
                    }

                    //Rabbit event creation
                    var messageDeleteEvent;
                    if(found.private){
                        memberIDs = parseMemberIDs(found);
                        messageDeleteEvent = {
                            "type": "message-delete",
                            "messageID":result._id,
                            "userIDs": memberIDs
                        };
                    }else{
                        messageDeleteEvent = {
                            "type": "message-delete",
                            "messageID":result._id
                        };
                    }
                    console.log(messageDeleteEvent);
                    //I think this is how you send an event to an exchange?
                    messageDeleteEvent = JSON.stringify(messageDeleteEvent);
                    messagingChannel.sendToQueue("MessagingQ", new Buffer(messageDeleteEvent));

                    console.log("Message deletion results: "+result);

                    res.status(200).send("Deleted message of ID "+req.params.messageID);
                    return;
                });
            });
        });
    }else{
        res.writeHead(401);
        res.end();
        return;
    }

});

app.listen( PORT, () => console.log("Server running and listening on port " + PORT));
