const amqp = require('amqplib/callback_api');
const Channel = require('../models/streamModel.js');


exports.sendMessage = function(channelID, type, data, dataName, members) {
    amqp.connect('amqp://rabbit', function(err, conn) {});

    amqp.connect('amqp://rabbit', function(err, conn) {
        conn.createChannel(function(err, ch) {});
    });

    Channel.findOne({id:channelID}, function(err, res){
        var messageObj = {
            type: type,
        }

        // if(!res.listeners){
        //     messageObj.userIDs = members
        // } else {
            messageObj.userIDs = res.listeners
        //}

   
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

//router.use(modifyResponseBody);