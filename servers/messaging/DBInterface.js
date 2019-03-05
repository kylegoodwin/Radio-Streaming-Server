
//require('mongodb');

//Constructor?
var DBInterface = function(dbClient){
    DBInterface.client = dbClient;
    // Connect to the db
    DBInterface.client.connect("mongodb://localhost:27017/", function(err, dbconn) {    
        if(err) { return console.dir(err); }

        //Get the db from the connection
        var db = dbconn.db("messagingDB");

        //Creates the collections if they don't already exist
        db.createCollection('channels', function(err, collection) {});
        db.createCollection('messages', function(err, collection) {});

        //Make sure general channel is in the db
        var generalChannel = {
            "name":"general",
            "description":"The default channel for messages",
            "private":false,
            "members":[],
            "creator":{
                "username":"admin",
                "firstname":"admin",
                "lastname":"admin"
            }
        };
        DBInterface.createChannel(generalChannel);
    });
}

//Fetcher for the client instance
DBInterface.getClient = function(){
    return DBInterface.client;
}

//Function to create a channel in the db
DBInterface.createChannel = function(givenChannel){
    this.getClient().connect("mongodb://localhost:27017/", function(err, dbconn){
        if(err) { return console.dir(err); }

        //Get the db from the connection
        var db = dbconn.db("messagingDB");
        
        //Collection to reference
        var collection = db.collection('channels');
        
        //Create the new channel doc to be inserted
        var channel = {"name":givenChannel.name,
         "description":givenChannel.description,
          "private":givenChannel.private,
           "members":givenChannel.members,
           "createdAt": DBInterface.getCurrentDateTime(),
            "creator":givenChannel.creator,
            "editedAt":""
        };

        //attempt nsert if it doesnt exist
        collection.insertOne(channel, {w:1}, function(err, result) {
            if(givenChannel.name != "general"){
                if(err) { return console.dir(err); }
            }
            return result;
        });
            
    });
    
    return null;
}

DBInterface.getChannel = function(channelName){
    this.client.connect("mongodb://localhost:27017/", function(err, dbconn){
        if(err) { return console.dir(err); }
        //Get the db from the connection
        var db = dbconn.db("messagingDB");
        //Specify collection to reference
        var collection = db.collection('channels');

        collection.findOne({"name":channelName},function(err,foundChannel){
            if(err) { return console.dir(err); }
            return foundChannel;
        });
        
    });

    return null;
}

DBInterface.getCurrentDateTime = function(){
    var today = new Date();
    var date = today.getFullYear()+'-'+(today.getMonth()+1)+'-'+today.getDate();
    var time = today.getHours() + ":" + today.getMinutes() + ":" + today.getSeconds();
    var dateTime = date+' '+time;
    return dateTime
}

module.exports = DBInterface;