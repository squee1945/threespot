
var kaiser = (function() {
	var opt = {
		pollingMs : 1000,
	}
	var id = null;
	var state = null;

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
			// TODO
			// If the state version has not changed, don't do anything.
			// Otherwise, set the state and fire event
			this.state = json
		})
		.fail(alertFailure);
	}

	function refreshGameState() {
		getGameState()
		setTimeout(refreshGameState, this.pollingMs);
	}

	function state() {
		return this.state;
	}

	function alertFailure(xhr, status, errorThrown) {
		if (xhr.responseJSON != null) {
			alert(xhr.responseJSON['Error']);
		} else {
			alert(Object.keys(xhr) + "\n" + errorThrown + " (status: " + status + ")");
		}
	}

	function callSetUser(name, done) {
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

	function callNewGame(done) {
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
    callSetUser: callSetUser,
    callNewGame: callNewGame,
  };

})();

if (typeof module !== 'undefined') {
  module.exports = kaiser;
}
