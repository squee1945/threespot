
var join = (function() {
	var _opt = {
		// pollingMs is how often to call the server for new joining state.
		pollingMs: 1000,
		// repaint is a callback function(joinState) called whenever the join state updates.
		repaint: null,
	}
	var _id = null;
	var _state = null;
	var _lastVersion = null;

	function init(gameID, options) {
	    if (options) {
	      for (let i in options) {
	        if (_opt.hasOwnProperty(i)) {
	          _opt[i] = options[i];
	        }
	      }
	    }
		_id = gameID;
		updateJoinState();
	}

	function updateJoinState() {
		server.joinState(_id, function(joinState) {
			if (joinState.Version != _lastVersion) {
				_lastVersion = joinState.Version;
				_state = joinState;
				if (_opt.repaint != null) {
					_opt.repaint(joinState);
				}
			}
		})
		if (_opt.repaint != null) {
			setTimeout(updateJoinState, _opt.pollingMs);
		}
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
