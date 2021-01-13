
var server = (function() {
	var opt = {
		// pollingMs : 1000,
	}
	// var id = null;
	// var state = null;
	// var lastVersion = null;

	function init(gameID, options) {
	 //    if (options) {
	 //      for (var i in options) {
	 //        if (opt.hasOwnProperty(i)) {
	 //          opt[i] = options[i];
	 //        }
	 //      }
	 //    }
		// id = gameID;
		// getGameState();
		// // setTimeout(refreshGameState, this.pollingMs);
	}

	function gameState(id, done) {
		$.ajax({
			url: "/api/state/" + id,
			type: "GET",
			dataType: "json",
			contentType: "application/json",
		})
		.done(done)
		.fail(alertFailure);
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

	function joinGame(id, pos, done) {
		var data = {
			ID: id,
			Position: pos,
		}
		$.ajax({
			url: "/api/join",
			type: "POST",
			dataType: "json",
			contentType: "json",
	 	    data: JSON.stringify(data),
		})
		.done(done)
		.fail(alertFailure);
	}

  return {
    init: init,
    gameState: gameState,
    updateUser: updateUser,
    newGame: newGame,
    joinGame: joinGame,
  };

})();

if (typeof module !== 'undefined') {
  module.exports = server;
}
