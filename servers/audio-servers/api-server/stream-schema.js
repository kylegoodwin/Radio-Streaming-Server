const mongoose = require('mongoose');
const Schema = mongoose.Schema;

let streamSchema = new Schema({
    channelID: {type: String, required: true, unique: true},
    displayName: {type: String, required: true},
    discription: {type: String, required: false},
    genre: {type: String, required: true},
    createdAt: {type: Date, required: true, default: Date.now},
    goLiveTime: {type: Date, required: false},
    creator: {type: Number, required: true},
    active: {type: Boolean, required: true},
    followers: {type: [Number], required: true}
});

module.exports = mongoose.model('Stream', streamSchema);