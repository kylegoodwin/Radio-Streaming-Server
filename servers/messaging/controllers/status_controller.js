const Channel = require('../models/streamModel.js');
const rabbit = require('./rabbit_handler')

exports.updateStatus = function(req, res){
    var currentUser = JSON.parse(req.header("X-User")).id;
    var channelID = req.params.channelID


    if( req.body.text && req.body.image ){

        let newStatus = {text: req.body.text, photoURL: req.body.image};
        console.log("xusid " + currentUser)
        Channel.findOneAndUpdate({ 'creator.id': currentUser, channelID: channelID }, {status: newStatus}, {new: true}, function(err,response){
            if( err || !response.status.text){
                if( err){
                    console.log("error " + error)
                }
                res.status = 403;
                res.send("Error updating status. You must own the channel to update its status.")
            }else{
                rabbit.sendStatus(channelID,newStatus)
                res.status = 200;
                res.send("Status updated!")
            }
        });


    }else{
        res.status = 400;
        res.send("Invalid Request, status update must have text and URL link to image");
        
    }


    
    
    /*
    var currentUser = JSON.parse(req.header("X-User"));
    var m = Message({
        channelID: channelID,
        body: req.body.body,
        createdAt: Date.now(),
        creator: currentUser
    })
    m.save(function(err) {
        if (err) {
            res.status = 500;
            res.send("Error saving new message " + err);
            return
        }
        res.status = 201
        res.setHeader('content-type', 'application/json')
        rabbit.sendMessage(channelID, "message-new", m, "message")
        res.send(m)
        return
    })
    */
}