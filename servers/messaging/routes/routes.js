const express = require('express');
const router = express.Router();

// Require the controllers
const message_controller = require('../controllers/message_controller.js');
const channel_controller = require('../controllers/channel_controller.js');
const status_controller = require('../controllers/status_controller.js');
//const channel_controller = require('../../audio-servers/api-server/server.js')

// a simple test url to check that all of our files are communicating correctly.
router.use('/comments/:channelID/:messageID', message_controller.specificMessage)
router.use('/comments/:channelID', channel_controller.specificChannel)
router.post('/status/:channelID', status_controller.updateStatus)
// router.use('/channels/:channelID/members', channel_controller.channelMembers);
// router.use('/channels/:channelID', channel_controller.specificChannel);
// router.use('/channels', channel_controller.allChannels);


module.exports = router;