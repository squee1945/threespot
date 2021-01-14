
var kaiser = (function() {
	var _opt = {
		pollingMs: 1000,
	}
	var _id = null;
	var _state = null;
	var _lastVersion = null;

	function init(gameID, options) {
	    if (options) {
	      for (var i in options) {
	        if (_opt.hasOwnProperty(i)) {
	          _opt[i] = options[i];
	        }
	      }
	    }
		_id = gameID;
		updateGameState();
	}

	function updateGameState() {
		server.gameState(_id, (gameState) => {
			if (gameState.Version != _lastVersion) {
				_lastVersion = gameState.Version;
				_state = gameState;
				updateBoard();
			}
		})
		setTimeout(updateGameState, _opt.pollingMs);
	}

	function state() {
		return _state;
	}

	function updateBoard() {
		let s = state().State;
		switch (s) {
			case "JOINING":
				updateBoardJoining();
				break;
			case "BIDDING":
				updateBoardBidding();
				break;
			case "CALLING":
				updateBoardCalling();
				break;
			case "PLAYING":
				updateBoardPlaying();
				break;
			case "COMPLETED":
				updateBoardCompleted();
				break;
			default:
				console.log("Unknown state '" + s + "'");
		}
	}

	function updateBoardJoining() {
		// TODO
	}

	function updateBoardBidding() {
		// TODO
	}

	function updateBoardCalling() {
		// TODO
	}

	function updateBoardPlaying() {
		// TODO
	}

	function updateBoardCompleted() {
		// TODO
	}

  	return {
    	init: init,
    	state: state,
  	};
})();

if (typeof module !== 'undefined') {
  module.exports = kaiser;
}
