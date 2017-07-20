'use strict';

// The "join_shared_replay" message is not actually sent by the client;
// we just store the logic here for organizational purposes
// "data" example:
/*
    {
        table_id: 15103,
    }
*/

// Imports
const globals = require('../globals');
const logger  = require('../logger');
const notify  = require('../notify');

exports.step1 = function(socket, data) {
    // Local variables
    data.userID = socket.userID;
    data.gameID = data.table_id;

    // Validate that this table exists
    let game;
    if (data.gameID in globals.currentGames) {
        game = globals.currentGames[data.gameID];
    } else {
        logger.warn(`messages.join_shared_replay was called for game #${data.gameID}, but it does not exist.`);
        data.reason = `Game #${data.gameID} does not exist.`;
        notify.playerDenied(socket, data);
        return;
    }

    logger.info(`User "${socket.username}" joined shared replay: #${data.gameID}`);

    // Keep track of the user that joined
    game.spectators.push({
        userID: socket.userID,
        username: socket.username,
        present: true,
        socket: socket,
    });
    notify.allTableChange(data);

    // Set their status
    socket.status = 'Replay';
    notify.allUserChange(socket);

    // Set that they are watching a replay
    socket.atTable = {
        id:         data.gameID,
        replay:     true,
        spectating: false,
    };

    // Send them a "game_start" message
    socket.emit('message', {
        type: 'game_start',
        resp: {
            replay: true,
        },
    });
};
