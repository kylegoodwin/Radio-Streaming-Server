const express = require('express');
const router = express.Router();

// Require the controllers
const message_controller = require('../controllers/message_controller.js');
const stream_controller = require('../controllers/stream_controller.js/index.js')

// a simple test url to check that all of our files are communicating correctly.
router.use('/messages/:messageID', message_controller.specificMessage)
router.use('/streams/:streamID/members', stream_controller.streamMembers);
router.use('/streams/:streamID', stream_controller.specificStream);
router.use('/streams', stream_controller.allStreams);


module.exports = router;