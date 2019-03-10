const mongoose = require('mongoose');
const AutoIncrement = require('mongoose-sequence')(mongoose);
const Schema = mongoose.Schema;

//streamID: {type: Number, required: true},

let streamSchema = new Schema({
    name: {type: String, required: true},
    description: {type: String, required: true},
    genre: {type: String, required: true},
    createdAt: {type: Date, required: true, default: Date.now()},
    creator: {
        id: Number,
        userName: String,
        firstName: String,
        lastName: String,
        photoURL: String
    },
    active: {type: Boolean, default: false},
    goLive: {type: Date},
    listeners: {type: [Number]}
});

streamSchema.plugin(AutoIncrement, {inc_field: 'stream_autoid'});

module.exports = mongoose.model('Stream', streamSchema);