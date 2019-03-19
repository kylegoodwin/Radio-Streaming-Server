const amqp = require('amqplib/callback_api');
const Channel = require('../models/streamModel.js');


exports.sendMessage = function(channelID, type, data, dataName, members) {
    amqp.connect('amqp://rabbit', function(err, conn) {});

    amqp.connect('amqp://rabbit', function(err, conn) {
        conn.createChannel(function(err, ch) {});
    });

    Channel.findOne({channelID:channelID}, function(err, res){
        var messageObj = {
            type: type,
        }

        if(res){
            messageObj.userIDs = res.activeListeners;
        }


   
        messageObj[dataName] = data
   
   
   
         amqp.connect('amqp://rabbit', function(err, conn) {
           conn.createChannel(function(err, ch) {
               var q = 'messages';
   
               ch.assertQueue(q, {durable: true});
               // Note: on Node 6 Buffer.from(msg) should be used
               ch.sendToQueue(q, new Buffer(JSON.stringify(messageObj)));
               console.log(messageObj);
               //setTimeout(function() { conn.close(); process.exit(0) }, 500);
   
               });
           });
    });
     
}

exports.sendStatus = function(channelID, status) {
    amqp.connect('amqp://rabbit', function(err, conn) {});

    amqp.connect('amqp://rabbit', function(err, conn) {
        conn.createChannel(function(err, ch) {});
    });

    Channel.findOne({channelID:channelID}, function(err, res){
        var statusObj = {
            type: "status-update",
        }

        if(res){
            statusObj.userIDs = res.activeListeners;
        }

   
        statusObj.status = status
   
   
   
         amqp.connect('amqp://rabbit', function(err, conn) {
           conn.createChannel(function(err, ch) {
               var q = 'messages';
   
               ch.assertQueue(q, {durable: true});
               // Note: on Node 6 Buffer.from(msg) should be used
               ch.sendToQueue(q, new Buffer(JSON.stringify(statusObj)));
               console.log(statusObj);
               //setTimeout(function() { conn.close(); process.exit(0) }, 500);
   
               });
           });
    });
     
}

//router.use(modifyResponseBody);