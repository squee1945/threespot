
var join = (function() {
    var _opt = {
        // pollingMs is how often to call the server for new joining state.
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
        updateJoinState();
    }

    function updateJoinState() {
        server.joinState(_id, function(joinState) {
            if (joinState.Version != _lastVersion) {
                _lastVersion = joinState.Version;
                _state = joinState;
                _lastUpdateEpochMs = Date.now();
                if (_opt.repaint) {
                    _opt.repaint(joinState);
                }
            }
            if (Date.now() - _lastUpdateEpochMs > (_opt.maxPollSeconds * 1000)) {
                alert("Timeout waiting for players. Refresh page to continue.");
                return;
            }
            if (_opt.repaint) {
                setTimeout(updateJoinState, _opt.pollingMs);
            }
        })
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
  module.exports = join;
}
