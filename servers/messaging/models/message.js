var mongoose = require('mongoose');
var Schema = mongoose.Schema;

var messageSchema = new Schema({
    channelID: String,
    body: {type: String, default: ""},
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
{collection: "Messages"}
);

var Message = mongoose.model('Message', messageSchema);

module.exports = Message;