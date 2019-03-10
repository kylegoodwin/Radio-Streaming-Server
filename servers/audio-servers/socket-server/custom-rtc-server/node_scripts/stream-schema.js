const mongoose = require('mongoose');
const Schema = mongoose.Schema;

let streamSchema = new Schema({
    channelID: {type: String, required: true},
    displayName: {type: String, required: true},
    createdAt: {type: Date, required: true, default: Date.now},
    creator: {type: Number, required: true}
});

module.exports = mongoose.model('Stream', streamSchema);