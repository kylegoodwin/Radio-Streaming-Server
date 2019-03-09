const Message = require('../models/messageModel.js');
const rabbit = require("./rabbit_handler")

exports.newMessage = function(req, res, channelID, messageCreator){
    console.log(req.body.body)
    var m = Message({
        channelID: channelID,
        body: req.body.body,
        createdAt: Date.now(),
        creator: messageCreator
    })
    m.save(function(err) {
        if (err) {
            res.status = 500;
            res.send("Error saving new message");
            return
        }
        res.status = 201
        res.setHeader('content-type', 'application/json')
        rabbit.sendMessage(channelID, "message-new", m, "message")
        res.send(m)
        return
    })
}

exports.specificMessage = function(req, res, next){
    var urlSplit = req.originalUrl.split("/")
    var messageID = parseInt(urlSplit[urlSplit.length-1], 10)
    var currentUser = parseInt(req.get("X-User"), 10)
    console.log(messageID + " messageID")

    //find message from the url
    if (req.method == "PATCH"){
        Message.findOneAndUpdate({messageid:messageID}, {$set: {body:req.body.body}}, {new:true}, function(err, message){
            if (err) {
                res.status = 400;
                res.send("Error finding message");
                return
            }
            if(currentUser != message.creator){
                res.status(403);
                res.send("Forbidden");
                return
            }
            res.setHeader('content-type', 'application/json')
            rabbit.sendMessage(message.channelID, "message-update", message, "message")
            res.send(message)
            return
        })
    } else if(req.method == "DELETE"){
        Message.findOneAndDelete({messageid:messageID}, function(err, message){
            if (err) {
                res.status = 400;
                res.send("Error finding message");
                return
            }
            if(currentUser != message.creator){
                res.status(403);
                res.send("Forbidden");
                return
            }
            rabbit.sendMessage(message.channelID, "message-delete", messageID, "messageID")
            res.send("Message successfully deleted")
        })
    }
}

//Simple version, without validation or sanitation
exports.test = function (req, res) {
    res.send('Greetings from the message controller!');
};