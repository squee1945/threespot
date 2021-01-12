
var kaiser = (function() {
	var opt = {
		pollingMs : 1000,
	}
	var id = null;
	var state = null;
	var lastVersion = null;

	function init(gameID, options) {
	    if (options) {
	      for (var i in options) {
	        if (opt.hasOwnProperty(i)) {
	          opt[i] = options[i];
	        }
	      }
	    }
		id = gameID;
		getGameState();
		// setTimeout(refreshGameState, this.pollingMs);
	}

	function getGameState() {
		$.ajax({
			url: "/api/state/" + id,
			type: "GET",
			dataType: "json",
			contentType: "application/json",
		})
		.done((json) => {
			if (lastVersion == null || lastVersion != json["Version"]) {
				lastVersion = json["Version"]
				state = json
				updateBoard();
			}
		})
		.fail(alertFailure);
	}

	function refreshGameState() {
		getGameState()
		setTimeout(refreshGameState, pollingMs);
	}

	function state() {
		return state;
	}

	function updateBoard() {
		let s = state()["State"];
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

	function alertFailure(xhr, status, errorThrown) {
		if (xhr.responseJSON != null) {
			alert(xhr.responseJSON['Error']);
		} else {
			alert(Object.keys(xhr) + "\n" + errorThrown + " (status: " + status + ")");
		}
	}

	function updateUser(name, done) {
		let data = {
			Name: name,
		};
		$.ajax({
		    url: "/api/user",
		    type: "POST",
		    dataType: "json",
	 	    data: JSON.stringify(data),
	    	contentType: "application/json",
		})
		.done(done)
		.fail(alertFailure);
	}

	function newGame(done) {
		$.ajax({
			url: "/api/new",
			type: "POST",
			dataType: "json",
			contentType: "application/json",
		})
		.done(done)
		.fail(alertFailure);
	}

  return {
    init: init,
    state: state,
    updateUser: updateUser,
    newGame: newGame,
  };

})();

if (typeof module !== 'undefined') {
  module.exports = kaiser;
}
