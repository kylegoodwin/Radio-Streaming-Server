const mongoose = require('mongoose');
const AutoIncrement = require('mongoose-sequence')(mongoose);
const Schema = mongoose.Schema;

let channelSchema = new Schema({
    name: {type: String},
    description: {type: String},
    private: {type: Boolean, default: false},
    members: {type: [Number]},
    createdAt: {type: Date, default: Date.now},
    creator: {type: Number},
    editedAt: {type: Date}
});

channelSchema.plugin(AutoIncrement, {inc_field: 'id'});

module.exports = mongoose.model("Channel", channelSchema, "Channel");