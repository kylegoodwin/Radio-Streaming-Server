const mongoose = require('mongoose');
const Schema = mongoose.Schema;

let streamSchema = new Schema({
    streamChannelID: {type: Number, required: true},
    name: {type: String, required: true},
    createdAt: {type: Date, required: true, default: Date.now},
    creator: {type: Number, required: true},
    active: {type: Boolean},
    activelisteners: {type: [Number]}
});

module.exports = mongoose.model('Stream', streamSchema);