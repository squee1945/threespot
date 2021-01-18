
var kaiser = (function() {
    var _opt = {
        // pollingMs is how often to call the server for new game state.
        pollingMs: 1000,
        // maxPollSeconds is the length of time to poll with no updates before telling the user to refresh.
        maxPollSeconds: 15 * 60,
        // repaint is a callback function(joinState) called whenever the join state updates.
        repaint: null,
    }
    var _id = null;
    var _state = null;
    var _lastVersion = null;
    var _lastUpdateEpochMs = null;

    function init(gameID, options) {
        if (options) {
            for (let i in options) {
                if (_opt.hasOwnProperty(i)) {
                    _opt[i] = options[i];
                }
            }
        }
        _id = gameID;
        _lastUpdateEpochMs = Date.now();
        updateGameState();
    }

    function updateGameState() {
        server.gameState(_id, function(gameState) {
            if (gameState.Version != _lastVersion) {
                _lastVersion = gameState.Version;
                _state = gameState;
                _lastUpdateEpochMs = Date.now();
                if (_opt.repaint) {
                    _opt.repaint(gameState);
                }
            }
            if (Date.now() - _lastUpdateEpochMs > (_opt.maxPollSeconds * 1000)) {
                alert("Timeout waiting for players. Refresh page to continue.");
                return;
            }
            if (_opt.repaint) {
                setTimeout(updateGameState, _opt.pollingMs);
            }
        });
    }

    function state() {
        return _state;
    }

    return {
        init: init,
        state: state,
    };
})();

if (typeof module !== 'undefined') {
  module.exports = kaiser;
}
