const mongoose = require('mongoose');
const AutoIncrement = require('mongoose-sequence')(mongoose);
const Schema = mongoose.Schema;

let streamSchema = new Schema({
    streamChannelID: {type: Number, required: true},
    name: {type: String, required: true},
    createdAt: {type: Date, required: true, default: Date.now},
    creator: {type: Number, required: true},
    active: {type: Boolean},
    listeners: {type: [Number]}
});

streamSchema.plugin(AutoIncrement, {inc_field: 'stream_autoid'});

module.exports = mongoose.model('Stream', streamSchema);