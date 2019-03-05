var mongoose = require('mongoose');
var Schema = mongoose.Schema;

var channelSchema = new Schema({
    name: String,
    description: {type: String, default: ""},
    private: {type: Boolean, default: false},
    members: [{
        id: Number,
        userName: String,
        firstName: String,
        lastName: String,
        photoURL: String
    }],
    createdAt: {type: Date, default: Date.now()},
    creator: {
        id: Number,
        userName: String,
        firstName: String,
        lastName: String,
        photoURL: String
    },
    editedAt: Date
},
{collection: "Channels"}
);

channelSchema.pre("save", function(next){
    var channel = this;
    Channel.find({name: channel.name}, function(err, docs){
        //If there are any results for this name already, dont save it
        if(!docs.length){
            next();
        }else{
            console.log("Channel of name "+channel.name+" already exists.");
            next(new Error("Duplicate Channel Creation"));
        }
    });
});

var Channel = mongoose.model('Channel', channelSchema);

module.exports = Channel;


//mongoose.connect("mongodb://localhost:27017/");


