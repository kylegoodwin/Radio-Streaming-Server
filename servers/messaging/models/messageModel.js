const mongoose = require('mongoose');
const AutoIncrement = require('mongoose-sequence')(mongoose);
const Schema = mongoose.Schema;

let messageSchema = new Schema({
    channelID: {type: String, required: true},
    body: {type: String, required: true},
    createdAt: {type: Date, required: true, default: Date.now},
    creator: {
        id: Number,
        userName: String,
        firstName: String,
        lastName: String,
        photoURL: String
    },
    editedAt: {type: Date}
});

messageSchema.plugin(AutoIncrement, {inc_field: 'messageid'});

module.exports = mongoose.model('Message', messageSchema);