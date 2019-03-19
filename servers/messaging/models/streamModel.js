const mongoose = require('mongoose');
const Schema = mongoose.Schema;

let streamSchema = new Schema({
    channelID: {type: String, required: true, unique: true},
    displayName: {type: String, required: true},
    description: {type: String, required: false},
    status:{ type: {
        text: String,
        image: String
    }, required: false} ,
    genre: {type: String, required: true},
    createdAt: {type: Date, required: true, default: Date.now},
    goLiveTime: {type: Date, required: false},
    creator:{ type: {
        id: Number,
        userName: String,
        firstName: String,
        lastName: String,
        photoURL: String
    }, required: true} ,
    active: {type: Boolean, required: true},
    activeListeners: {type: [Number], required: true},
    followers: {type: [Number], required: true}
});

module.exports = mongoose.model('Stream', streamSchema);