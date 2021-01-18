
var server = (function() {
	var _opt = {
		// pollingMs : 1000,
	}
	// var id = null;
	// var state = null;
	// var lastVersion = null;

	function init(options) {
	 //    if (options) {
	 //      for (var i in options) {
	 //        if (_opt.hasOwnProperty(i)) {
	 //          _opt[i] = options[i];
	 //        }
	 //      }
	 //    }
		// id = gameID;
		// getGameState();
		// // setTimeout(refreshGameState, this.pollingMs);
	}

	function joinState(id, done) {
		$.ajax({
			url: "/api/join-state/" + id,
			type: "GET",
			dataType: "json",
			contentType: "application/json",
		})
		.done(done)
		.fail(alertFailure);
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
			alert(errorThrown);
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

	function placeBid(id, bid, done) {
		var data = {
			ID: id,
			Bid: bid,
		}
		$.ajax({
			url: "/api/bid",
			type: "POST",
			dataType: "json",
			contentType: "json",
	 	    data: JSON.stringify(data),
		})
		.done(done)
		.fail(alertFailure);
	}

	function playCard(id, card, done) {
		var data = {
			ID: id,
			Card: card,
		}
		$.ajax({
			url: "/api/play",
			type: "POST",
			dataType: "json",
			contentType: "json",
	 	    data: JSON.stringify(data),
		})
		.done(done)
		.fail(alertFailure);
	}

	function callTrump(id, suit, done) {
		var data = {
			ID: id,
			Suit: suit,
		}
		$.ajax({
			url: "/api/trump",
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
		joinState: joinState,
		updateUser: updateUser,
		newGame: newGame,
		joinGame: joinGame,
		placeBid: placeBid,
		playCard: playCard,
		callTrump: callTrump,
	};
})();

if (typeof module !== 'undefined') {
  module.exports = server;
}
