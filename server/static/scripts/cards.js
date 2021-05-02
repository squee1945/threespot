var cards = (function() {

  var zIndexCounter = 1;

  var opt = {
    cardSize: {
      width: 69,
      height: 96,
    },
    table: 'body'
  };

  function init(options) {
    if (options) {
      for (var i in options) {
        if (opt.hasOwnProperty(i)) {
          opt[i] = options[i];
        }
      }
    }

    opt.table = $(opt.table)[0];
    if ($(opt.table).css('position') == 'static') {
      $(opt.table).css('position', 'relative');
    }
  }

  function Card(code, left, top) {
    this.init(code, left, top);
  }

  Card.prototype = {
    init: function(code, left, top) {
      this.code = code;
      this.el = $('<div/>')
      .css({
        left: left,
        top: top,
        width: opt.cardSize.width,
        height: opt.cardSize.height,
        position: 'absolute',
        cursor: 'pointer',
        display: 'none'
      }).addClass('card').data('card', this).appendTo($(opt.table));
      this.showCard();
      this.moveToFront();
    },

    showCard: function() {
      $(this.el).css({
        'object-fit': 'cover',
        'background-repeat': 'no-repeat',
        'background-image': 'url(/static/images/cards/' + this.code + '.png)',
        'background-size': '' + opt.cardSize.width + 'px',
      });
    },

    hideCard: function(position) {
      $(this.el).css('background-image', 'url(/static/images/cards/BACK.png)');
    },

    moveToFront: function() {
      $(this.el).css('z-index', zIndexCounter++);
    },

    makeVisible: function() {
      $(this.el).css('display', '');
    },

    makeInvisible: function() {
      $(this.el).css('display', 'none');
    },

    scale: function(percentage) {
      $(this.el).css('transform', 'scale(' + percentage + ')');
    }
  };

  return {
    init: init,
    Card: Card
  };

})();

if (typeof module !== 'undefined') {
  module.exports = cards;
}
