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
      this.angle = 0;
      this.percentage = 1.0;
      this.left = left;
      this.top = top;
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

    transform: function() {
      let t = 'scale(' + this.percentage + ') rotate('+ this.angle +'deg)';
      $(this.el).css({
        '-webkit-transform': t,
        '-moz-transform': t,
        '-ms-transform': t,
        'transform': t
      });
    },

    scale: function(percentage) {
      this.percentage = percentage;
      this.transform();
    },

    rotate: function(angle) {
      this.angle = angle;
      this.transform();
    },

    nudge: function(deltaLeft, deltaTop) {
      this.left = this.left + deltaLeft;
      this.top = this.top + deltaTop;
      $(this.el).css({left: this.left, top: this.top});
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
