const mongoose = require('mongoose');
const AutoIncrement = require('mongoose-sequence')(mongoose);
const Schema = mongoose.Schema;

let messageSchema = new Schema({
    streamID: {type: Number, required: true},
    body: {type: String, required: true},
    createdAt: {type: Date, required: true, default: Date.now},
    creator: {type: Number, required: true},
    editedAt: {type: Date}
});

messageSchema.plugin(AutoIncrement, {inc_field: 'messageid'});

module.exports = mongoose.model('Message', messageSchema);